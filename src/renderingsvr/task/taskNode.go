package task

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"renderingsvr.com/filesys"
	"renderingsvr.com/message"
)

// go mod init renderingsvr.com/task

var TaskReqSvrUrl string = ""

func NotifyTaskInfoToSvr(phase string, progress int, taskId int64, taskName string) {
	progressStr := strconv.Itoa(progress)
	taskIdStr := strconv.FormatInt(taskId, 10)
	url := TaskReqSvrUrl + "?phase=" + phase + "&progress=" + progressStr
	if taskId > 0 {
		url += "&taskid=" + taskIdStr + "&taskname=" + taskName
	}
	resp, err := http.Get(url)
	flag := true
	if err != nil {
		flag = false
		fmt.Printf("taskNode::NotifyTaskInfoToSvr() get url failed, err: %v\n", err)

	} else {
		defer resp.Body.Close()
	}
	if flag {
		data, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			switch phase {
			case "running":
				fmt.Println("taskNode::NotifyTaskInfoToSvr() receive running req info, ", string(data))
			case "finish":
				fmt.Println("taskNode::NotifyTaskInfoToSvr() receive finish req info, ", string(data))
			case "rtaskerror":
				fmt.Println("taskNode::NotifyTaskInfoToSvr() receive rendering task error req info, ", string(data))
			default:
			}
		}
	}
}

type ResLoadParam struct {
	Url      string
	TaskName string
	PathDir  string
}
type TaskOutputParam struct {
	PicPath  string
	TaskName string
	TaskID   int64
	Error    bool
}

func GetFileNameFromUrl(url string) string {
	nameStr := url[strings.LastIndex(url, "/")+1 : len(url)]
	i := strings.LastIndex(nameStr, "?")
	if i > 0 {
		nameStr = nameStr[0:i]
	}
	return nameStr
}
func GetFileNameAndSuffixFromUrl(url string) (string, string) {
	nameStr := url[strings.LastIndex(url, "/")+1 : len(url)]
	i := strings.LastIndex(nameStr, "?")
	if i > 0 {
		nameStr = nameStr[0:i]
	}
	parts := strings.Split(nameStr, ".")
	return nameStr, strings.ToLower(parts[1])
}
func loadRenderingRes(out chan<- int, param ResLoadParam) bool {

	resUrl := param.Url
	fmt.Println("loadRenderingRes(), resUrl: ", resUrl)

	// Get the data
	resp, loadErr := http.Get(resUrl)
	if loadErr != nil {
		fmt.Printf("load a file failed, loadErr: %v\n", loadErr)

		out <- 0
		return false
	}
	defer resp.Body.Close()

	data, wErr := ioutil.ReadAll(resp.Body)
	if wErr != nil {
		fmt.Printf("write a file failed,wErr: %v\n", wErr)

		out <- 0
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

func StartupATask(rootDir string, resDirPath string, rendererPath string, taskID int64, times int64, taskName string, resUrl string) {

	fmt.Println("StartupATask(), resDirPath: ", resDirPath)

	hasStatusDir := filesys.HasSceneResDir(resDirPath)
	fmt.Println("#### ### hasStatusDir: ", hasStatusDir)
	fmt.Println("#### ### rootDir: ", rootDir)

	NotifyTaskInfoToSvr("task_rendering_load_res", 0, taskID, taskName)

	var configParam filesys.RenderingConfigParam
	configParam.ResourceType = "none"
	configParam.Models = "[]"
	configParam.TaskID = taskID
	configParam.Times = times
	configParam.Progress = 0
	configParam.RootDir = rootDir
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
	nameStr, suffix := GetFileNameAndSuffixFromUrl(resParam.Url)
	configParam.ResourceType = suffix
	configParam.Models = `["` + nameStr + `"]`
	filesys.CreateRenderingConfigFileToPath(resDirPath, rendererPath, configParam)

	fmt.Println("StartupATask(), ready exec the exe program !")
	rendererExeName := "./renderer.exe"
	cmdParams := getCmdParamsString(rendererExeName, resDirPath)

	fmt.Println("StartupATask(), exe cmdParams: ", cmdParams)

	NotifyTaskInfoToSvr("task_rendering_begin", 0, taskID, taskName)
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

	PathDir     string
	TaskName    string
	ResUrl      string
	FilePath    string
	TaskID      int64
	Times       int64
	ReqProgress int
	Progress    int
	TaskOutput  TaskOutputParam
	RootDir     string
}

func (self *TaskExecNode) Init() *TaskExecNode {
	self.Uid = 0
	self.Index = 0
	self.Desc = "a TaskExecNode instance."
	self.PathDir = ""
	self.TaskName = ""
	self.ResUrl = ""
	self.FilePath = "renderingStatus.json"
	self.RootDir = ""
	self.RunningStatus = 0
	self.TaskID = 1
	self.Phase = "running"
	self.Times = 1
	self.ReqProgress = 0
	self.Progress = 0
	self.TaskOutput.PicPath = ""
	self.TaskOutput.Error = false
	return self
}
func (self *TaskExecNode) Reset() *TaskExecNode {
	self.Init()
	return self
}
func (self *TaskExecNode) CheckTaskStatus() (bool, int) {

	if self.PathDir != "" {
		fmt.Println("CheckTaskStatus(), task checking ...")
		hasStatusFile := filesys.HasSceneResStatusJson(self.PathDir)
		fmt.Println("CheckTaskStatus(), >>> filePath: ", self.FilePath)
		fmt.Println("CheckTaskStatus(), >>> hasStatusFile: ", hasStatusFile)

		if hasStatusFile {

			ins, err := filesys.ReadRenderingStatusJson(self.PathDir)
			if err == nil {
				// fmt.Println("CheckTaskStatus(), ins: ", ins)
				task := ins.Rendering_task
				taskID := task.TaskID
				times := task.Times
				progress := task.Progress
				phase := task.Phase
				self.Progress = progress

				fmt.Println("CheckTaskStatus(), taskID: ", taskID, ", times: ", times)
				fmt.Println("CheckTaskStatus(), ### progress: ", progress, "%")
				taskFlag := false
				if taskID == self.TaskID && times == self.Times {
					if progress >= 100 {

						fmt.Println("CheckTaskStatus(), >>> self.PathDir: ", self.PathDir)
						fmt.Println("CheckTaskStatus(), >>> phase: ", phase)

						taskStatus := 1
						if phase == "error" {
							taskStatus = -1
							fmt.Println("CheckTaskStatus(), >>> rendering task process has a error !!!")
						} else {
							// check output pic file
							taskFlag, _ = filesys.CheckPicFileInCurrDir(self.PathDir)
							fmt.Println("CheckTaskStatus(), >>> has output rendering pic: ", taskFlag)
							if !taskFlag {
								taskStatus = -1
							}
						}
						return taskFlag, taskStatus
					}
					return taskFlag, 5
				}
				return taskFlag, 6
			} else {
				fmt.Println("CheckTaskStatus(), >>> read renderingStatusJson failed !!!")
			}
		}
	}
	return false, 0
}
func (self *TaskExecNode) toTaskFinish() *TaskExecNode {

	fmt.Println("toTaskFinish(), >>> pathDif: ", self.PathDir)
	picFlag, picNames := filesys.CheckPicFileInCurrDir(self.PathDir)
	self.TaskOutput.PicPath = ""
	if picFlag {
		// ready send the pic to a remote data center server
		picFilePath := self.PathDir + picNames[0]
		fmt.Println("toTaskFinish(), >>> picFilePath: ", picFilePath)
		self.TaskOutput.PicPath = picFilePath
	}
	self.TaskOutput.Error = false
	self.TaskOutput.TaskID = self.TaskID
	self.TaskOutput.TaskName = self.TaskName

	self.RunningStatus = 0
	self.PathDir = ""
	self.TaskName = ""
	fmt.Println("toTaskFinish(), >>> waiting for the next task ...")
	return self
}
func (self *TaskExecNode) toTaskError() *TaskExecNode {

	fmt.Println("toTaskError(), >>> pathDif: ", self.PathDir)

	self.TaskOutput.Error = true
	self.TaskOutput.PicPath = ""
	self.TaskOutput.TaskID = self.TaskID
	self.TaskOutput.TaskName = self.TaskName

	self.RunningStatus = 0
	self.PathDir = ""
	self.TaskName = ""
	fmt.Println("toTaskError(), >>> waiting for the next task ...")
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
			_, taskStatus := self.CheckTaskStatus()
			if taskStatus == 1 {
				fmt.Println("Exec(), the task output result is directly available !!!")
				self.toTaskFinish()
			} else {
				if taskStatus != 0 {
					// clear the status info file
					filePath := resDirPath + "renderingStatus.json"
					flag := filesys.RemoveFileWithPath(filePath)
					fmt.Println("Exec(), clear the rtask status info file, flag: ", flag, filePath)
				}
				go StartupATask(self.RootDir, resDirPath, rendererPath, self.TaskID, self.Times, self.TaskName, self.ResUrl)
			}
		} else {
			fmt.Println("Exec(), error:  self.pathDir is not empty.")
		}
	}
	return self
}
func (self *TaskExecNode) CheckRendering() *TaskExecNode {
	if self.RunningStatus == 2 {
		flag, status := self.CheckTaskStatus()
		if flag {
			if status == 1 {
				fmt.Println("CheckRendering(), >>> rendering task process finish !!!")
				self.toTaskFinish()
			} else {
				fmt.Println("CheckRendering(), >>> rendering task process failed A !!!")
				self.toTaskError()
			}
		} else {
			if status == -1 {
				fmt.Println("CheckRendering(), >>> rendering task process failed B !!!")
				self.toTaskError()
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
