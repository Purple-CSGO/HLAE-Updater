package main

import (
	//"archive/zip"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	//"io"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/url"
	"os"
	"os/exec"
	"os/user"
	"path"
	//"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/gen2brain/go-unarr"
	jsoniter "github.com/json-iterator/go"
)

///// util
func pause() {
	err := os.RemoveAll("./temp")
	if err != nil {
		log.Fatal(err)
	}
	var b byte
	fmt.Println("\n请按Enter结束...")
	_, _ = fmt.Scanf("%v", b)
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
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return err
		}
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

//利用HTTP Get请求获得数据json
func getHttpData(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	//body, err := resp.Js
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
//func Zip(from string, toZip string) error {
//	zipfile, err := os.Create(toZip)
//	if err != nil {
//		return err
//	}
//	defer zipfile.Close()
//
//	archive := zip.NewWriter(zipfile)
//	defer archive.Close()
//
//	_ = filepath.Walk(from, func(path string, info os.FileInfo, err error) error {
//		if err != nil {
//			return err
//		}
//
//		header, err := zip.FileInfoHeader(info)
//		if err != nil {
//			return err
//		}
//
//		header.Name = strings.TrimPrefix(path, filepath.Dir(from)+"/")
//		// header.Name = path
//		if info.IsDir() {
//			header.Name += "/"
//		} else {
//			header.Method = zip.Deflate
//		}
//
//		writer, err := archive.CreateHeader(header)
//		if err != nil {
//			return err
//		}
//
//		if !info.IsDir() {
//			file, err := os.Open(path)
//			if err != nil {
//				return err
//			}
//			defer file.Close()
//			_, err = io.Copy(writer, file)
//		}
//		return err
//	})
//
//	return err
//}
//
////解压
//func Unzip(zipFile string, to string) error {
//	zipReader, err := zip.OpenReader(zipFile)
//	if err != nil {
//		return err
//	}
//	defer zipReader.Close()
//
//	for _, f := range zipReader.File {
//		fpath := filepath.Join(to, f.Name)
//		if f.FileInfo().IsDir() {
//			_ = os.MkdirAll(fpath, os.ModePerm)
//		} else {
//			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
//				return err
//			}
//
//			inFile, err := f.Open()
//			defer inFile.Close()
//			if err != nil {
//				return err
//			}
//
//			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
//			defer outFile.Close()
//			if err != nil {
//				return err
//			}
//
//			_, err = io.Copy(outFile, inFile)
//			if err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}

//解压zip 7z rar tar
func decompress(from string, to string) error {
	a, err := unarr.NewArchive(from)
	if err != nil {
		return err
	}
	defer a.Close()

	_, err = a.Extract(to)
	if err != nil {
		return err
	}

	return nil
}

//func Un7z(filePath string, to string) error {
//	//判断路径是否存在
//	ok, err := isFileExisted(filePath)
//	if err != nil {
//		return err
//	} else if ok == false {
//		return errors.New("7z file does not exist")
//	}
//
//	sz, err := go7z.OpenReader(filePath)
//	if err != nil {
//		panic(err)
//	}
//	defer sz.Close()
//
//	for {
//		hdr, err := sz.Next()
//		if err == io.EOF {
//			break // End of archive
//		}
//		if err != nil {
//			return err
//		}
//
//		// If empty stream (no contents) and isn't specifically an empty file...
//		// then it's a directory.
//		if hdr.IsEmptyStream && !hdr.IsEmptyFile {
//			if err := os.MkdirAll(to + "/" + hdr.Name, os.ModePerm); err != nil {
//				fmt.Println("line281")
//				return err
//			}
//			continue
//		}
//
//		// Create file
//		f, err := os.Create(to + "/" + hdr.Name)
//		if err != nil {
//			fmt.Println("line290")
//			return err
//		}
//		defer f.Close()
//
//		if _, err := io.Copy(f, sz); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}

//规格化路径
func FormatPath(s string) string {
	switch runtime.GOOS {
	case "windows":
		s = strings.Replace(s, "/", "\\", -1)
		return strings.TrimRight(s, "\\")
	case "darwin", "linux":
		s = strings.Replace(s, "\\", "/", -1)
		return strings.TrimRight(s, "\\")
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
	Version       string   //Version of this tool
	LatestVersion string   //Latest version of hlae
	LocalVersion  string   //Local version of hlae
	FFmpegVersion string   //Local FFmpeg version
	HlaeAPI       string   //API for getting hlae's latest info
	HlaeCdnAPI    []string //API for speed up downloading hlae
	FFmpegAPI     string   //API for getting FFmpeg's latest info
	FFmpegCdnAPI  []string //API for speed up downloading ffmpeg
	//Temporary for functions
	Url         string //download link
	FileName    string //Name of file to be dealed with
	HlaeExist   bool   //If hlae exists in this computer
	FFmpegExist bool   //If FFmpeg exists in this computer
}

///// 全局变量 TODO 修改备份hlae api获取方式 保留一个手动HLAE-Backup仓库
var Updater = &Setting{
	Version:       "0.3.4",
	LatestVersion: "",
	LocalVersion:  "",
	FFmpegVersion: "",
	HlaeAPI:       "https://api.github.com/repos/advancedfx/advancedfx/releases/latest",
	HlaeCdnAPI: []string{
		"https://cdn.jsdelivr.net/gh/Purple-CSGO/HLAE-Archieve",
		"https://cdn.jsdelivr.net/gh/yellowfisherz/HLAE-Release",
		"https://cdn.jsdelivr.net/gh/Tucd7v/Hlaefarmer",
		"https://cdn.jsdelivr.net/gh/Purple-CSGO/HLAE-Manual-Archieve",
		"https://cdn.jsdelivr.net/gh/Purple-CSGO/afx-backup",
	},
	FFmpegAPI: "https://api.github.com/repos/FFmpeg/FFmpeg/tags",
	FFmpegCdnAPI: []string{
		"https://cdn.jsdelivr.net/gh/Purple-CSGO/FFmpeg-Archieve",
	},
	//Temporary for functions
	Url:         "",
	FileName:    "",
	HlaeExist:   false,
	FFmpegExist: false,
}

//Github Asset
type Asset struct {
	URL                string `json:"url"`
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	ContentType        string `json:"content_type"`
	State              string `json:"state"`
	Size               int    `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

//Github latest info
type Latest struct {
	URL     string  `json:"url"`
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Message string  `json:"message"`
	Assets  []Asset `json:"assets"`
}

//HLAE changelog.xml
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

//Github FFmpeg info
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
//func parseReleaseInfo(owner string, repo string) (string, []Asset, error) {
//	//GET请求获得JSON
//	jsonData, err := getHttpData("https://api.github.com/repos/" + owner + "/" + repo + "/releases/latest")
//	if err != nil {
//		log.Println(err)
//		return "", nil, err
//	}
//
//	//初始化实例并解析JSON
//	var latestInst Latest
//	err = json.Unmarshal([]byte(jsonData), &latestInst) //第二个参数要地址传递
//	if err != nil {
//		return "", nil, err
//	}
//
//	//链接有问题也会返回Json，且 "Message": "Not Found"
//	if latestInst.Message == "Not Found" {
//		return "", nil, errors.New("got Json but no valid. Check URL")
//	}
//
//	return latestInst.TagName, latestInst.Assets, nil
//}

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

//解析Json，获取最新版本号和下载地址
func parseLatestInfo(jsonData string) (string, string, string, error) {
	//初始化实例
	var latestInst Latest
	//
	//jsonData = strings.Trim(jsonData, "\"")
	//jsonData = strings.Trim(jsonData, "[")
	//jsonData = strings.Trim(jsonData, "]")
	//注释下面一行->使用encoding/json库
	var jsonx = jsoniter.ConfigCompatibleWithStandardLibrary //使用高性能json-iterator/go库
	//fmt.Println(jsonData)
	err := jsonx.Unmarshal([]byte(jsonData), &latestInst) //第二个参数要地址传递
	if err != nil {
		return "", "", "", err
	}

	//链接有问题也会返回Json，且 "Message": "Not Found"
	if latestInst.Message == "Not Found" {
		return "", "", "", errors.New("got Json but no valid. Check URL")
	}
	if strings.Contains(latestInst.Message, "API rate limit") {
		return "", "", "", errors.New("reach the rate limit of API. Wait for some time")
	}
	//获得zip附件信息
	var url, fileName string
	for _, file := range latestInst.Assets {
		//过滤掉源码文件
		if file.State == "uploaded" && !strings.Contains(file.Name, ".asc") && strings.Contains(file.Name, ".zip") {
			url = file.BrowserDownloadURL
			fileName = file.Name
		}
	}

	return latestInst.TagName, url, fileName, nil
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
	err := writeFast(path, ver)
	if err != nil {
		fmt.Println("·版本文件生成失败")
		log.Println(err)
		pause()
		os.Exit(1)
	}
}

func getFFmpegLatestVersion(jsonData string) (string, error) {
	//初始化实例
	var FFmpegInst []FFmpegTag

	//注释下面一行->使用encoding/json库
	var jsonx = jsoniter.ConfigCompatibleWithStandardLibrary //使用高性能json-iterator/go库
	err := jsonx.Unmarshal([]byte(jsonData), &FFmpegInst)    //第二个参数要地址传递
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

//TODO os.exit() 错误代码
func main() {

	//1.读取settings.json，不存在或出错则赋默认值
	temp, err := readSettings("./settings.json")
	if err != nil {
		log.Println(err)
		pause()
		os.Exit(11)
	} else if temp.Version != "" {
		Updater = &temp
	}

	//11.保存设置
	defer func() {
		err = saveSettings("./settings.json")
		if err != nil {
			log.Println(err)
			pause()
			os.Exit(88)
		}
	}()
	//TODO 多语言支持
	//2.Welcome~

	fmt.Println("┏┓ ┏┓┏┓   ┏━━━┓┏━━━┓    ┏┓ ┏┓┏━━━┓┏━━━┓┏━━━┓┏━━━━┓┏━━━┓┏━━━┓    ┏━━━┓    ┏━━━┓    ┏┓ ┏┓")
	fmt.Println("┃┃ ┃┃┃┃   ┃┏━┓┃┃┏━━┛    ┃┃ ┃┃┃┏━┓┃┗┓┏┓┃┃┏━┓┃┃┏┓┏┓┃┃┏━━┛┃┏━┓┃    ┃┏━┓┃    ┃┏━┓┃    ┃┃ ┃┃")
	fmt.Println("┃┗━┛┃┃┃   ┃┃ ┃┃┃┗━━┓    ┃┃ ┃┃┃┗━┛┃ ┃┃┃┃┃┃ ┃┃┗┛┃┃┗┛┃┗━━┓┃┗━┛┃    ┃┃ ┃┃    ┗┛┏┛┃    ┃┗━┛┃")
	fmt.Println("┃┏━┓┃┃┃ ┏┓┃┗━┛┃┃┏━━┛    ┃┃ ┃┃┃┏━━┛ ┃┃┃┃┃┗━┛┃  ┃┃  ┃┏━━┛┃┏┓┏┛    ┃┃ ┃┃    ┏┓┗┓┃    ┗━━┓┃")
	fmt.Println("┃┃ ┃┃┃┗━┛┃┃┏━┓┃┃┗━━┓    ┃┗━┛┃┃┃   ┏┛┗┛┃┃┏━┓┃  ┃┃  ┃┗━━┓┃┃┃┗┓    ┃┗━┛┃ ┏┓ ┃┗━┛┃ ┏┓    ┃┃")
	fmt.Println("┗┛ ┗┛┗━━━┛┗┛ ┗┛┗━━━┛    ┗━━━┛┗┛   ┗━━━┛┗┛ ┗┛  ┗┛  ┗━━━┛┗┛┗━┛    ┗━━━┛ ┗┛ ┗━━━┛ ┗┛    ┗┛")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("·HLAE+FFmpeg自动安装/更新工具 by Purp1e")
	fmt.Println("·项目地址：\thttps://github.com/Purple-CSGO/HLAE-Updater")
	fmt.Println("·中文站地址：\thttps://hlae.site/topic/453")
	fmt.Println("·反馈邮箱：\t438518244@qq.com")
	fmt.Println("─────────────────────────────────────  说明 ────────────────────────────────────────────")
	fmt.Println("1. 本工具暂时只为CSGO Demos Manager安装HLAE/FFmpeg服务")
	fmt.Println("2. 系统用户名最好不包含空格/中文/俄文等字符")
	fmt.Println("3. CDN加速和备用API相比比官方最新版滞后2分钟")
	fmt.Println("4. FFmpeg加速为手动维护，不一定是最新版")
	fmt.Println("────────────────────────────────────────────────────────────────────────────────────────")

	//┏━━━┓     ┏┓  ┏━━━┓ ┏━━━┓ ┏┓ ┏┓  ┏━━━┓ ┏━━━┓ ┏━━━┓ ┏━━━┓ ┏━━━┓"
	//┃┏━┓┃    ┏┛┃  ┃┏━┓┃ ┃┏━┓┃ ┃┃ ┃┃  ┃┏━━┛ ┃┏━━┛ ┃┏━┓┃ ┃┏━┓┃ ┃┏━┓┃"
	//┃┃ ┃┃    ┗┓┃  ┗┛┏┛┃ ┗┛┏┛┃ ┃┗━┛┃  ┃┗━━┓ ┃┗━━┓ ┗┛ ┃┃ ┃┗━┛┃ ┃┗━┛┃"
	//┃┃ ┃┃     ┃┃  ┏━┛┏┛ ┏┓┗┓┃ ┗━━┓┃  ┗━━┓┃ ┃┏━┓┃    ┃┃ ┃┏━┓┃ ┗━━┓┃"
	//┃┗━┛┃ ┏┓ ┏┛┗┓ ┃ ┗━┓ ┃┗━┛┃    ┃┃  ┏━━┛┃ ┃┗━┛┃    ┃┃ ┃┗━┛┃ ┏━━┛┃"
	//┗━━━┛ ┗┛ ┗━━┛ ┗━━━┛ ┗━━━┛    ┗┛  ┗━━━┛ ┗━━━┛    ┗┛ ┗━━━┛ ┗━━━┛"

	//3.通过检测"%HOMEDIR%/AppData/Local/AkiVer/"是否存在判断是否安装了CSGO DEMOS MANAGER，否则退出
	usr, err := user.Current()
	if err != nil {
		log.Println(err)
		pause()
		os.Exit(23)
	}
	ok, err := isFileExisted(usr.HomeDir + "/AppData/Local/AkiVer")
	if err != nil {
		log.Println(err)
		pause()
		os.Exit(33)
	} else if ok == false {
		fmt.Println("·没有检测到CSGO Demos Manager，请确认安装后再使用本工具")
		fmt.Printf("·官方最新版：https://github.com/akiver/CSGO-Demos-Manager/releases/latest")
		fmt.Printf("·中文站搬运贴：https://hlae.site/topic/390")
		fmt.Printf("·搬运链接：https://cloud.189.cn/t/BVZbQvUJFrum（访问码：jt7e）")
		pause()
		os.Exit(0)
	}

	fmt.Println("·正在获取本地HLAE版本信息...")
	//4.通过检测"%HOMEDIR%/AppData/Local/AkiVer/hlae/hlae.exe"是否存在判断是否安装了HLAE，否则跳过XML解析
	Updater.HlaeExist, err = isFileExisted(usr.HomeDir + "/AppData/Local/AkiVer/hlae/hlae.exe")
	if err != nil {
		log.Println(err)
		pause()
		os.Exit(55)
	}

	//5.解析包含本地版本信息的XML文件"HLAE/changelog.xml"，获得当前版本
	if Updater.HlaeExist == false {
		fmt.Println("·检测到尚未给CSGO Demos Manager安装HLAE")
	} else {
		xmlData, err := readAll(usr.HomeDir + "/AppData/Local/AkiVer/hlae/changelog.xml")
		if err != nil {
			log.Println(err)
			pause()
			os.Exit(35)
		}

		Updater.LocalVersion, err = parseChangelog(xmlData)
		if err != nil {
			log.Println(err)
			pause()
			os.Exit(36)
		} else {
			Updater.HlaeExist = true
			fmt.Println("·本地HLAE版本：", Updater.LocalVersion)
			fmt.Println("────────────────────────────────────────────────────────────────────────────────────────")
		}
	}

	//6.利用API获取包含HLAE仓库信息的JSON文件并解析，获得版本号和下载地址
	fmt.Println("·正在获取HLAE最新版本信息...")
	jsonData, err := getHttpData(Updater.HlaeAPI)
	if err != nil {
		log.Println(err)
		fmt.Println("·HLAE API访问失败，正在使用备用API...")
		for i, API := range Updater.HlaeCdnAPI {
			jsonData, err = getHttpData(API + "/release.json")
			if err != nil {
				fmt.Println("·第" + strconv.Itoa(i) + "个备用API访问失败")
				log.Println(err)
			} else {
				break
			}
		}
	}
	var tagName string
	tagName, Updater.Url, Updater.FileName, err = parseLatestInfo(jsonData)
	if err != nil {
		log.Println(err)
		fmt.Println("·HLAE API访问失败，正在使用备用API...")
		for i, API := range Updater.HlaeCdnAPI {
			jsonData, err = getHttpData(API + "/release.json")
			if err != nil {
				fmt.Println("·第" + strconv.Itoa(i) + "个备用API访问失败")
				log.Println(err)
			} else {
				break
			}
		}
		if err != nil {
			os.Exit(7)
		}
	} else {
		Updater.LatestVersion = tagName
		fmt.Println("·最新HLAE版本：", Updater.LatestVersion)
		fmt.Println("────────────────────────────────────────────────────────────────────────────────────────")
		fmt.Println("·下载地址：\n  " + Updater.Url)
		fmt.Println("────────────────────────────────────────────────────────────────────────────────────────")
	}

	//7.判断是否要下载/更新，是则利用CDN加速尝试
	res := strings.Compare(Updater.LatestVersion, Updater.LocalVersion)
	if Updater.HlaeExist == true && res < 0 {
		fmt.Println("·发生异常，本地版本号>最新版本号，请检查本地HLAE文件")
		pause()
		os.Exit(8)
	} else if Updater.HlaeExist == true && res == 0 {
		fmt.Println("·HLAE已是最新版本")
		fmt.Println("────────────────────────────────────────────────────────────────────────────────────────")
	} else if Updater.HlaeExist == false || res > 0 {
		//hlae不存在或者版本低于最新版时更新
		fmt.Println("·正在尝试加速下载...")
		for i, API := range Updater.HlaeCdnAPI {
			cdnURL := API + "@" + Updater.LatestVersion + "/" + Updater.LatestVersion + "/" + Updater.FileName
			fmt.Println("·CDN加速地址:\n " + cdnURL)
			err = downloadFile(cdnURL, "./temp")
			if err != nil {
				fmt.Println("·第" + strconv.Itoa(i+1) + "次加速尝试失败")
				log.Println(err)
				fmt.Println("────────────────────────────────────────────────────────────────────────────────────────")
			} else {
				break
			}
		}

		//7.下载成功则进行下一步，否则直接从advancedfx原仓库下载
		exist, err := isFileExisted("./temp/" + Updater.FileName)
		if err != nil {
			log.Println(err)
			pause()
			os.Exit(37)
		} else if exist == false {
			fmt.Println("·正在从GitHub原地址下载...")
			err = downloadFile(Updater.Url, "./temp/")
			if err != nil {
				fmt.Println("·原地址下载失败，请检查网络连接")
				log.Println(err)
				pause()
				os.Exit(9)
			}
		}

		//8.解压到临时目录"./temp/"检查"changelog.xml和"hlae.exe"的正确性，然后移动文件，覆盖原目录
		fmt.Println("·下载成功，正在解压...")
		tempDir := "./temp/hlae/"
		_ = os.RemoveAll(tempDir)
		err = decompress("./temp/"+Updater.FileName, tempDir)
		if err != nil {
			fmt.Println("·解压失败")
			log.Println(err)
			pause()
			os.Exit(10)
		} else {
			ok, err := isFileExisted(tempDir + "hlae.exe")
			if err != nil {
				log.Println(err)
				pause()
				os.Exit(11)
			} else if ok == false {
				log.Println(errors.New("successfully unzipped but no file is found"))
				pause()
				os.Exit(12)
			}
		}

		//移动，覆盖原目录
		fmt.Println("·解压成功，正在移动文件...")
		err = copyDir(tempDir, usr.HomeDir+"/AppData/Local/AkiVer/hlae")
		if err != nil {
			log.Println(err)
			pause()
			os.Exit(13)
		}
		fmt.Println("·HLAE安装/更新成功！")

		//9.生成/更新"Version"文件，格式"2.102.0"
		generateVersion(Updater.LatestVersion, usr.HomeDir+"/AppData/Local/AkiVer/hlae/version")

		//更新/安装成功的提示
		fmt.Println("────────────────────────────────────────────────────────────────────────────────────────")
		if Updater.HlaeExist == true {
			fmt.Println("·HLAE更新完成，当前版本号：", Updater.LatestVersion)
		} else {
			fmt.Println("·HLAE安装完成，当前版本号：", Updater.LatestVersion,
				"\n·请在CSGO Demos Manager的设置中点击`启用HLAE`")
		}
		fmt.Println("────────────────────────────────────────────────────────────────────────────────────────")
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
	//检查FFMPEG是否存在

	//8.通过检测"%HOMEDIR%/AppData/Local/AkiVer/hlae/ffmpeg/bin/ffmpeg.exe"是否存在判断是否安装了FFmpeg
	Updater.FFmpegExist, err = isFileExisted(usr.HomeDir + "/AppData/Local/AkiVer/hlae/ffmpeg/bin/ffmpeg.exe")
	if err != nil {
		log.Println(err)
		pause()
		os.Exit(59)
	}

	//9.获取FFMPEG最新版本
	exist, err := isFileExisted(usr.HomeDir + "/AppData/Local/AkiVer/hlae/ffmpeg/bin/ffmpeg.exe")
	if err != nil {
		log.Println(err)
		pause()
		os.Exit(33)
	} else if exist == true && Updater.FFmpegVersion == "" {
		confirm := "y"
		fmt.Printf("·检测到已有FFmpeg，是否安装最新版（Y/N）：")
		fmt.Scanf("%v", &confirm)
		if confirm == "n" || confirm == "N" {
			pause()
			os.Exit(0)
		}
	}

	fmt.Println("·正在获取FFMPEG最新版本信息...")
	var ver string
	jsonData, err = getHttpData(Updater.FFmpegAPI)
	if err != nil {
		log.Println(err)
		fmt.Println("·FFmpeg API访问失败，正在使用备用API...")
		for i, API := range Updater.FFmpegCdnAPI {
			jsonData, err = getHttpData(API + "/release.json")
			if err != nil {
				fmt.Println("·第" + strconv.Itoa(i) + "个备用API访问失败")
				log.Println(err)
			} else {
				break
			}
		}
	}

	ver, err = getFFmpegLatestVersion(jsonData)
	if err != nil {
		log.Println(err)
		pause()
		os.Exit(89)
	}

	//10.判断是否要下载/更新FFmpeg
	res = strings.Compare(ver, Updater.FFmpegVersion)
	if Updater.FFmpegExist == true && res < 0 {
		fmt.Println("·发生异常，本地版本号>最新版本号，请检查本地FFmpeg文件")
		pause()
		os.Exit(8)
	} else if Updater.FFmpegExist == true && res == 0 {
		fmt.Println("·FFmpeg已是最新版本")
		fmt.Println("────────────────────────────────────────────────────────────────────────────────────────")
	} else if Updater.FFmpegExist == false || res > 0 {
		Updater.FFmpegVersion = ""
		//FFmpeg不存在或者版本低于最新版时更新
		//Linux 64位地址 https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz
		//Windows 64位地址 需要版本号 shared/static https://ffmpeg.zeranoe.com/builds/win64/static/ffmpeg-4.3.1-win64-static.zip
		//MacOS 64位地址 需要版本号 shared/static https://ffmpeg.zeranoe.com/builds/macos64/static/ffmpeg-4.3.1-macos64-static.zip
		fmt.Println("·最新版本:", ver)
		fmt.Println("·正在尝试加速下载...")
		fileName := "ffmpeg-" + ver + "-win64-static.7z"
		for i, API := range Updater.FFmpegCdnAPI {
			cdnURL := API + "@" + ver + "/" + fileName
			fmt.Println("·CDN加速地址:\n" + cdnURL)
			err = downloadFile(cdnURL, "./temp")
			if err != nil {
				fmt.Println("·第" + strconv.Itoa(i+1) + "次加速尝试失败")
				log.Println(err)
				fmt.Println("────────────────────────────────────────────────────────────────────────────────────────")
			} else {
				break
			}
		}

		//检查是否下载成功，否则从原始地址下载
		exist, err := isFileExisted("./temp/" + fileName)
		if err != nil {
			log.Println(err)
			pause()
			os.Exit(37)
		} else if exist == false {
			fileName = "ffmpeg-" + ver + "-win64-shared.zip"
			originalURL := "https://ffmpeg.zeranoe.com/builds/win64/static/" + fileName
			fmt.Println("·正在从GitHub原地址下载...\n - " + originalURL)
			err = downloadFile(originalURL, "./temp/")
			if err != nil {
				fmt.Println("·原地址下载失败，请检查网络连接")
				log.Println(err)
				pause()
				os.Exit(9)
			}
		}

		//8.解压到临时目录"./temp/"检查"ffmpeg.exe"的正确性，然后移动文件，覆盖原目录
		fmt.Println("·下载成功，正在解压...")
		tempDir := "./temp/ffmpeg/"
		_ = os.RemoveAll(tempDir)
		//TODO 更换7z解压包 现在的包太重了 P.S. 压缩后还好
		err = decompress("./temp/"+fileName, tempDir)
		if err != nil {
			fmt.Println("·解压失败")
			log.Println(err)
			pause()
			os.Exit(10)
		} else {
			ok, err := isFileExisted(tempDir + "bin/ffmpeg.exe")
			if err != nil {
				log.Println(err)
				pause()
				os.Exit(11)
			} else if ok == false {
				log.Println(errors.New("successfully unzipped but no file is found"))
				pause()
				os.Exit(12)
			}
		}

		//移动，覆盖原目录
		fmt.Println("·解压成功，正在移动文件...")
		err = copyDir(tempDir, usr.HomeDir+"/AppData/Local/AkiVer/hlae/ffmpeg")
		if err != nil {
			log.Println(err)
			pause()
			os.Exit(13)
		}
		fmt.Println("·FFmpeg安装/更新成功！")
		Updater.FFmpegVersion = ver
		//更新/安装成功的提示
		fmt.Println("────────────────────────────────────────────────────────────────────────────────────────")
		if Updater.FFmpegExist == true {
			fmt.Println("·FFmpeg更新完成，当前版本号：", Updater.FFmpegVersion)
		} else {
			fmt.Println("·FFMPEG安装完成，当前版本号：", Updater.FFmpegVersion)
		}
		fmt.Println("────────────────────────────────────────────────────────────────────────────────────────")
	}

	pause()
}
