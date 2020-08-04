package main

import "fmt"

/*
流程(CLI)：
1.读取settings.json，无则赋默认值，创建文件
2.通过检测"%HOMEDIR%/AppData/Local/AkiVer/"是否存在判断是否安装了CSGO DEMOS MANAGER，否则退出
3.通过检测"%HOMEDIR%/AppData/Local/AkiVer/hlae/hlae.exe"是否存在判断是否安装了HLAE，否则跳过XML解析
4.利用API获取包含HLAE仓库信息的JSON文件并解析，获得版本号和下载地址
5.解析包含本地版本信息的XML文件"HLAE/changelog.xml"，获得当前版本
6.判断是否要下载/更新，是则利用CDN加速尝试下载HLAE-Release仓库的文件
7.下载成功则进行下一步，否则直接从advancedfx原仓库下载
8.解压到临时目录"./temp/"检查"changelog.xml和"hlae.exe"的正确性，然后移动文件，覆盖原目录
9.生成/更新"Version"文件，格式"2.102.0"
*/

func main() {
	fmt.Println("这是CLI命令行工具")
}
