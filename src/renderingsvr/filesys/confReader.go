package filesys

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var SysConfMap = make(map[string]string)

func decompressFile(compressFilePath string, dstDir string) {
	fmt.Println("decompressFile() init ...")
	archive, err := zip.OpenReader(compressFilePath)
	if err != nil {
		// panic(err)
		fmt.Printf("decompressFile(), decompress error:%v", err)
	}
	defer archive.Close()
	for _, f := range archive.File {
		filePath := filepath.Join(dstDir, f.Name)
		fmt.Println("decompressFile(), unzipping file ", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(dstDir)+string(os.PathSeparator)) {
			fmt.Println("decompressFile(), invalid file path")
			return
		}
		if f.FileInfo().IsDir() {
			fmt.Println("decompressFile(), creating dir...")
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			panic(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}
}

func syncRProcRes(param *SysStartupParam) {

	fmt.Println("syncRProcRes() init ...")
	srcDir := param.SvrRootUrl + "static/dsrdiffusion/sys/package/"
	dstDir := "static/sys/bpyc/"
	envSrcDir := param.SvrRootUrl + "static/dsrdiffusion/common/env/"
	envDstDir := "static/common/env/"

	var srcDirs = [3]string{srcDir, srcDir, envSrcDir}
	var dstDirs = [3]string{dstDir, dstDir, envDstDir}
	var files = [3]string{"render.zip", "model.zip", "default.hdr"}
	var flags = [3]bool{true, true, false}

	flagValue := 0
	filesTotal := len(files)
	for i := 0; i < filesTotal; i++ {
		url := srcDirs[i] + files[i]
		// fmt.Println("syncRProcRes() init url: ", url)
		loaderChannel := make(chan int, 1)
		go DownloadFile(loaderChannel, dstDirs[i], url, 0, "")

		for flag := range loaderChannel {
			len := len(loaderChannel)
			if len == 0 {
				flagValue += flag
				fmt.Println("syncRProcRes(), loaded flag: ", flag, ", url: ", url)
				fmt.Println("syncRProcRes(), dstDir: ", dstDirs[i])
				close(loaderChannel)
				if flag == 1 && flags[i] {
					// decompress
					filePath := dstDirs[i] + files[i]
					// fmt.Println("ready compress filePath: ", filePath)
					decompressFile(filePath, dstDirs[i])
				}
			}
		}
	}
	if flagValue == filesTotal {
		fmt.Println("syncRProcRes() success ...")
		// write rcfg.json file
		rcfg := &LocalSysCfg
		render := &rcfg.Renderer
		render.MainProc = "python " + dstDir + "render/renderShell.py"
		render.RenerderProc = rendererCmdParam + " rmodule=" + dstDir + "render/modelRendering.py"

		mtd := &rcfg.ModelToDrc
		mtd.MainProc = "python " + dstDir + "model/encodeAModelToDrcs.py -- encoder=" + dstDir + "model/draco_encoder.exe"
		mtd.ExportProc = rendererCmdParam + " exportPy=" + dstDir + "model/exportMeshesToDrcObjs.py"

		WriteTxtFileToPath(rcfgFilePath, rcfg.GetJsonString())
	}
	fmt.Println("syncRProcRes() end ...")
}
func GetSysConfValueWithName(keyName string) string {
	value, hasKey := SysConfMap[keyName]
	if hasKey {
		return value
	}
	return ""
}
func ReadSysConfFile() {
	filePath := "./conf.md"
	hasFilePath, _ := PathExists(filePath)
	if hasFilePath {
		confFile, err := os.OpenFile(filePath, os.O_RDONLY, os.ModeDevice)
		if err == nil {
			defer confFile.Close()
			confValue, err := ioutil.ReadAll(confFile)
			if err == nil {
				fileContent := string(confValue)
				// fmt.Println("ReadConfFile(), fileContent: ", fileContent)
				lines := strings.Split(strings.ReplaceAll(fileContent, "\r\n", "\n"), "\n")
				for _, value := range lines {
					value = strings.TrimSpace(value)
					index := strings.Index(value, "#")
					if index != 0 {
						index = strings.Index(value, "=")
						if index > 0 && index < (len(value)-1) {
							key := value[0:index]
							if key != "" {

								value = value[index+1:]
								key = strings.TrimSpace(key)
								value = strings.TrimSpace(value)

								// fmt.Println("ReadConfFile(), key, value: ", key+","+value)
								SysConfMap[key] = value
								switch key {
								case "cmdparams":
									args := strings.Split(value, " ")
									argsLen := len(args)
									for i := 0; i < argsLen; i++ {
										if strings.Index(args[i], "=") > 0 {
											parts := strings.Split(args[i], "=")
											if len(parts) > 1 {
												SysConfMap[parts[0]] = parts[1]
												// fmt.Println("		cmdparams key: ", parts[0], ", value: ", parts[1])
											}
										}
									}
								default:
								}
							}
						}
					}
				}
			}
		}
	}
}
