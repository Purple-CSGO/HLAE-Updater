@echo off
chcp 65001

:: 编译APP可执行文件
go build -o %GOPATH%/bin/HLAE-Updater.exe app.go
:: 编译CLI命令行可执行文件
cd CLI
go build
move "./CLI.exe" "%GOPATH%/bin/HLAE-Updater-CLI.exe"
cd %GOPATH%/bin/
echo ----------------------- APP -----------------------
:: 运行APP可执行文件
HLAE-Updater.exe
echo ----------------------- CLI -----------------------
:: 运行CLI命令行可执行文件
HLAE-Updater-CLI.exe
echo ----------------------- END -----------------------
