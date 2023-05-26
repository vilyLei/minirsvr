package main

import (
	"fmt"

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

func main() {
	fmt.Println("renderingTaskSvr init ...")
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
	svr.Init()
	fmt.Println("renderingTaskSvr end ...")
}
