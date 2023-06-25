package svr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"renderingsvr.com/message"
	"renderingsvr.com/task"
)

// go mod init renderingsvr.com/svr

func ReadyAddANewTask(taskName string) {
	var st message.RenderingSTChannelData
	st.TaskName = taskName
	st.StType = 1
	st.Flag = 2
	message.STRenderingCh <- st
}

var taskMap map[string]*message.RenderingSTChannelData

func startupRProxyTicker(out chan<- message.RenderingSTChannelData) {

	for range time.Tick(1 * time.Second) {
		// fmt.Println("tick does...")
		var st message.RenderingSTChannelData
		st.PathDir = ""
		st.StType = 0
		st.Flag = 0
		out <- st
	}
}
func startupAutoCheckTaskTicker() {

	for range time.Tick(2 * time.Second) {
		ReadyAddANewTask("random-task")
	}
}
func checkTaskOutput(op *task.TaskOutputParam) {
	// op := &execNode.TaskOutput
	if op.Error {
		fmt.Println("StartupTaskCheckingTicker() >>> upload process failed !!!")
		NotifyTaskInfoToSvr("rtaskerror", 100, op.TaskID, op.TaskName)
		op.Error = false
	} else if op.PicPath != "" {
		// upload the rendering output pic to remote data svr
		fmt.Println("StartupTaskCheckingTicker() >>> upload the rendering output pic to remote data svr.")
		uploadErr := postFileToResSvr(op.PicPath, uploadSvrUrl, "finish", op.TaskID, op.TaskName)
		if uploadErr == nil {
			fmt.Println("StartupTaskCheckingTicker() >>> upload process success !!!")
			// notify task finish into to the server
			NotifyTaskInfoToSvr("finish", 100, op.TaskID, op.TaskName)
		} else {
			fmt.Println("StartupTaskCheckingTicker() >>> upload process failed !!!")
		}
		op.PicPath = ""
	}
}
func StartupTaskCheckingTicker(in <-chan message.RenderingSTChannelData) {

	// var nodes [8]TaskExecNode
	var execNode task.TaskExecNode
	execNode.Uid = 1
	execNode.Index = 1
	execNode.TaskID = 1
	execNode.Times = 1

	// for range time.Tick(time.Second) {
	for range time.Tick(500 * time.Millisecond) {

		var st message.RenderingSTChannelData
		status := execNode.RunningStatus

		switch status {
		case 1:
			execNode.Exec()
		case 2:
			execNode.CheckRendering()
			if execNode.ReqProgress != execNode.Progress {
				execNode.ReqProgress = execNode.Progress
				fmt.Println("StartupTaskCheckingTicker() >>> A execNode.ReqProgress: ", execNode.ReqProgress, "%")
				NotifyTaskInfoToSvr("running", execNode.Progress, execNode.TaskID, execNode.TaskName)
				if execNode.IsFree() {
					checkTaskOutput(&execNode.TaskOutput)
				}
			}
		default:
			status = 0
			// if execNode.IsFree() {
			// 	checkTaskOutput(&execNode.TaskOutput)
			// }
			// op := &execNode.TaskOutput
			// if op.Error {
			// 	fmt.Println("StartupTaskCheckingTicker() >>> upload process failed !!!")
			// 	NotifyTaskInfoToSvr("rtaskerror", 100, op.TaskID, op.TaskName)
			// 	op.Error = false
			// } else if op.PicPath != "" {
			// 	// upload the rendering output pic to remote data svr
			// 	fmt.Println("StartupTaskCheckingTicker() >>> upload the rendering output pic to remote data svr.")
			// 	uploadErr := postFileToResSvr(op.PicPath, uploadSvrUrl, "finish", op.TaskID, op.TaskName)
			// 	if uploadErr == nil {
			// 		fmt.Println("StartupTaskCheckingTicker() >>> upload process success !!!")
			// 		// notify task finish into to the server
			// 		NotifyTaskInfoToSvr("finish", 100, op.TaskID, op.TaskName)
			// 	} else {
			// 		fmt.Println("StartupTaskCheckingTicker() >>> upload process failed !!!")
			// 	}
			// 	op.PicPath = ""
			// }
		}

		st = <-in
		// fmt.Println("StartupTaskCheckingTicker() >>> ticker st.Flag : ", st.Flag)
		if st.Flag > 0 {
			switch st.Flag {
			case 1:
				fmt.Println("StartupTaskCheckingTicker() >>> get a new task.")
				fmt.Println("StartupTaskCheckingTicker() >>> execNode.IsWaitingTask(): ", execNode.IsWaitingTask(), st.TaskName)
				fmt.Println("StartupTaskCheckingTicker() >>> execNode.RunningStatus: ", execNode.RunningStatus)
				if execNode.IsFree() {
					execNode.RunningStatus = 5
				}
				if execNode.IsWaitingTask() {
					execNode.RunningStatus = 1
					execNode.TaskName = st.TaskName
					execNode.TaskID = st.TaskID
					execNode.ResUrl = st.ResUrl
					execNode.RootDir = st.RootDir
					execNode.Action = st.TaskAction
					execNode.Resolution = st.Resolution
					execNode.Camdvs = st.Camdvs
					fmt.Println("	$$$->>> execNode.TaskID: ", execNode.TaskID)
					fmt.Println("	$$$->>> execNode.Resolution: ", execNode.Resolution)
					fmt.Println("	$$$->>> execNode.Camdvs: ", execNode.Camdvs)
					fmt.Println("	$$$->>> execNode.TaskName: ", execNode.TaskName)
					fmt.Println("	$$$->>> execNode.ResUrl: ", execNode.ResUrl)
				}
			case 2:
				fmt.Println("StartupTaskCheckingTicker() >>> ready add a new task, execNode.IsFree(): ", execNode.IsFree(), st.TaskName)
				if execNode.IsFree() {
					execNode.RunningStatus = 5
					go StartupATaskReq()
				}
			case 11:
				fmt.Println("StartupTaskCheckingTicker() >>> nothing a new task.")
				if execNode.RunningStatus == 5 {
					execNode.RunningStatus = 0
				}
			default:
				st.Flag = 0
			}
		}
	}
}
func StartTaskMonitor() {
	if AutoCheckRTask {
		go startupAutoCheckTaskTicker()
	}
	// if AutoCheckRTask {
	// 	startupAutoReqNewTaskTicker()
	// }
	go startupRProxyTicker(message.STRenderingCh)
	go StartupTaskCheckingTicker(message.STRenderingCh)
}

func HasTaskByName(ns string) bool {
	_, hasKey := taskMap[ns]
	return hasKey
}

func AddANewTaskFromTaskInfo(tasks []RTaskJsonNode) {

	total := len(tasks)
	nothingFlag := true
	if total > 0 {
		var tk *RTaskJsonNode = nil
		for i := 0; i < total; i++ {

			tk = &tasks[i]
			flag := HasTaskByName(tk.Name)
			if flag {
				if tk.Action == "query-re-rendering-task" {
					fmt.Println("AddANewTaskFromTaskInfo() >>> have a re-rendering task:", tk)
					flag = false
				}
			}
			if !flag {
				fmt.Println("AddANewTaskFromTaskInfo() >>> got a new task:", tk)
				var st message.RenderingSTChannelData
				st.TaskID = tk.Id
				st.TaskName = tk.Name
				st.TaskAction = tk.Action
				st.ResUrl = tk.ResUrl
				st.Resolution = tk.Resolution
				st.Camdvs = tk.Camdvs
				st.RootDir = RootDir
				st.StType = 1
				st.Flag = 1
				taskMap[tk.Name] = &st
				message.STRenderingCh <- st
				nothingFlag = false
				break
			}
		}
	}
	if nothingFlag {
		fmt.Println("AddANewTaskFromTaskInfo() >>> nothing new test rendering task !!!!!!!")
		var st message.RenderingSTChannelData
		// st.TaskID = 0
		// st.TaskName = ""
		// st.ResUrl = ""
		// st.StType = 0
		st.Reset()
		st.Flag = 11
		message.STRenderingCh <- st
	}
}

func StartSvr(portStr string) {

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, fmt.Sprintf("There is Home Page."))
	})
	router.GET("/rendering", func(c *gin.Context) {
		taskName := c.DefaultQuery("taskName", "default")
		taskType := c.DefaultQuery("taskType", "none")
		fmt.Println("xxx taskName: ", taskName)
		fmt.Println("xxx taskType: ", taskType)
		switch taskType {
		case "new":
			fmt.Println("xxx ready a new rendering task")
			RequestANewTask()
		default:
			ReadyAddANewTask(taskName)
		}
		c.String(http.StatusOK, fmt.Sprintf("This task is currently executing now."))
	})
	router.Run(":" + portStr)
}

func receiveTasksReq(data []byte) {

	var taskInfo RTasksJson
	err := json.Unmarshal(data, &taskInfo)
	if err != nil {
		fmt.Printf("receiveTaskReq() Unmarshal failed, err: %v\n", err)
	} else {
		tasks := taskInfo.Tasks
		total := len(tasks)
		fmt.Println("receiveTaskReq(), tasks total: ", total)
		// fmt.Println("receiveTaskReq(), taskInfo: ", taskInfo)
		if total > 0 {
			fmt.Println("receiveTaskReq(), request some new tasks.")
			// task := tasks[0]
			// pageSTNodeMap[node.name] = &node
			AddANewTaskFromTaskInfo(taskInfo.Tasks)
		}
	}
}

func receiveATaskReq(data []byte) {

	fmt.Println("receiveATaskReq(), string(data): ", string(data))
	var taskInfo RTaskJson
	err := json.Unmarshal(data, &taskInfo)
	if err != nil {
		fmt.Printf("receiveTaskReq() Unmarshal failed, err: %v\n", err)
	} else {
		task := taskInfo.Task
		if task.Id > 0 {
			fmt.Println("receiveATaskReq(), task: ", task)
			if !(strings.Contains(task.ResUrl, "https://") || strings.Contains(task.ResUrl, "http://")) {
				// if strings.Contains(task.ResUrl, "./") {
				if strings.Index(task.ResUrl, "./") == 0 {
					task.ResUrl = task.ResUrl[2:]
				}
				task.ResUrl = svrRootUrl + task.ResUrl
			}
			fmt.Println("receiveATaskReq(), task.ResUrl: ", task.ResUrl)
			// var tasks := []RTaskJsonNode{task}
			// ReadyAddANewTask("anewtask")
			AddANewTaskFromTaskInfo([]RTaskJsonNode{task})
		} else {
			fmt.Println("receiveATaskReq(), has not a new task.")
			AddANewTaskFromTaskInfo([]RTaskJsonNode{})
		}
	}
}
func NotifyTaskInfoToSvr(phase string, progress int, taskId int64, taskName string) {
	progressStr := strconv.Itoa(progress)
	taskIdStr := strconv.FormatInt(taskId, 10)
	url := taskReqSvrUrl + "?srcType=renderer&phase=" + phase + "&progress=" + progressStr
	if taskId > 0 {
		url += "&taskid=" + taskIdStr + "&taskname=" + taskName
	}
	resp, err := http.Get(url)
	flag := true
	if err != nil {
		flag = false
		fmt.Printf("NotifyTaskInfoToSvr() get url failed, err: %v\n", err)
		if phase == "reqanewrtask" {
			AddANewTaskFromTaskInfo([]RTaskJsonNode{})
		}

	} else {
		defer resp.Body.Close()
	}
	if flag {
		data, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			switch phase {
			case "running":
				fmt.Println("NotifyTaskInfoToSvr() receive running req info, ", string(data))
			case "finish":
				fmt.Println("NotifyTaskInfoToSvr() receive finish req info, ", string(data))
			case "re-rendering":
				fmt.Println("NotifyTaskInfoToSvr() re-rendering finish req info, ", string(data))
			case "rtaskerror":
				fmt.Println("NotifyTaskInfoToSvr() receive rendering task error req info, ", string(data))
			case "reqanewrtask":
				receiveATaskReq(data)
			default:
				receiveTasksReq(data)
			}
		}
	}
}
func StartupATaskReq() {

	fmt.Println("### startup a task req ...")
	if AutoCheckRTask {
		NotifyTaskInfoToSvr("reqanewrtask", 0, 0, "")
	} else {
		NotifyTaskInfoToSvr("taskreq", 0, 0, "")
	}
}
func RequestANewTask() {
	fmt.Println("### RequestANewTask() ...")
	NotifyTaskInfoToSvr("reqanewrtask", 0, 0, "")
}

var uploadSvrUrl string = "http://localhost:9090/uploadRTData"
var taskReqSvrUrl string = "http://localhost:9090/renderingTask"
var svrRootUrl string = "http://localhost:9090/"

func Init(portStr string) {
	fmt.Println("svrInstance init ...")
	taskMap = make(map[string]*message.RenderingSTChannelData)

	svrRootUrl = "http://localhost:9090/"
	svrRootUrl = "http://localhost:9091/"
	// svrRootUrl = "http://www.artvily.com:9093/"
	uploadSvrUrl = svrRootUrl + "uploadRTData"
	taskReqSvrUrl = svrRootUrl + "renderingTask"

	task.TaskReqSvrUrl = taskReqSvrUrl

	StartTaskMonitor()
	StartSvr(portStr)
	fmt.Println("svrInstance end ...")
}
