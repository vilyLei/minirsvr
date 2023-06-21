package task

var TaskReqSvrUrl string = ""

type ResLoadParam struct {
	Url      string
	TaskName string
	PathDir  string
}
type TaskOutputParam struct {
	PicPath  string
	TaskName string
	TaskID   int64
	Error    bool
}
