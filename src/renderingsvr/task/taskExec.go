package task

import (
	"fmt"
	"os/exec"
	"strings"

	"renderingsvr.com/filesys"
)

// go mod init renderingsvr.com/task

func GetFileNameAndSuffixFromUrl(url string) (string, string) {
	nameStr := url[strings.LastIndex(url, "/")+1 : len(url)]
	i := strings.LastIndex(nameStr, "?")
	if i > 0 {
		nameStr = nameStr[0:i]
	}
	parts := strings.Split(nameStr, ".")
	return nameStr, strings.ToLower(parts[1])
}
func getCmdParamsString(rendererExeName string, paths ...string) string {

	// path := ".\\static\\sceneres\\scene01\\"
	path := "./static/sceneres/scene01/"
	if len(paths) > 0 {
		path = paths[0]
	}
	deviceType := "d3d12"
	//renderer.exe "./static/scene/car001/" --device-type "d3d12"
	// taskIDStr := strconv.FormatInt(int64(taskID), 10)
	// renderingTimesStr := strconv.FormatInt(int64(renderingTimes), 10)
	// cmdParams := "./exeForGo.exe .\\static\\sceneres\\scene01\\ " + taskIDStr + " " + renderingTimesStr
	cmdParams := rendererExeName + ` ` + path + ` --device-type ` + deviceType + ``
	path = strings.ReplaceAll(path, `\`, `/`)
	rtaskDir := path
	fmt.Println("### path: ", path)
	// path = "D:/dev/webdev/minirsvr/src/renderingsvr/static/sceneres/modelTask01/"
	path = " --rcp " + `` + path + ``
	rendererExeName = "D:/dev/rendering/minirenderer/rendererRelease/TerminusApp.exe"
	cmdParams = rendererExeName + path

	fmt.Println("rtaskDir: ", rtaskDir)
	cmdParams = filesys.GetRenderCMD(rtaskDir)
	return cmdParams
}

func execRenderer(cmdParams string) {
	cmd := exec.Command("cmd.exe", "/c", "start "+cmdParams)
	cmd.Run()
}
func execModelExport(modelFilePath string) {
	cmdParams := filesys.GetModelExportCMD(modelFilePath)
	fmt.Println("execModelExport(), exe cmdParams: ", cmdParams)
	cmd := exec.Command("cmd.exe", "/c", "start "+cmdParams)
	//stdout, err := cmd.StdoutPipe() // 从命令行获取输出内容
	cmd.Run()
}
func StartupATask(rootDir string, resDirPath string, rendererPath string, rtNode TaskExecNode) {

	fmt.Println("StartupATask(), resDirPath: ", resDirPath)

	taskID := rtNode.TaskID
	taskName := rtNode.TaskName
	resUrl := rtNode.ResUrl
	times := rtNode.Times

	hasStatusDir := filesys.HasSceneResDir(resDirPath)
	fmt.Println("#### ### hasStatusDir: ", hasStatusDir)
	fmt.Println("#### ### rootDir: ", rootDir)

	NotifyTaskInfoToSvr("task_rendering_load_res", 0, taskID, taskName)

	var configParam filesys.RenderingConfigParam
	configParam.Resolution = rtNode.Resolution
	configParam.Camdvs = rtNode.Camdvs
	configParam.BGTransparent = rtNode.BGTransparent
	configParam.RNode = rtNode.RNode
	configParam.ResourceType = "none"
	configParam.Models = []string{""}
	configParam.TaskID = taskID
	configParam.Times = times
	configParam.Progress = 0
	configParam.RootDir = rootDir
	configParam.OutputPath = ""

	if !hasStatusDir {
		flag := filesys.CreateDirWithPath(resDirPath)
		if flag {
			// filesys.CreateRenderingConfigFileToPath(resDirPath, rendererPath, configParam)
		}
	}

	if rtNode.CheckModelDrcStatus() == 0 {
		fmt.Println("#### ### need exec export model to drc.")
		execModelExport(filesys.GetModelFilePath(resDirPath, resUrl))
	}

	// hasStatusFile := filesys.HasSceneResStatusJson(resDirPath)
	// fmt.Println("#### ### hasStatusFile: ", hasStatusFile)
	// req remote rendering res
	fmt.Println("StartupATask(), ready to load rendering resource !")
	var resParam ResLoadParam
	// resParam.Url = "http://www.artvily.com/static/assets/obj/base.obj"
	// resParam.Url = "http://www.artvily.com/static/assets/obj/cylinder_obj.zip"

	resParam.Url = resUrl
	resParam.TaskName = taskName
	resParam.PathDir = resDirPath
	// check the file exists
	// go loadRenderingRes(loaderChannel, resParam)
	if !filesys.CheckModelFileExists(resDirPath, resUrl) {

		loaderChannel := make(chan int, 1)
		go filesys.DownloadFile(loaderChannel, resDirPath, resUrl, taskID, taskName)

		for flag := range loaderChannel {
			len := len(loaderChannel)
			if len == 0 {
				fmt.Println("loader_channel flag: ", flag)
				close(loaderChannel)
			}
		}
	}

	fmt.Println("StartupATask(), ready to load rendering resource finish !")
	nameStr, suffix := GetFileNameAndSuffixFromUrl(resParam.Url)
	configParam.ResourceType = suffix
	configParam.Models = []string{nameStr}
	// 这里应该加一个锁
	filesys.CreateRenderingConfigFileToPath(resDirPath, rendererPath, configParam)

	fmt.Println("StartupATask(), ready exec the exe program !")
	rendererExeName := "./renderer.exe"
	cmdParams := getCmdParamsString(rendererExeName, resDirPath)

	fmt.Println("StartupATask(), exe cmdParams: ", cmdParams)

	NotifyTaskInfoToSvr("task_rendering_begin", 0, taskID, taskName)
	execRenderer(cmdParams)
}
