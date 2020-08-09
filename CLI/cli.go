package main

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/url"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

/*
流程(CLI)：
1.读取settings.json，无则赋默认值，创建文件
2.通过检测"%HOMEDIR%/AppData/Local/AkiVer/"是否存在判断是否安装了CSGO DEMOS MANAGER，否则退出
3.通过检测"%HOMEDIR%/AppData/Local/AkiVer/hlae/hlae.exe"是否存在判断是否安装了HLAE，否则跳过XML解析
4.解析包含本地版本信息的XML文件"HLAE/changelog.xml"，获得当前版本
5.利用API获取包含HLAE仓库信息的JSON文件并解析，获得版本号和下载地址
6.判断是否要下载/更新，是则利用CDN加速尝试下载HLAE-Release仓库的文件
7.下载成功则进行下一步，否则直接从advancedfx原仓库下载
8.解压到临时目录"./temp/"检查"changelog.xml和"hlae.exe"的正确性，然后移动文件，覆盖原目录
9.生成/更新"Version"文件，格式"2.102.0"
*/

///// util
func pause() {
	err := os.RemoveAll("./temp")
	if err != nil {
		log.Println(err)
		os.Exit(14)
	}
	var b byte
	fmt.Println("\n请按Enter结束...")
	fmt.Scanf("%v", b)
}

//打开文件和读内容 利用io/ioutil
func readAll(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	//对内容的操作
	//ReadFile返回的是[]byte字节切片，要用string()方法转变成字符串
	//去除内容结尾的换行符
	str := strings.TrimRight(string(content), "\n")
	return str, nil
}

//文件写入 先清空再写入 利用ioutil
func writeFast(filePath string, content string) error {
	dir, _ := path.Split(filePath)
	exist, err := isFileExisted(dir)
	if err != nil {
		return err
	} else if exist == false {
		os.Mkdir(dir, os.ModePerm)
	}
	err = ioutil.WriteFile(filePath, []byte(content), 0666)
	if err != nil {
		return err
	} else {
		return nil
	}
}

//判断文件/文件夹是否存在
func isFileExisted(path string) (bool, error) {
	//返回 true, nil = 存在
	//返回 false, nil = 不存在
	//返回 _, !nil = 位置错误，无法判断
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//利用HTTP Get请求获得数据
func getHttpData(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	_ = resp.Body.Close()

	return string(data), nil
}

//下载文件 (下载地址，存放位置)
func downloadFile(url string, location string) error {
	//利用HTTP下载文件并读取内容给data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		errorInfo := "http failed, check if file exists, HTTP Status Code:" + strconv.Itoa(resp.StatusCode)
		return errors.New(errorInfo)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()

	//确保下载位置存在
	_, fileName := path.Split(url)
	ok, err := isFileExisted(location)
	if err != nil {
		return err
	} else if ok == false {
		err := os.Mkdir(location, os.ModePerm)
		if err != nil {
			return err
		}
	}
	//文件写入 先清空再写入 利用ioutil
	err = ioutil.WriteFile(location+"/"+fileName, data, 0666)
	if err != nil {
		return err
	} else {
		return nil
	}
}

//压缩
func Zip(from string, toZip string) error {
	zipfile, err := os.Create(toZip)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	_ = filepath.Walk(from, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(path, filepath.Dir(from)+"/")
		// header.Name = path
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
		}
		return err
	})

	return err
}

//解压
func Unzip(zipFile string, to string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		fpath := filepath.Join(to, f.Name)
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}

			inFile, err := f.Open()
			defer inFile.Close()
			if err != nil {
				return err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			defer outFile.Close()
			if err != nil {
				return err
			}

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//规格化路径
func FormatPath(s string) string {
	switch runtime.GOOS {
	case "windows":
		return strings.Replace(s, "/", "\\", -1)
	case "darwin", "linux":
		return strings.Replace(s, "\\", "/", -1)
	default:
		log.Println("only support linux,windows,darwin, but os is " + runtime.GOOS)
		return s
	}
}

//复制文件夹
func copyDir(from string, to string) error {
	from = FormatPath(from)
	to = FormatPath(to)

	//确保目标路径存在，否则复制报错exit status 4
	exist, err := isFileExisted(to)
	if err != nil {
		return err
	} else if exist == false {
		err := os.Mkdir(to, os.ModePerm)
		if err != nil {
			return err
		}
	}
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("xcopy", from, to, "/I", "/E", "/Y")
	case "darwin", "linux":
		cmd = exec.Command("cp", "-R", from, to)
	}

	_, err = cmd.Output()
	if err != nil {
		return err
	}
	//fmt.Println(string(output))
	return nil
}

///// struct
type Setting struct {
	Version       string
	LatestVersion string
	LocalVersion  string
	FFmpegVersion string
	Url           string
	FileName      string
	HlaeAPI       string
	API           string

	FFmpegAPI string
	HlaeExist bool
	//launchOption	string
	//CsgoPath		string
}
type Asset struct {
	URL                string `json:"url"`
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	ContentType        string `json:"content_type"`
	State              string `json:"state"`
	Size               int    `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Latest struct {
	URL     string  `json:"url"`
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Message string  `json:"message"`
	Assets  []Asset `json:"assets"`
}

type Changelog struct {
	XMLName xml.Name `xml:"changelog"`
	Text    string   `xml:",chardata"`
	Release []struct {
		Text    string `xml:",chardata"`
		Name    string `xml:"name"`
		Version string `xml:"version"`
		Time    string `xml:"time"`
		Changes struct {
			Text   string `xml:",chardata"`
			Change []struct {
				Text string   `xml:",chardata"`
				Type string   `xml:"type,attr"`
				Br   []string `xml:"br"`
			} `xml:"change"`
			Changed struct {
				Text string `xml:",chardata"`
				Type string `xml:"type,attr"`
			} `xml:"changed"`
		} `xml:"changes"`
		Comments struct {
			Text string   `xml:",chardata"`
			Br   []string `xml:"br"`
		} `xml:"comments"`
	} `xml:"release"`
	H1 string `xml:"h1"`
}

type FFmpegTag struct {
	Message    string `json:"message"`
	Name       string `json:"name"`
	ZipballURL string `json:"zipball_url"`
	TarballURL string `json:"tarball_url"`
	Commit     struct {
		Sha string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
	NodeID string `json:"node_id"`
}

///// important functions
//解析Json，获取最新版本号和下载地址 return TagName, Asset Slice, nil
func parseReleaseInfo(owner string, repo string) (string, []Asset, error) {
	//GET请求获得JSON
	jsonData, err := getHttpData("https://api.github.com/repos/" + owner + "/" + repo + "/releases/latest")
	if err != nil {
		log.Println(err)
		return "", nil, err
	}

	//初始化实例并解析JSON
	var latestInst Latest
	err = json.Unmarshal([]byte(jsonData), &latestInst) //第二个参数要地址传递
	if err != nil {
		return "", nil, err
	}

	//链接有问题也会返回Json，且 "Message": "Not Found"
	if latestInst.Message == "Not Found" {
		return "", nil, errors.New("got Json but no valid. Check URL")
	}

	return latestInst.TagName, latestInst.Assets, nil
}

func readSettings(path string) (Setting, error) {
	//检查文件是否存在
	exist, err := isFileExisted(path)
	if err != nil {
		return Setting{}, err
	} else if exist == true {
		//存在则读取文件
		content, err := readAll(path)
		if err != nil {
			return Setting{}, err
		}

		//初始化实例并解析JSON
		var settingInst Setting
		err = json.Unmarshal([]byte(content), &settingInst) //第二个参数要地址传递
		if err != nil {
			return Setting{}, err
		}
		settingInst.Files = nil //清空API，防止累加

		return settingInst, nil
	} else {

		return Setting{}, nil
	}
}

func saveSettings(path string) error {
	//检查文件是否存在
	exist, err := isFileExisted(path)
	if err != nil {
		return err
	} else if exist == true {
		//存在则删除文件
		ok, err := isFileExisted(path)
		if err != nil {
			return err
		} else if ok == true {
			err := os.Remove(path)
			if err != nil {
				return err
			}
		}
	}

	JsonData, err := json.Marshal(Updater) //第二个参数要地址传递
	if err != nil {
		return err
	}

	err = writeFast(path, string(JsonData))
	if err != nil {
		return err
	}

	return nil
}

func parseChangelog(xmlData string) (string, error) {
	//初始化实例并解析
	var ChangelogInst Changelog
	//使用encoding/xml库
	err := xml.Unmarshal([]byte(xmlData), &ChangelogInst) //第二个参数要地址传递
	if err != nil {
		log.Println(err)
		return "", err
	}
	//返回Changelog里的版本号
	return "v" + ChangelogInst.Release[0].Version, nil
}

func generateVersion(version string, path string) {
	ver := strings.Replace(version, "v", "", -1)
	fmt.Println("正在生成版本文件：", ver)
	err := writeFast(path, ver)
	if err != nil {
		fmt.Println("版本文件生成失败")
		log.Println(err)
		pause()
		os.Exit(14)
	} else {
		fmt.Println("版本文件生成成功")
	}
}

func getFFmpegLatestVersion(jsonData string) (string, error) {
	//初始化实例
	var FFmpegInst []FFmpegTag

	//注释下面一行->使用encoding/json库
	var json = jsoniter.ConfigCompatibleWithStandardLibrary //使用高性能json-iterator/go库
	err := json.Unmarshal([]byte(jsonData), &FFmpegInst)    //第二个参数要地址传递
	if err != nil {
		return "", err
	}

	//链接有问题也会返回Json，且 "Message": "Not Found"
	if FFmpegInst[0].Message == "Not Found" {
		return "", errors.New("got Json but no valid. Check URL")
	}
	//获得最新版本号
	latestTag := ""
	for _, tag := range FFmpegInst {
		//过滤旧版本和开发版-dev
		if !strings.Contains(tag.Name, "v") && !strings.Contains(tag.Name, "dev") && strings.Compare(tag.Name, latestTag) > 0 {
			latestTag = tag.Name
		}
	}
	//去除版本号开头的n
	latestTag = strings.Replace(latestTag, "n", "", -1)
	return latestTag, nil
}

///// 全局变量
var Updater = &Setting{
	Version:       "0.2.0",
	LatestVersion: "",
	LocalVersion:  "",
	FFmpegVersion: "",
	Url:           "",
	FileName:      "",
	HlaeAPI:       "https://api.github.com/repos/advancedfx/advancedfx/releases/latest",
	//CdnAPI:        "https://cdn.jsdelivr.net/gh/Purple-CSGO/HLAE-Release@",
	FFmpegAPI: "https://api.github.com/repos/FFmpeg/FFmpeg/tags",
	HlaeExist: false,
}

func main() {

	//1.读取settings.json，不存在或出错则赋默认值
	temp, err := readSettings("./settings.json")
	if err != nil {
		log.Fatal(err)
	} else if temp.Version != "" {
		Updater = &temp
	}

	//2.Welcome~
	fmt.Println("-----------------------------------------------------------------")
	fmt.Println("HLAE Updater CLI -", Updater.Version)
	fmt.Println("-----------------------------------------------------------------")

	//3.通过检测"%HOMEDIR%/AppData/Local/AkiVer/"是否存在判断是否安装了CSGO DEMOS MANAGER，否则退出
	usr, err := user.Current()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	dir := usr.HomeDir + "/AppData/Local/AkiVer"
	ok, err := isFileExisted(dir)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	} else if ok == false {
		fmt.Println("没有检测到CSGO Demos Manager，请确认安装后再使用本工具")
		pause()
		os.Exit(3)
	}

	//3.通过检测"%HOMEDIR%/AppData/Local/AkiVer/hlae/hlae.exe"是否存在判断是否安装了HLAE，否则跳过XML解析
	hlaePath := usr.HomeDir + "/AppData/Local/AkiVer/hlae/hlae.exe"
	Updater.HlaeExist, err = isFileExisted(hlaePath)
	if err != nil {
		log.Println(err)
		os.Exit(4)
	}

	//4.解析包含本地版本信息的XML文件"HLAE/changelog.xml"，获得当前版本
	if Updater.HlaeExist == false {
		fmt.Println("检测到尚未给CSGO Demos Manager安装HLAE")
	} else {
		changelogPath := usr.HomeDir + "/AppData/Local/AkiVer/hlae/changelog.xml"

		xmlData, err := readAll(changelogPath)
		if err != nil {
			log.Println(err)
			os.Exit(5)
		}

		Updater.LocalVersion, err = parseChangelog(xmlData)
		if err != nil {
			fmt.Println("获取本地版本号失败")
			pause()
			log.Println(err)
		} else {
			Updater.HlaeExist = true
			fmt.Println("本地HLAE版本：", Updater.LocalVersion)
		}
	}

	//5.利用API获取包含HLAE仓库信息的JSON文件并解析，获得版本号和下载地址
	fmt.Println("正在获取最新版本信息...")
	jsonData, err := getHttpData(Updater.HlaeAPI)
	if err != nil {
		log.Println(err)
		os.Exit(6)
	}

	Updater.LatestVersion, Updater.Url, Updater.FileName, err = parseLatestInfo(jsonData)
	if err != nil {
		log.Println(err)
		os.Exit(7)
	} else {
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println("最新HLAE版本：", Updater.LatestVersion)
		fmt.Println("下载地址：", Updater.Url)
		fmt.Println("-----------------------------------------------------------------")
	}

	//6.判断是否要下载/更新，是则利用CDN加速尝试下载HLAE-Release仓库的文件
	if Updater.HlaeExist == true {
		res := strings.Compare(Updater.LatestVersion, Updater.LocalVersion)
		if res == 0 {
			fmt.Println("已是最新版本")
			generateVersion(Updater.LatestVersion, usr.HomeDir+"/AppData/Local/AkiVer/hlae/version")
			pause()
			os.Exit(0)
		} else if res < 0 {
			fmt.Println("发生异常，本地版本号>最新版本号，请检查本地HLAE文件")
			pause()
			os.Exit(8)
		}
	}

	///// 利用grab下载
	//generatePath := "./temp/"
	//resp, err := grab.Get(generatePath, cdnURL)
	//if err != nil {
	//	fmt.Println("加速下载失败×")
	//	log.Println(err)
	//} else {
	//	fmt.Println(resp.Filename, " 已下载")
	//	//解压文件
	//	err = Unzip(generatePath+"hlae_2_102_0.zip", "./temp/"+Updater.LatestVersion)
	//	if err != nil {
	//		log.Println(err)
	//	}
	//}

	fmt.Println("正在尝试加速下载...")
	cdnURL := Updater.CdnAPI + Updater.LatestVersion + "/" + Updater.FileName
	fmt.Println("cdnURL:", cdnURL)
	err = downloadFile(cdnURL, "./temp")
	if err != nil {
		fmt.Println("加速下载失败")
		log.Println(err)

		//7.下载成功则进行下一步，否则直接从advancedfx原仓库下载
		fmt.Println("正在从GitHub原地址下载...")
		err = downloadFile(Updater.Url, "./temp/")
		if err != nil {
			fmt.Println("原地址下载失败，请检查网络连接")
			log.Println(err)
			pause()
			os.Exit(9)
		}
	}

	//8.解压到临时目录"./temp/"检查"changelog.xml和"hlae.exe"的正确性，然后移动文件，覆盖原目录
	fmt.Println("下载成功，正在解压...")
	tempDir := "./temp/hlae/"
	os.RemoveAll(tempDir)
	err = Unzip("./temp/"+Updater.FileName, tempDir)
	if err != nil {
		fmt.Println("解压失败")
		log.Println(err)
		pause()
		os.Exit(10)
	} else {
		ok, err := isFileExisted(tempDir + "hlae.exe")
		if err != nil {
			log.Println(err)
			os.Exit(11)
		} else if ok == false {
			log.Println(errors.New("successfully unzipped but no file is found"))
			pause()
			os.Exit(12)
		}
	}

	//移动，覆盖原目录
	fmt.Println("解压成功，正在移动文件...")
	err = copyDir(tempDir, usr.HomeDir+"/AppData/Local/AkiVer/hlae")
	if err != nil {
		log.Println(err)
		pause()
		os.Exit(13)
	}

	//9.生成/更新"Version"文件，格式"2.102.0"
	generateVersion(Updater.LatestVersion, usr.HomeDir+"/AppData/Local/AkiVer/hlae/version")

	//10.利用API获取FFMPEG最新版本号
	fmt.Println("正在获取FFMPEG最新版本信息...")
	jsonData, err = getHttpData(Updater.FFmpegAPI)
	if err != nil {
		log.Println(err)
		os.Exit(6)
	}

	Updater.FFmpegVersion, err = getFFmpegLatestVersion(jsonData)
	if err != nil {
		log.Println(err)
	} else {
		//Linux 64位地址 https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz
		//Windows 64位地址 需要版本号 shared/static https://ffmpeg.zeranoe.com/builds/win64/static/ffmpeg-4.3.1-win64-static.zip
		//MacOS 64位地址 需要版本号 shared/static https://ffmpeg.zeranoe.com/builds/macos64/static/ffmpeg-4.3.1-macos64-static.zip
		fmt.Println("最新版本:", Updater.FFmpegVersion)
	}

	fmt.Println("-----------------------------------------------------------------")
	if Updater.HlaeExist == true {
		fmt.Println("HLAE更新完成，当前版本号：", Updater.LatestVersion)
	} else {
		fmt.Println("HLAE安装完成，当前版本号：", Updater.LatestVersion,
			"\n请在CSGO Demos Manager的设置中点击'启用HLAE'")
	}
	fmt.Println("-----------------------------------------------------------------")
	pause()
}
