package filesys

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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

var LocalSysCfg LocalSysConfig

func GetLocalSysCfg() {
	filePath := "static/sys/local/config.json"
	hasFilePath, _ := PathExists(filePath)
	if hasFilePath {
		jsonFile, err := os.OpenFile(filePath, os.O_RDONLY, os.ModeDevice)
		if err == nil {
			defer jsonFile.Close()
			jsonValue, _ := ioutil.ReadAll(jsonFile)

			err = json.Unmarshal([]byte(jsonValue), &LocalSysCfg)
			if err != nil {
				fmt.Printf("GetLocalSysCfg() Unmarshal failed, err: %v\n", err)
			}
			fmt.Println("GetLocalSysCfg(), LocalSysCfg: ", LocalSysCfg)
		}
	}
}
func GetRenderCMD(rtaskDir string) string {
	return LocalSysCfg.GetRenderCMD(rtaskDir)
}
func RemoveFileWithPath(filePath string) bool {
	err := os.Remove(filePath)
	return !(err != nil)
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

func GetAllFilesNamesInCurrDir(dir string) []string {
	// names := make([]string, 0)
	var names []string
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return names
	}
	for _, file := range files {
		if !file.IsDir() {
			// fmt.Println(file.Name())
			names = append(names, file.Name())
		}
	}
	return names
}
func GetFileNameSuffix(ns string) string {

	if strings.Contains(ns, ".") {
		parts := strings.Split(ns, ".")
		pns := parts[len(parts)-1]
		return strings.ToLower(pns)
	}
	return ""
}
func CheckPicFileInCurrDir(dir string) (bool, []string) {
	// names := make([]string, 0)
	var names []string = GetAllFilesNamesInCurrDir(dir)
	flag := false
	var picNames []string

	for _, ns := range names {
		// parts := strings.Split(ns, ".")
		// pns := parts[len(parts)-1]
		// pns = strings.ToLower(pns)
		// // fmt.Println("pns: ", pns)
		pns := GetFileNameSuffix(ns)
		switch pns {
		case "jpg", "jpeg", "png":
			picNames = append(picNames, ns)
			flag = true
		}
	}
	return flag, picNames
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

	/*
		sizes := param.Resolution
		fileContent := `{
			"renderer-proc":"` + rendererPath + `",
			"renderer-instance":
				{
					"name":"high-image-renderer",
					"status":"stop"
				},
			"sys": {
				"rootDir":"` + param.RootDir + `"
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
					"outputPath": "` + param.OutputPath + `",
					"outputResolution": [` + strconv.Itoa(sizes[0]) + `,` + strconv.Itoa(sizes[1]) + `]
				}
			}`
		//*/

	var rcfg RenderTaskConfig
	rcfg.Reset()
	rcfg.SetValueFromParam(&param)
	rcfg.RendererProc = rendererPath

	fileContent := rcfg.GetJsonString()

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
