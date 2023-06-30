package main

import (
	"fmt"
	"os"
	"strings"

	"renderingsvr.com/filesys"
	"renderingsvr.com/message"
	"renderingsvr.com/svr"
)

// go mod init renderingsvr.com/main
// go mod edit -replace renderingsvr.com/message=./message
// go mod edit -replace renderingsvr.com/filesys=./filesys
// go mod edit -replace renderingsvr.com/task=./task
// go mod edit -replace renderingsvr.com/svr=./svr

// go build dsrdiffusion.go
// go build -o ./ dsrdiffusion.go
// go run dsrdiffusion.go
// go run dsrdiffusion.go 9092 auto
// go run dsrdiffusion.go 9092 false remote-debug localProc
// go run dsrdiffusion.go -- port=9092 auto=true rsvr=localhost proc=local
// dsrdiffusion.exe -- port=9092 auto=true rsvr=localhost proc=local

// 调用 http://localhost:9092/rendering 这个请求就可以本地测试渲染任务调度

// go run dsrdiffusion.go -- port=9092 auto=false rsvr=remote-debug proc=local
// go run dsrdiffusion.go -- port=9092 auto=false rsvr=remote-debug proc=local
// go run dsrdiffusion.go -- port=9092 auto=true rsvr=localhost proc=local
// go run dsrdiffusion.go -- port=9092 auto=true rsvr=localhost proc=remote

var startupParam filesys.SysStartupParam

func main() {

	fmt.Println("dsrdiffusion init ...")

	rootDir, errOGT := os.Getwd()
	if errOGT != nil {
		fmt.Println("os.Getwd(), errOGT: %v", errOGT)
	} else {
		rootDir = strings.ReplaceAll(rootDir, `\`, `/`) + "/"
	}
	fmt.Println("rootDir: ", rootDir)

	argsLen := len(os.Args)
	filesys.ReadSysConfFile()

	if argsLen > 3 {
		var cmdMap = make(map[string]string)
		for i := 2; i < argsLen; i++ {
			parts := strings.Split(os.Args[i], "=")
			if len(parts) > 1 {
				cmdMap[parts[0]] = parts[1]
			}
			// fmt.Println("os.Args[", i, "]: ", os.Args[i])
		}
		for k, v := range cmdMap {
			fmt.Println("key: ", k, ", value: ", v)
		}
		fmt.Println("dsrdiffusion cmds direct parsing end ...")
		startupParam.SetParam(cmdMap)
	} else {
		fmt.Println("dsrdiffusion cmds parsing end from conf file ...")
		startupParam.SetParam(filesys.SysConfMap)
		// fmt.Println("Error cmd params !!!")
		// fmt.Println("example: svr.exe -- port=9092 auto=true rsvr=remote-release proc=local")
	}

	message.Init()

	fmt.Println("dsrdiffusion startupParam.ProcType: ", startupParam.ProcType)
	fmt.Println("dsrdiffusion startupParam.PortStr: ", startupParam.PortStr)
	fmt.Println("dsrdiffusion startupParam.SvrRootUrl: ", startupParam.SvrRootUrl)
	fmt.Println("dsrdiffusion startupParam.AutoCheckRTask: ", startupParam.AutoCheckRTask)

	svr.RootDir = rootDir
	svr.AutoCheckRTask = startupParam.AutoCheckRTask

	svr.Init(startupParam.PortStr, startupParam.SvrRootUrl)
	fmt.Println("dsrdiffusion end ...")
}
