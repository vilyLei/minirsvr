package task

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

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
