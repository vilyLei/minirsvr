package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// go build -o ./ rmsvr.go

/*
{
"rendering-ins":"jetty-scene-renderer",
"rendering-task":

	{
	    "uuid":"rtrt88970-8990",
	    "taskID":1005,
	    "name":"high-image-rendering",
	    "phase":"finish",
	    "times":15
	    "progress":25
	},

"rendering-status":"task:running"
}
*/
type RenderingTask struct {
	Uuid     string `json:"uuid"`
	TaskID   int64  `json:"taskID"`
	Name     string `json:"name"`
	Phase    string `json:"phase"`
	Times    int64  `json:"times"`
	Progress int64  `json:"progress"`
}
type RenderingIns struct {
	Rendering_ins    string        `json:"rendering-ins"`
	Rendering_task   RenderingTask `json:"rendering-task"`
	Rendering_status string        `json:"rendering-status"`
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func readRenderingStatusJson(pathDir string) (RenderingIns, error) {
	pathStr := pathDir + "renderingStatus.json"
	jsonFile, err := os.OpenFile(pathStr, os.O_RDONLY, os.ModeDevice)
	var rIns RenderingIns
	if err == nil {
		defer jsonFile.Close()
		fi, _ := jsonFile.Stat()
		fileBytesTotal := int(fi.Size())
		fmt.Println("fileBytesTotal: ", fileBytesTotal)

		jsonValue, _ := ioutil.ReadAll(jsonFile)

		json.Unmarshal([]byte(jsonValue), &rIns)
		// fmt.Println("readRenderingStatusJson(), rIns.Rendering_ins: ", rIns.Rendering_ins)
		// fmt.Println("readRenderingStatusJson(), rIns.Rendering_task: ", rIns.Rendering_task)
	} else {
		fmt.Printf("readRenderingStatusJson() failed, err: %v\n", err)
	}
	return rIns, err
}
func getCmdParamsString(rendererExeName string, taskID int64, renderingTimes int64, paths ...string) string {
	// taskID := 1003
	// renderingTimes := 11

	path := ".\\static\\sceneres\\scene01\\"
	if len(paths) > 0 {
		path = paths[0]
	}

	taskIDStr := strconv.FormatInt(int64(taskID), 10)
	renderingTimesStr := strconv.FormatInt(int64(renderingTimes), 10)
	// cmdParams := "./exeForGo.exe .\\static\\sceneres\\scene01\\ " + taskIDStr + " " + renderingTimesStr
	cmdParams := rendererExeName + " " + path + " " + taskIDStr + " " + renderingTimesStr
	return cmdParams
}
func HasSceneResDir(resDirPath string) bool {
	// fmt.Println("\nHasSceneResDir(), resDirPath: ", resDirPath)
	hasResDirPath, _ := PathExists(resDirPath)

	// fmt.Println("HasSceneResDir(), hasResDirPath: ", hasResDirPath)
	return hasResDirPath
}
func HasSceneResStatusJson(resDirPath string) bool {
	// fmt.Println("\nHasSceneResStatusJson(), resDirPath: ", resDirPath)
	hasResDirPath, _ := PathExists(resDirPath)

	// fmt.Println("HasSceneResStatusJson(), hasResDirPath: ", hasResDirPath)
	if hasResDirPath {
		filePath := resDirPath + "renderingStatus.json"
		hasFilePath, _ := PathExists(filePath)
		// fmt.Println("HasSceneResStatusJson(), hasFilePath: ", hasFilePath)
		return hasFilePath
	}
	return false
}
func CreateDirWithPath(path string) bool {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		fmt.Printf("CreateDirWithPath() failed, err: %v", err)
		return false
	}
	fmt.Println("CreateDirWithPath(), success !!!")
	return true
}
func CreateRenderingInfoFileToPath(path string, rendererPath string) {

	fileContent := `{
		"renderer-proc":"` + rendererPath + `",
		"renderer-instance":
			{
				"name":"high-image-renderer",
				"status":"stop"
			}
		}`
	filePath := path + "renderingInfo.json"
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("CreateRenderingInfoFile(), err: %v\n", err)
		return
	}
	defer file.Close()
	// 写入内容
	// str := "a text info data.\n" // \n\r表示换行  txt文件要看到换行效果要用 \r\n
	// 写入时，使用带缓存的 *Writer
	writer := bufio.NewWriter(file)
	// for i := 0; i < 3; i++ {
	writer.WriteString(fileContent)
	// }
	//因为 writer 是带缓存的，因此在调用 WriterString 方法时，内容是先写入缓存的
	//所以要调用 flush方法，将缓存的数据真正写入到文件中。
	writer.Flush()
	fmt.Println("CreateRenderingInfoFileToPath(), success !!!")
}

func StartupATask(resDirPath string, rendererPath string, taskID int64, times int64) {

	fmt.Println("StartupATask(), resDirPath: ", resDirPath)

	hasStatusDir := HasSceneResDir(resDirPath)
	fmt.Println("#### ### hasStatusDir: ", hasStatusDir)
	if !hasStatusDir {
		flag := CreateDirWithPath(resDirPath)
		if flag {
			CreateRenderingInfoFileToPath(resDirPath, rendererPath)
		}
	}

	hasStatusFile := HasSceneResStatusJson(resDirPath)
	fmt.Println("#### ### hasStatusFile: ", hasStatusFile)

	// readRenderingStatusJson(resDirPath)

	fmt.Println("StartupATask(), ready exec the exe program !")
	rendererExeName := "./renderer.exe"
	cmdParams := getCmdParamsString(rendererExeName, taskID, times, resDirPath)

	fmt.Println("StartupATask(), exe cmdParams: ", cmdParams)

	cmd := exec.Command("cmd.exe", "/c", "start "+cmdParams)
	cmd.Run()
}

type RenderingSTChannelData struct {
	pathDir  string
	taskName string
	stType   int
	flag     int
	uid      int64
	index    int
}

var stRenderingCh chan RenderingSTChannelData

type TaskExecNode struct {
	uid           int64
	index         int
	desc          string
	runningStatus int
	rstData       RenderingSTChannelData

	pathDir  string
	taskName string
	filePath string
	taskID   int64
	times    int64
	progress int64
}

func (self *TaskExecNode) Init() *TaskExecNode {
	self.uid = 0
	self.index = 0
	self.desc = "a TaskExecNode instance."
	self.pathDir = ""
	self.taskName = ""
	self.filePath = "renderingStatus.json"
	self.runningStatus = 0
	self.taskID = 1
	self.times = 1
	self.progress = 0
	return self
}
func (self *TaskExecNode) Reset() *TaskExecNode {
	self.Init()
	return self
}

func (self *TaskExecNode) Exec() *TaskExecNode {
	if self.runningStatus == 1 {
		if self.pathDir == "" {
			fmt.Println("Exec(), ready startup a new task")
			resDirPath := ".\\static\\sceneres\\scene01\\"
			rendererPath := "./renderer.exe"
			if self.taskName != "" {
				resDirPath = ".\\static\\sceneres\\" + self.taskName + "\\"
			}

			self.pathDir = resDirPath
			self.filePath = self.pathDir + "renderingStatus.json"

			self.times++
			self.progress = 0
			self.runningStatus = 2
			go StartupATask(resDirPath, rendererPath, self.taskID, self.times)
		} else {
			fmt.Println("Exec(), error:  self.pathDir is not empty.")
		}
	}
	return self
}
func (self *TaskExecNode) CheckRendering() *TaskExecNode {
	if self.runningStatus == 2 {
		if self.pathDir != "" {
			fmt.Println("CheckRendering(), task checking ...")
			hasStatusFile := HasSceneResStatusJson(self.pathDir)
			fmt.Println("CheckRendering(), >>> filePath: ", self.filePath)
			fmt.Println("CheckRendering(), >>> hasStatusFile: ", hasStatusFile)
			if hasStatusFile {

				ins, err := readRenderingStatusJson(self.pathDir)
				if err == nil {
					// fmt.Println("CheckRendering(), ins: ", ins)
					task := ins.Rendering_task
					taskID := task.TaskID
					times := task.Times
					progress := task.Progress
					self.progress = progress
					fmt.Println("CheckRendering(), taskID: ", taskID, ", times: ", times)
					fmt.Println("CheckRendering(), ### progress: ", progress, "%")
					if taskID == self.taskID && times == self.times && progress >= 100 {
						fmt.Println("CheckRendering(), >>> rendering task process finish !!!")
						fmt.Println("CheckRendering(), >>> waiting for the next task ...")
						self.runningStatus = 0
						self.pathDir = ""
						self.taskName = ""
					}
				} else {
					fmt.Println("CheckRendering(), >>> read renderingStatusJson failed !!!")
				}
			}
		}
	}
	return self
}
func (self *TaskExecNode) isFree() bool {
	return self.runningStatus == 0
}

// func (self *TaskExecNode) isEnabled() bool {
// 	return self.runningStatus == 2
// }

func StartupTaskCheckingTicker(in <-chan RenderingSTChannelData) {

	// var nodes [8]TaskExecNode
	var execNode TaskExecNode
	execNode.uid = 1
	execNode.index = 1
	execNode.taskID = 1
	execNode.times = 1

	for range time.Tick(1 * time.Second) {
		var st RenderingSTChannelData
		// if execNode.isEnabled() {
		// 	fmt.Println("StartupTaskCheckingTicker() execNode is enable a task.")
		// 	execNode.Exec()
		// } else {
		// 	execNode.CheckRendering()
		// }

		status := execNode.runningStatus
		switch status {
		case 1:
			execNode.Exec()
		case 2:
			execNode.CheckRendering()
		default:
			status = 0
		}

		st = <-in
		// fmt.Println("StartupTaskCheckingTicker() >>> ticker st flag: ", st.flag)
		if st.flag > 0 {
			fmt.Println("StartupTaskCheckingTicker() >>> get a new task.")
			if execNode.isFree() {
				fmt.Println("StartupTaskCheckingTicker() execNode is free.")
				execNode.runningStatus = 1
				execNode.taskName = st.taskName
			}
		}
	}
}
func startupRProxyTicker(out chan<- RenderingSTChannelData) {

	for range time.Tick(1 * time.Second) {
		// fmt.Println("tick does...")
		var st RenderingSTChannelData
		st.pathDir = ""
		st.stType = 0
		st.flag = 0
		out <- st
	}
}
func StartTaskMonitor() {
	stRenderingCh = make(chan RenderingSTChannelData, 8)
	go startupRProxyTicker(stRenderingCh)
	go StartupTaskCheckingTicker(stRenderingCh)
}

func AddATaskFromTaskName(taskName string) {
	var st RenderingSTChannelData
	st.taskName = taskName
	st.stType = 1
	st.flag = 1
	stRenderingCh <- st
}

func StartSvr() {

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, fmt.Sprintf("There is Home Page."))
	})
	router.GET("/rendering", func(c *gin.Context) {
		taskName := c.DefaultQuery("taskName", "default")
		fmt.Println("xxx taskName: ", taskName)
		AddATaskFromTaskName(taskName)
		c.String(http.StatusOK, fmt.Sprintf("This task is currently executing now."))
	})
	router.Run(":9092")
}

func main() {
	fmt.Println("renderingmsvr init ...")

	StartTaskMonitor()
	StartSvr()
	fmt.Println("renderingmsvr end ...")
}
