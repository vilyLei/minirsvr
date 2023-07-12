package task

import (
	"fmt"
	"os"
	"path/filepath"

	"renderingsvr.com/filesys"
	"renderingsvr.com/message"
	"renderingsvr.com/rdata"
)

// go mod init renderingsvr.com/task

type TaskTransformNode struct {
	model [9]float64
}

type TaskExecNode struct {
	Uid           int64
	Index         int
	Desc          string
	Phase         string
	Action        string
	RunningStatus int
	RstData       message.RenderingSTChannelData

	PathDir          string
	TaskName         string
	ResUrl           string
	FilePath         string
	TaskID           int64
	Times            int64
	ReqProgress      int
	Progress         int
	TaskOutput       TaskOutputParam
	RootDir          string
	ModelExportDrcST int
	ResFilePath      string

	RNode rdata.RTRenderingNode `json:"rnode"`
}

func (self *TaskExecNode) Init() *TaskExecNode {
	self.Uid = 0
	self.Index = 0
	self.Desc = "a TaskExecNode instance."
	self.PathDir = ""
	self.TaskName = ""
	self.ResUrl = ""
	self.FilePath = "renderingStatus.json"
	self.RootDir = ""
	self.RunningStatus = 0
	self.TaskID = 1
	self.Phase = "running"
	self.Action = "new"
	self.Times = 1
	self.ReqProgress = 0
	self.Progress = 0
	self.TaskOutput.PicPath = ""
	self.TaskOutput.Error = false
	self.ModelExportDrcST = -1
	self.ResFilePath = ""
	return self
}
func (self *TaskExecNode) Reset() *TaskExecNode {
	self.Init()
	return self
}
func (self *TaskExecNode) CheckTaskStatus() (bool, int) {

	if self.PathDir != "" {
		fmt.Println("CheckTaskStatus(), task checking ...")
		hasStatusFile := filesys.HasSceneResStatusJson(self.PathDir)
		fmt.Println("CheckTaskStatus(), >>> filePath: ", self.FilePath)
		fmt.Println("CheckTaskStatus(), >>> hasStatusFile: ", hasStatusFile)

		if hasStatusFile {

			ins, err := filesys.ReadRenderingStatusJson(self.PathDir)
			if err == nil {
				// fmt.Println("CheckTaskStatus(), ins: ", ins)
				task := ins.Rendering_task
				taskID := task.TaskID
				times := task.Times
				progress := task.Progress
				phase := task.Phase
				self.Progress = progress

				fmt.Println("CheckTaskStatus(), taskID: ", taskID, ", times: ", times)
				fmt.Println("CheckTaskStatus(), ### progress: ", progress, "%", "taskID,self.TaskID, times == self.Times: ", taskID, self.TaskID, times == self.Times)
				taskFlag := false
				if taskID == self.TaskID && times == self.Times {
					if progress >= 100 {

						fmt.Println("CheckTaskStatus(), >>> self.PathDir: ", self.PathDir)
						fmt.Println("CheckTaskStatus(), >>> phase: ", phase)

						taskStatus := 1
						if phase == "error" {
							taskStatus = -1
							fmt.Println("CheckTaskStatus(), >>> rendering task process has a error !!!")
						} else {
							// check output pic file
							pic_type := ""
							if self.IsBGTransparentOutput() {
								pic_type = "png"
							}
							taskFlag, _ = filesys.CheckPicFileInCurrDir(self.PathDir, pic_type)
							fmt.Println("CheckTaskStatus(), >>> has output rendering pic: ", taskFlag)
							if !taskFlag {
								taskStatus = -1
							}
						}
						return taskFlag, taskStatus
					}
					return taskFlag, 5
				}
				return taskFlag, 6
			} else {
				fmt.Println("CheckTaskStatus(), >>> read renderingStatusJson failed !!!")
			}
		}
	}
	return false, 0
}
func (self *TaskExecNode) CheckModelDrcStatus() int {

	drcDir := filepath.Dir(self.ResFilePath) + "/draco/"
	hasFilePath, _ := filesys.PathExists(drcDir + "status.json")
	if hasFilePath {
		return 1
	}
	hasFilePath, _ = filesys.PathExists(drcDir)
	if hasFilePath {
		return 2
	}
	fmt.Println("CheckModelDrcStatus(), >>> drcDir: ", drcDir)
	return 0
}
func (self *TaskExecNode) IsBGTransparentOutput() bool {
	output := self.RNode.Output
	return output.BGTransparent == 1
}
func (self *TaskExecNode) toTaskFinish() *TaskExecNode {

	fmt.Println("toTaskFinish(), >>> pathDif: ", self.PathDir)
	pic_type := ""
	if self.IsBGTransparentOutput() {
		pic_type = "png"
	}
	picFlag, picNames := filesys.CheckPicFileInCurrDir(self.PathDir, pic_type)
	self.TaskOutput.PicPath = ""
	if picFlag {
		// ready send the pic to a remote data center server
		picFilePath := self.PathDir + picNames[0]
		fmt.Println("toTaskFinish(), >>> picFilePath: ", picFilePath)
		self.TaskOutput.PicPath = picFilePath
	}
	self.TaskOutput.Error = false
	self.TaskOutput.TaskID = self.TaskID
	self.TaskOutput.TaskName = self.TaskName

	self.RunningStatus = 0
	self.PathDir = ""
	self.TaskName = ""
	fmt.Println("toTaskFinish(), >>> waiting for the next task ...")
	return self
}
func (self *TaskExecNode) toTaskError() *TaskExecNode {

	fmt.Println("toTaskError(), >>> pathDif: ", self.PathDir)

	self.TaskOutput.Error = true
	self.TaskOutput.PicPath = ""
	self.TaskOutput.TaskID = self.TaskID
	self.TaskOutput.TaskName = self.TaskName

	self.RunningStatus = 0
	self.PathDir = ""
	self.TaskName = ""
	fmt.Println("toTaskError(), >>> waiting for the next task ...")
	return self
}
func (self *TaskExecNode) Exec() *TaskExecNode {
	if self.RunningStatus == 1 {
		if self.PathDir == "" {
			fmt.Println("Exec(), ready startup a new task")

			rendererPath := "./renderer.exe"

			if self.TaskName == "" {
				self.TaskName = "scene01"
			}
			// resDirPath := "./static/sceneres/" + self.TaskName + "/"
			rootDir, err := os.Getwd()
			if err != nil {
				fmt.Println("os.Getwd(), err: %v", rootDir)
				rootDir = "."
			}
			fmt.Println("rootDir: ", rootDir)
			resDirPath := rootDir + "/static/sceneres/" + self.TaskName + "/"

			self.PathDir = resDirPath
			self.FilePath = self.PathDir + "renderingStatus.json"
			// if the action is rerendering proc, remove the renderingStatus.json file
			if self.Action == "query-re-rendering-task" {
				// todo: chech the task has finished
				filePath := resDirPath + "renderingStatus.json"
				flag := filesys.RemoveFileWithPath(filePath)
				fmt.Println("Exec(), query-re-rendering-task clear the rtask status info file, flag: ", flag, filePath)
			}

			self.Times++
			self.Progress = 0
			self.RunningStatus = 2
			_, taskStatus := self.CheckTaskStatus()
			if taskStatus == 1 {
				taskStatus = -1
			}
			if taskStatus == 1 {
				// 没价值的逻辑
				fmt.Println("Exec(), the task output result is directly available !!!")
				self.toTaskFinish()
			} else {
				if taskStatus != 0 {
					// clear the status info file
					filePath := resDirPath + "renderingStatus.json"
					flag := filesys.RemoveFileWithPath(filePath)
					// filePath = resDirPath + "draco/status.json"
					dirPath := resDirPath + "draco/"
					flag = filesys.RemoveDirAndFiles(dirPath)
					fmt.Println("Exec(), clear the rtask status info file, flag: ", flag, filePath)
				}
				self.ResFilePath = filesys.GetModelFilePath(resDirPath, self.ResUrl)

				if self.CheckModelDrcStatus() == 0 {
					self.ModelExportDrcST = 0
				}

				self.TaskOutput.TaskID = self.TaskID
				self.TaskOutput.TaskName = self.TaskName

				// go StartupATask(self.RootDir, resDirPath, rendererPath, self.TaskID, self.Times, self.TaskName, self.ResUrl)
				go StartupATask(self.RootDir, resDirPath, rendererPath, *self)
			}
		} else {
			fmt.Println("Exec(), error:  self.pathDir is not empty.")
		}
	}
	return self
}
func (self *TaskExecNode) CheckRendering() *TaskExecNode {
	if self.RunningStatus == 2 {
		flag, status := self.CheckTaskStatus()

		fmt.Println("CheckRendering(), >>> flag, status: ", flag, status)
		if flag {
			if status == 1 {
				fmt.Println("CheckRendering(), >>> rendering task process finish !!!")
				self.toTaskFinish()
			} else {
				fmt.Println("CheckRendering(), >>> rendering task process failed A !!!")
				self.toTaskError()
			}
		} else {
			if status == -1 {
				fmt.Println("CheckRendering(), >>> rendering task process failed B !!!")
				self.toTaskError()
			}
		}
	}
	return self
}

func (self *TaskExecNode) IsWaitingTask() bool {
	return self.RunningStatus == 5
}
func (self *TaskExecNode) IsFree() bool {
	return self.RunningStatus == 0
}
