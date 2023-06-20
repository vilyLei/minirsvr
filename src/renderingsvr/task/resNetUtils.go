package task

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func GetFileNameFromUrl(url string) string {
	nameStr := url[strings.LastIndex(url, "/")+1 : len(url)]
	i := strings.LastIndex(nameStr, "?")
	if i > 0 {
		nameStr = nameStr[0:i]
	}
	return nameStr
}
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

	// fmt.Println("#### >> ## resp.ContentLength: ", resp.ContentLength, ", len(data): ", len(data))
	if resp.ContentLength < 300 {
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
