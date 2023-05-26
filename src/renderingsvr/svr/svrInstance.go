package svr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"renderingsvr.com/message"
	"renderingsvr.com/task"
)

// go mod init renderingsvr.com/svr

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

func StartupTaskCheckingTicker(in <-chan message.RenderingSTChannelData) {

	// var nodes [8]TaskExecNode
	var execNode task.TaskExecNode
	execNode.Uid = 1
	execNode.Index = 1
	execNode.TaskID = 1
	execNode.Times = 1

	for range time.Tick(1 * time.Second) {

		var st message.RenderingSTChannelData
		status := execNode.RunningStatus

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
		if st.Flag > 0 {
			switch st.Flag {
			case 1:
				fmt.Println("StartupTaskCheckingTicker() >>> get a new task.")
				if execNode.IsWaitingTask() {
					execNode.RunningStatus = 1
					execNode.TaskName = st.TaskName
					execNode.ResUrl = st.ResUrl
					fmt.Println("	>>> execNode.TaskName: ", execNode.TaskName)
					fmt.Println("	>>> execNode.ResUrl: ", execNode.ResUrl)
				}
			case 2:
				fmt.Println("StartupTaskCheckingTicker() >>> ready add a new task.")
				if execNode.IsFree() {
					execNode.RunningStatus = 5
					go StartupATaskReq()
				}
			default:
				st.Flag = 0
			}
		}
	}
}
func StartTaskMonitor() {
	go startupRProxyTicker(message.STRenderingCh)
	go StartupTaskCheckingTicker(message.STRenderingCh)
}

func HasTaskByName(ns string) bool {
	_, hasKey := taskMap[ns]
	return hasKey
}
func AddANewTaskFromTaskInfo(taskInfo RTaskJson) {

	// taskMap[node.name] = &node
	tasks := taskInfo.Tasks
	total := len(tasks)
	var task RTaskJsonNode
	task.Name = ""
	for i := 0; i < total; i++ {

		flag := HasTaskByName(tasks[i].Name)
		if !flag {
			task = tasks[i]
			var st message.RenderingSTChannelData
			st.TaskName = task.Name
			st.ResUrl = task.ResUrl
			st.StType = 1
			st.Flag = 1
			taskMap[task.Name] = &st
			message.STRenderingCh <- st
			break
		}
	}
	if task.Name == "" {
		fmt.Println("*** nothing new test rendering task ***")
	}
}
func ReadyAddANewTask(taskName string) {
	var st message.RenderingSTChannelData
	st.TaskName = taskName
	st.StType = 1
	st.Flag = 2
	message.STRenderingCh <- st
}

func StartSvr() {

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, fmt.Sprintf("There is Home Page."))
	})
	router.GET("/rendering", func(c *gin.Context) {
		taskName := c.DefaultQuery("taskName", "default")
		fmt.Println("xxx taskName: ", taskName)
		ReadyAddANewTask(taskName)
		c.String(http.StatusOK, fmt.Sprintf("This task is currently executing now."))
	})
	router.Run(":9092")
}

/*
	{
		"tasks":[
			{
				"name":"modelTask01",
				"resUrl:"http://www.artvily.com/static/assets/obj/base.obj"
			}
		]
	}
*/
type RTaskJsonNode struct {
	Name   string `json:"name"`
	ResUrl string `json:"resUrl"`
}
type RTaskJson struct {
	Tasks []RTaskJsonNode `json:"tasks"`
}

func StartupATaskReq() {
	url := "http://www.artvily.com/renderingTask"
	// for range time.Tick(9 * time.Second) {

	resp, err := http.Get(url)
	flag := true
	if err != nil {
		flag = false
		fmt.Printf("StartupATaskReq() get url failed, err: %v\n", err)
	} else {
		defer resp.Body.Close()
	}
	if flag {
		data, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			var taskInfo RTaskJson
			err = json.Unmarshal(data, &taskInfo)
			if err != nil {
				fmt.Printf("StartupATaskReq() Unmarshal failed, err: %v\n", err)
			} else {
				tasks := taskInfo.Tasks
				total := len(tasks)
				fmt.Println("tasks total: ", total)
				// fmt.Println("taskInfo: ", taskInfo)
				if total > 0 {
					fmt.Println("request some new tasks.")
					// task := tasks[0]
					//pageSTNodeMap[node.name] = &node
					AddANewTaskFromTaskInfo(taskInfo)
				}
			}
		}
	}
	// }
}
func Init() {
	fmt.Println("svrInstance init ...")
	taskMap = make(map[string]*message.RenderingSTChannelData)

	// go StartupTaskReqSys()

	StartTaskMonitor()
	StartSvr()
	fmt.Println("svrInstance end ...")
}
