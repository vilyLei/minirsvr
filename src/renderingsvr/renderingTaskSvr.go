package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
// go run renderingTasksvr.go 9092 false remote-debug localProc
// go run renderingTasksvr.go -- port=9092 auto=true rsvr=localhost proc=local
// renderingTasksvr.exe -- port=9092 auto=true rsvr=localhost proc=local

// 调用 http://localhost:9092/rendering 这个请求就可以本地测试渲染任务调度

var svrRootUrl string = "http://localhost:9090/"

func decompressFile(compressFilePath string, dstDir string) {
	fmt.Println("decompressFile() init ...")
	archive, err := zip.OpenReader(compressFilePath)
	if err != nil {
		// panic(err)
		fmt.Printf("decompressFile(), decompress error:%v", err)
	}
	defer archive.Close()
	for _, f := range archive.File {
		filePath := filepath.Join(dstDir, f.Name)
		fmt.Println("decompressFile(), unzipping file ", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(dstDir)+string(os.PathSeparator)) {
			fmt.Println("decompressFile(), invalid file path")
			return
		}
		if f.FileInfo().IsDir() {
			fmt.Println("decompressFile(), creating dir...")
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			panic(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}
}

var rcfgFilePath = "static/sys/bpyc/rcfg.json"

func syncRProcRes() {

	fmt.Println("syncRProcRes() init ...")
	srcDir := svrRootUrl + "static/dsrdiffusion/sys/package/"
	dstDir := "static/sys/bpyc/"
	// url := srcDir + "render.zip"

	var files = [2]string{"render.zip", "model.zip"}
	flagValue := 0
	for i := 0; i < len(files); i++ {
		url := srcDir + files[i]
		// fmt.Println("syncRProcRes() init url: ", url)
		loaderChannel := make(chan int, 1)
		go task.DownloadFile(loaderChannel, dstDir, url, 0, "")

		for flag := range loaderChannel {
			len := len(loaderChannel)
			if len == 0 {
				flagValue += flag
				fmt.Println("syncRProcRes(), loaded flag: ", flag, ", url: ", url)
				fmt.Println("syncRProcRes(), dstDir: ", dstDir)
				close(loaderChannel)
				if flag == 1 {
					// decompress
					filePath := dstDir + files[i]
					// fmt.Println("ready compress filePath: ", filePath)
					decompressFile(filePath, dstDir)
				}
			}
		}
	}
	if flagValue == 2 {
		fmt.Println("syncRProcRes() success ...")
		renderPath := "renderer=D:/programs/blender/blender.exe"
		// write rcfg.json file
		rcfg := &filesys.LocalSysCfg
		render := &rcfg.Renderer
		render.MainProc = "python " + dstDir + "render/renderShell.py"
		render.RenerderProc = renderPath + " rmodule=" + dstDir + "render/modelRendering.py"

		mtd := &rcfg.ModelToDrc
		mtd.MainProc = "python " + dstDir + "model/encodeAModelToDrcs.py -- encoder=" + dstDir + "model/draco_encoder.exe"
		mtd.ExportProc = renderPath + " exportPy=" + dstDir + "model/exportMeshesToDrcObjs.py"

		filesys.WriteTxtFileToPath(rcfgFilePath, rcfg.GetJsonString())
	}
	fmt.Println("syncRProcRes() end ...")
}

// go run renderingTasksvr.go -- port=9092 auto=false rsvr=remote-debug proc=local
// go run renderingTasksvr.go -- port=9092 auto=false rsvr=remote-debug proc=local
// go run renderingTasksvr.go -- port=9092 auto=true rsvr=localhost proc=local
// go run renderingTasksvr.go -- port=9092 auto=true rsvr=localhost proc=remote

func main() {

	svrRootUrl = "http://localhost:9091/"

	fmt.Println("renderingTaskSvr init ...")

	rootDir, errOGT := os.Getwd()
	if errOGT != nil {
		fmt.Println("os.Getwd(), errOGT: %v", errOGT)
	} else {
		rootDir = strings.ReplaceAll(rootDir, `\`, `/`) + "/"
	}
	fmt.Println("rootDir: ", rootDir)

	argsLen := len(os.Args)

	var cmdMap = make(map[string]string)
	if argsLen > 3 {
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
		fmt.Println("renderingTaskSvr cmds parsing end ...")
		// return
	} else {
		fmt.Println("Error cmd params !!!")
		fmt.Println("example: svr.exe -- port=9092 auto=true rsvr=remote-release proc=local")
	}
	// fmt.Println("argsLen: ", argsLen)
	var portStr = "9092"
	cmdValue, hasKey := cmdMap["port"]
	if hasKey {
		portStr = cmdValue
	}
	var taskAutoTracing = true
	cmdValue, hasKey = cmdMap["auto"]
	if hasKey {
		taskAutoTracing = cmdValue == "true"
	}
	var procType = "local"
	cmdValue, hasKey = cmdMap["proc"]
	if hasKey {
		procType = cmdValue
	}

	var rsvrType = "local"
	cmdValue, hasKey = cmdMap["rsvr"]
	if hasKey {
		rsvrType = cmdValue
	}
	switch rsvrType {
	case "remote-debug":
		svrRootUrl = "http://www.artvily.com:9093/"
	case "remote-release":
		svrRootUrl = "http://www.artvily.com/"
	default:
	}

	fmt.Println("taskAutoTracing: ", taskAutoTracing)
	// for test
	if procType == "local" {
		rcfgPath := "static/sys/local/config.json"
		filesys.GetLocalSysCfg(rcfgPath)
	} else {
		rcfgPath := rcfgFilePath
		hasFilePath, _ := filesys.PathExists(rcfgPath)
		if hasFilePath {
			filesys.GetLocalSysCfg(rcfgPath)
		} else {
			syncRProcRes()
		}
	}

	message.Init()

	var taskNode task.TaskExecNode
	taskNode.Uid = 101
	fmt.Println("renderingTaskSvr taskNode: ", taskNode)
	svr.RootDir = rootDir
	svr.AutoCheckRTask = taskAutoTracing

	svr.Init(portStr, svrRootUrl)
	fmt.Println("renderingTaskSvr end ...")
}
