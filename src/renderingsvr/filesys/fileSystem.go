package filesys

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

)

// go mod init renderingsvr.com/filesys

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
type RenderingConfigParam struct {
	Uuid       string
	TaskID     int64
	Name       string
	OutputPath string
	Times      int64
	Progress   int

	ResourceType string
	Models       string
}
type RenderingTask struct {
	Uuid     string `json:"uuid"`
	TaskID   int64  `json:"taskID"`
	Name     string `json:"name"`
	Phase    string `json:"phase"`
	Times    int64  `json:"times"`
	Progress int    `json:"progress"`
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
func ReadRenderingStatusJson(pathDir string) (RenderingIns, error) {
	pathStr := pathDir + "renderingStatus.json"
	jsonFile, err := os.OpenFile(pathStr, os.O_RDONLY, os.ModeDevice)
	var rIns RenderingIns
	if err == nil {
		defer jsonFile.Close()
		fi, _ := jsonFile.Stat()
		fileBytesTotal := int(fi.Size())
		fmt.Println("fileBytesTotal: ", fileBytesTotal)

		jsonValue, _ := ioutil.ReadAll(jsonFile)

		err = json.Unmarshal([]byte(jsonValue), &rIns)
		if err != nil {
			fmt.Printf("readRenderingStatusJson() Unmarshal failed, err: %v\n", err)
		}
		// fmt.Println("readRenderingStatusJson(), rIns.Rendering_ins: ", rIns.Rendering_ins)
		fmt.Println("readRenderingStatusJson(), rIns.Rendering_task: ", rIns.Rendering_task)
	} else {
		fmt.Printf("readRenderingStatusJson() failed, err: %v\n", err)
	}
	return rIns, err
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
func CreateRenderingConfigFileToPath(path string, rendererPath string, param RenderingConfigParam) {

	fileContent := `{
		"renderer-proc":"` + rendererPath + `",
		"renderer-instance":
			{
				"name":"high-image-renderer",
				"status":"stop"
			},
		"resource":
			{
				"type": "` + param.ResourceType + `",
				"models": ` + param.Models + `
			},
		"task":
			{
				"taskID": ` + strconv.FormatInt(param.TaskID, 10) + `,
				"times": ` + strconv.FormatInt(param.Times, 10) + `,
				"outputPath": "` + param.OutputPath + `"
			}
		}`
	filePath := path + "config.json"
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("CreateRenderingConfigFileToPath(), err: %v\n", err)
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
	fmt.Println("CreateRenderingConfigFileToPath(), success !!!")
}
