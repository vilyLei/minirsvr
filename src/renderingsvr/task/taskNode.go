package task

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"renderingsvr.com/filesys"
	"renderingsvr.com/message"

)

// go mod init renderingsvr.com/task

type ResLoadParam struct {
	Url      string
	TaskName string
	PathDir  string
}

func GetFileNameFromUrl(url string) string {
	nameStr := url[strings.LastIndex(url, "/")+1 : len(url)]
	i := strings.LastIndex(nameStr, "?")
	if i > 0 {
		nameStr = nameStr[0:i]
	}
	return nameStr
}
func loadRenderingRes(out chan<- int, param ResLoadParam) bool {

	resUrl := param.Url
	// Get the data
	resp, loadErr := http.Get(resUrl)
	if loadErr != nil {
		fmt.Printf("load a file failed, loadErr: %v\n", loadErr)
		// panic(loadErr)

		out <- 0
		return false
	}
	defer resp.Body.Close()

	data, wErr := ioutil.ReadAll(resp.Body)
	if wErr != nil {
		fmt.Printf("write a file failed,wErr: %v\n", wErr)

		out <- 0
		// panic(wErr)
		return false
	}
	if len(data) < 300 {
		fmt.Println("data: ", data)
		fmt.Println("data len: ", len(data))
		str := string(data)
		fmt.Println("data to str: ", str)
		strI := strings.Index(str, "Error:")
		fmt.Println("strI: ", strI)
		if strI > 0 {
			out <- 0
			panic("load error")
			return false
		}
	}
	nameStr := GetFileNameFromUrl(resUrl)
	fmt.Println("remote res nameStr: ", nameStr)
	fmt.Println("remote res pathDir: ", param.PathDir)
	ioutil.WriteFile(param.PathDir+nameStr, data, 0777)

	fmt.Println("load a remote res file success !!!")
	out <- 1
	return true
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
	// path = ""

	// cmdParams = `python D:\dev\webProj\voxblender\pysrc\programs\tutorials\bcRenderShell.py -- renderer=D:/programs/blender/blender.exe rmodule=D:/dev/webProj/voxblender/pysrc/programs/tutorials/modelFileRendering.py rtaskDir=D:/dev/webProj/voxblender/models/model02/`
	fmt.Println("rtaskDir: ", rtaskDir)
	cmdParams = `python D:\dev\webProj\voxblender\pysrc\programs\tutorials\bcRenderShell.py -- renderer=D:/programs/blender/blender.exe rmodule=D:/dev/webProj/voxblender/pysrc/programs/tutorials/modelFileRendering.py rtaskDir=` + rtaskDir
	return cmdParams
}

func StartupATask(resDirPath string, rendererPath string, taskID int64, times int64, taskName string, resUrl string) {

	fmt.Println("StartupATask(), resDirPath: ", resDirPath)

	hasStatusDir := filesys.HasSceneResDir(resDirPath)
	fmt.Println("#### ### hasStatusDir: ", hasStatusDir)

	var configParam filesys.RenderingConfigParam
	configParam.ResourceType = "none"
	configParam.Models = "[]"
	configParam.TaskID = taskID
	configParam.Times = times
	configParam.Progress = 0
	configParam.OutputPath = ""
	if !hasStatusDir {
		flag := filesys.CreateDirWithPath(resDirPath)
		if flag {
			filesys.CreateRenderingConfigFileToPath(resDirPath, rendererPath, configParam)
		}
	}

	// hasStatusFile := filesys.HasSceneResStatusJson(resDirPath)
	// fmt.Println("#### ### hasStatusFile: ", hasStatusFile)
	// req remote rendering res
	fmt.Println("StartupATask(), ready to load rendering resource !")
	loaderChannel := make(chan int, 1)
	var resParam ResLoadParam
	// resParam.Url = "http://www.artvily.com/static/assets/obj/base.obj"
	// resParam.Url = "http://www.artvily.com/static/assets/obj/cylinder_obj.zip"
	resParam.Url = resUrl
	resParam.TaskName = taskName
	resParam.PathDir = resDirPath

	go loadRenderingRes(loaderChannel, resParam)

	for flag := range loaderChannel {
		len := len(loaderChannel)
		if len == 0 {
			fmt.Println("loader_channel flag: ", flag)
			close(loaderChannel)
		}
	}
	fmt.Println("StartupATask(), ready to load rendering resource finish !")
	nameStr := GetFileNameFromUrl(resParam.Url)
	configParam.ResourceType = "obj"
	configParam.Models = `["` + nameStr + `"]`
	filesys.CreateRenderingConfigFileToPath(resDirPath, rendererPath, configParam)

	fmt.Println("StartupATask(), ready exec the exe program !")
	rendererExeName := "./renderer.exe"
	cmdParams := getCmdParamsString(rendererExeName, resDirPath)

	fmt.Println("StartupATask(), exe cmdParams: ", cmdParams)

	cmd := exec.Command("cmd.exe", "/c", "start "+cmdParams)
	cmd.Run()
}

type TaskExecNode struct {
	Uid           int64
	Index         int
	Desc          string
	Phase         string
	RunningStatus int
	RstData       message.RenderingSTChannelData

	PathDir  string
	TaskName string
	ResUrl   string
	FilePath string
	TaskID   int64
	Times    int64
	Progress int
}

func (self *TaskExecNode) Init() *TaskExecNode {
	self.Uid = 0
	self.Index = 0
	self.Desc = "a TaskExecNode instance."
	self.PathDir = ""
	self.TaskName = ""
	self.ResUrl = ""
	self.FilePath = "renderingStatus.json"
	self.RunningStatus = 0
	self.TaskID = 1
	self.Phase = "running"
	self.Times = 1
	self.Progress = 0
	return self
}
func (self *TaskExecNode) Reset() *TaskExecNode {
	self.Init()
	return self
}

func (self *TaskExecNode) Exec() *TaskExecNode {
	if self.RunningStatus == 1 {
		if self.PathDir == "" {
			fmt.Println("Exec(), ready startup a new task")

			rendererPath := "./renderer.exe"

			if self.TaskName == "" {
				self.TaskName = "scene01"
			}
			// resDirPath := "./static/sceneres/" + self.TaskName + "/"
			rootDir, err := os.Getwd()
			if err != nil {
				fmt.Println("os.Getwd(), err: %v", rootDir)
				rootDir = "."
			}
			fmt.Println("rootDir: ", rootDir)
			resDirPath := rootDir + "/static/sceneres/" + self.TaskName + "/"

			self.PathDir = resDirPath
			self.FilePath = self.PathDir + "renderingStatus.json"

			self.Times++
			self.Progress = 0
			self.RunningStatus = 2
			go StartupATask(resDirPath, rendererPath, self.TaskID, self.Times, self.TaskName, self.ResUrl)
		} else {
			fmt.Println("Exec(), error:  self.pathDir is not empty.")
		}
	}
	return self
}
func (self *TaskExecNode) CheckRendering() *TaskExecNode {
	if self.RunningStatus == 2 {
		if self.PathDir != "" {
			fmt.Println("CheckRendering(), task checking ...")
			hasStatusFile := filesys.HasSceneResStatusJson(self.PathDir)
			fmt.Println("CheckRendering(), >>> filePath: ", self.FilePath)
			fmt.Println("CheckRendering(), >>> hasStatusFile: ", hasStatusFile)

			if hasStatusFile {

				ins, err := filesys.ReadRenderingStatusJson(self.PathDir)
				if err == nil {
					// fmt.Println("CheckRendering(), ins: ", ins)
					task := ins.Rendering_task
					taskID := task.TaskID
					times := task.Times
					progress := task.Progress
					phase := task.Phase
					self.Progress = progress
					fmt.Println("CheckRendering(), taskID: ", taskID, ", times: ", times)
					fmt.Println("CheckRendering(), ### progress: ", progress, "%")
					if taskID == self.TaskID && times == self.Times && progress >= 100 {
						fmt.Println("CheckRendering(), >>> phase: ", phase)
						if phase == "error" {
							fmt.Println("CheckRendering(), >>> rendering task process has a error !!!")
						}
						fmt.Println("CheckRendering(), >>> rendering task process finish !!!")
						fmt.Println("CheckRendering(), >>> waiting for the next task ...")
						self.RunningStatus = 0
						self.PathDir = ""
						self.TaskName = ""
					}
				} else {
					fmt.Println("CheckRendering(), >>> read renderingStatusJson failed !!!")
				}
			}
		}
	}
	return self
}

func (self *TaskExecNode) IsWaitingTask() bool {
	return self.RunningStatus == 5
}
func (self *TaskExecNode) IsFree() bool {
	return self.RunningStatus == 0
}
