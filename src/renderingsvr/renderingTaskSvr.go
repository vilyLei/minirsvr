package main

import (
	"fmt"
	"os"

	"renderingsvr.com/filesys"
	"renderingsvr.com/message"
	"renderingsvr.com/svr"
	"renderingsvr.com/task"

)

// go mod init renderingsvr.com/main
// go mod edit -replace renderingsvr.com/message=./message
// go mod edit -replace renderingsvr.com/filesys=./filesys
// go mod edit -replace renderingsvr.com/task=./task
// go mod edit -replace renderingsvr.com/svr=./svr

// go build renderingTasksvr.go
// go build -o ./ renderingTasksvr.go
// go run renderingTasksvr.go
// go run renderingTasksvr.go 9092 auto

// 调用 http://localhost:9092/rendering 这个请求就可以本地测试渲染任务调度

func main() {
	fmt.Println("renderingTaskSvr init ...")
	rootDir, err := os.Getwd()
	if err != nil {
		fmt.Println("os.Getwd(), err: %v", rootDir)
	}
	fmt.Println("rootDir: ", rootDir)

	argsLen := len(os.Args)
	// fmt.Println("argsLen: ", argsLen)
	var portStr = "9092"
	var param0 = ""
	if argsLen > 1 {
		portStr = "" + os.Args[1]
		fmt.Println("portStr: ", portStr)
	}
	if argsLen > 2 {
		param0 = "" + os.Args[2]
		fmt.Println("param0: ", param0)
	}

	message.Init()
	resDirPath := ".\\static\\sceneres\\scene01\\"
	flagBool, _ := filesys.PathExists(resDirPath)

	fmt.Println("renderingTaskSvr flagBool: ", flagBool)
	if flagBool {

		fmt.Println("renderingTaskSvr path exist ok ...")
	}
	var taskNode task.TaskExecNode
	taskNode.Uid = 101
	fmt.Println("renderingTaskSvr taskNode: ", taskNode)
	svr.AutoCheckRTask = false
	if param0 == "auto" {
		fmt.Println("auto checking rendering task")
		svr.AutoCheckRTask = true
	}
	svr.Init()
	fmt.Println("renderingTaskSvr end ...")
}
