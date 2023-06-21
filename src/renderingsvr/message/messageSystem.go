package message

import (
	"fmt"
)

// go mod init renderingsvr.com/message

type RenderingSTChannelData struct {
	PathDir    string
	TaskName   string
	TaskAction string
	ResUrl     string
	RootDir    string
	StType     int
	Flag       int
	TaskID     int64
	Resolution [2]int
}

var STRenderingCh chan RenderingSTChannelData

func Init() {
	fmt.Println("message module Init() ...")
	STRenderingCh = make(chan RenderingSTChannelData, 8)
}
