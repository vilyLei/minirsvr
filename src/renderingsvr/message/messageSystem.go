package message

import (
	"fmt"
)

// go mod init renderingsvr.com/message

type RenderingSTChannelData struct {
	PathDir  string
	TaskName string
	ResUrl   string
	StType   int
	Flag     int
}

var STRenderingCh chan RenderingSTChannelData

func Init() {
	fmt.Println("message module Init() ...")
	STRenderingCh = make(chan RenderingSTChannelData, 8)
}
