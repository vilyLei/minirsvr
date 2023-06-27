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
func GetModelExportCMD(modelFilePath string) string {
	return LocalSysCfg.GetModelExportCMD(modelFilePath)
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
func CheckPicFileInCurrDir(dir string, pic_type string) (bool, []string) {
	// names := make([]string, 0)
	var names []string = GetAllFilesNamesInCurrDir(dir)
	flag := false
	var picNames []string
	if pic_type == "" {
		for _, ns := range names {
			pns := GetFileNameSuffix(ns)
			switch pns {
			case "jpg", "jpeg", "png":
				picNames = append(picNames, ns)
				flag = true
			}
		}
	} else {
		for _, ns := range names {
			pns := GetFileNameSuffix(ns)
			if pns == pic_type {
				picNames = append(picNames, ns)
				flag = true
			}
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

	var rcfg RenderTaskConfig
	rcfg.Reset()
	rcfg.SetValueFromParam(&param)
	rcfg.RendererProc = rendererPath

	fileContent := rcfg.GetJsonString()

	filePath := path + "config.json"
	hasResDirPath, _ := PathExists(filePath)
	if hasResDirPath {
		os.Remove(filePath)
	}
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("CreateRenderingConfigFileToPath(), err: %v\n", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	writer.WriteString(fileContent)

	writer.Flush()
	fmt.Println("CreateRenderingConfigFileToPath(), success !!!")
}
