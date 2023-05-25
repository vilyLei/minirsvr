package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
)

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
	},

"rendering-status":"task:running"
}
*/
type RenderingTask struct {
	Uuid   string `json:"uuid"`
	TaskID int64  `json:"taskID"`
	Name   string `json:"name"`
	Phase  string `json:"phase"`
	Times  string `json:"times"`
}
type RenderingIns struct {
	Rendering_ins    string `json:"rendering-ins"`
	Rendering_task   RenderingTask
	Rendering_status string `json:"rendering-status"`
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
func readRenderingStatusJson() {
	pathStr := "./renderingStatus.json"
	jsonFile, err := os.OpenFile(pathStr, os.O_RDONLY, os.ModeDevice)
	if err == nil {
		defer jsonFile.Close()
		fi, _ := jsonFile.Stat()
		fileBytesTotal := int(fi.Size())
		fmt.Println("fileBytesTotal: ", fileBytesTotal)

		jsonValue, _ := ioutil.ReadAll(jsonFile)

		var rIns RenderingIns
		json.Unmarshal([]byte(jsonValue), &rIns)
		fmt.Println("readRenderingStatusJson(), rIns.Rendering_ins: ", rIns.Rendering_ins)
		// models := modelInfo.Models
		// modelsTotal := len(models)
		// fmt.Println("modelsTotal: ", modelsTotal)
		// for i := 0; i < modelsTotal; i++ {
		// 	model := models[i]
		// 	fmt.Println("model.Url: ", model.Url)
		// }
	}
}
func getCmdParamsString(taskID int64, renderingTimes int64, paths ...string) string {
	// taskID := 1003
	// renderingTimes := 11

	path := ".\\static\\sceneres\\scene01\\"
	if len(paths) > 0 {
		path = paths[0]
	}

	taskIDStr := strconv.FormatInt(int64(taskID), 10)
	renderingTimesStr := strconv.FormatInt(int64(renderingTimes), 10)
	// cmdParams := "./exeForGo.exe .\\static\\sceneres\\scene01\\ " + taskIDStr + " " + renderingTimesStr
	cmdParams := "./exeForGo.exe " + path + " " + taskIDStr + " " + renderingTimesStr
	return cmdParams
}
func HasSceneResDir(resDirPath string) bool {
	fmt.Println("\nHasSceneResDir(), resDirPath: ", resDirPath)
	hasResDirPath, _ := PathExists(resDirPath)

	fmt.Println("HasSceneResDir(), hasResDirPath: ", hasResDirPath)
	return hasResDirPath
}
func HasSceneResStatusJson(resDirPath string) bool {
	fmt.Println("\nHasSceneResStatusJson(), resDirPath: ", resDirPath)
	hasResDirPath, _ := PathExists(resDirPath)

	fmt.Println("HasSceneResStatusJson(), hasResDirPath: ", hasResDirPath)
	if hasResDirPath {
		filePath := resDirPath + "renderingStatus.json"
		hasFilePath, _ := PathExists(filePath)
		fmt.Println("HasSceneResStatusJson(), hasFilePath: ", hasFilePath)
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
	// str := "http://c.biancheng.net/golang/\n" // \n\r表示换行  txt文件要看到换行效果要用 \r\n
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
func main() {
	fmt.Println("go exe caller Hello World!")

	// resDirPath := ".\\static\\sceneres\\scene01\\"
	resDirPath := ".\\static\\sceneres\\scene02\\"
	rendererPath := "./proc02.exe"
	fmt.Println("resDirPath: ", resDirPath)

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

	readRenderingStatusJson()

	fmt.Println("ready exec the exe program !")
	cmdParams := getCmdParamsString(1006, 16, resDirPath)

	fmt.Println("exe cmdParams: ", cmdParams)

	cmd := exec.Command("cmd.exe", "/c", "start "+cmdParams)
	cmd.Run()
}
