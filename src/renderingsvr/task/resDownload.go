package task

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"renderingsvr.com/filesys"
)

var fileTotalBytes uint64 = 1

// "github.com/dustin/go-humanize"
// WriteCounter counts the number of bytes written to it. It implements to the io.Writer interface
// and we can pass this into io.TeeReader() which will report progress on each write cycle.
type WriteCounter struct {
	Total    uint64
	OutChan  chan<- int
	TaskID   int64
	TaskName string
	Progress int
}

func (self *WriteCounter) Reset() {
	self.Total = 0
	self.OutChan = nil
	self.Progress = 0
}
func (self *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	self.Total += uint64(n)
	self.PrintProgress()
	return n, nil
}
func (self *WriteCounter) ToFail() {

	if self.OutChan != nil {
		self.OutChan <- 0
	}
}
func (self *WriteCounter) ToSuccess() {

	if self.OutChan != nil {
		self.OutChan <- 1
	}
}
func (self *WriteCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	// fmt.Printf("\r%s", strings.Repeat(" ", 35))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	// fmt.Printf("\rDownloading... %s complete", humanize.Bytes(self.Total))
	pf := float64(self.Total) / float64(fileTotalBytes)
	// pro := int(math.Round(pf * 100.0))
	pro := int(math.Floor(pf * 100.0))
	fmt.Println("\rDownloading: ", self.Total, "bytes, pro: ", strconv.Itoa(pro)+"%")
	if self.TaskID > 0 {
		if self.Progress != pro {
			self.Progress = pro
			fmt.Println("Send >>> Downloading: ", self.Total, "self.Progress: ", strconv.Itoa(self.Progress)+"%")
			NotifyTaskInfoToSvr("task_rendering_load_res", pro, self.TaskID, self.TaskName)
		}
	}
	if pro >= 100 {
		self.ToSuccess()
		self.OutChan = nil
	}
	/*
			f := 3.14159265358979323846
		    s := strconv.FormatFloat(f, 'f', 6, 64)
		    fmt.Printf("f = %f, s = %s\\n", f, s)
	*/
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory. We pass an io.TeeReader
// into Copy() to report progress on the download.

func DownloadFile(outChan chan<- int, fileDir string, url string, taskID int64, taskName string) error {

	nameStr := GetFileNameFromUrl(url)
	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	// out, err := os.Create(fileDir + nameStr + ".tmp")
	// if err != nil {
	// 	if outChan != nil {
	// 		outChan <- 0
	// 	}
	// 	return err
	// }

	filePath := fileDir + nameStr
	hasFilePath, _ := filesys.PathExists(filePath)
	if hasFilePath {
		if outChan != nil {
			outChan <- 1
		}
		fmt.Println("The model file exist, filePath: ", filePath)
		return nil
	}

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		if outChan != nil {
			outChan <- 0
		}
		// out.Close()
		return err
	}
	defer resp.Body.Close()

	fileTotalBytes = 1
	if resp.ContentLength > 0 {
		fileTotalBytes = uint64(resp.ContentLength)
	}
	if fileTotalBytes < 300 {
		data, wErr := ioutil.ReadAll(resp.Body)
		if wErr != nil {
			fmt.Printf("write a file failed,wErr: %v\n", wErr)
			outChan <- 0
			// out.Close()
			return nil
		}
		// fmt.Println("data: ", data)
		// fmt.Println("data len: ", len(data))
		str := string(data)
		strI := strings.Index(str, "Error:")
		// fmt.Println("strI: ", strI)
		if strI > 0 {
			fmt.Println("data to str: ", str)
			outChan <- 0
			fmt.Println("load req error !!!")
			// panic("load req error")
			// out.Close()
			return nil
		} else {
			fmt.Println("remote res nameStr: ", nameStr)
			fmt.Println("remote res pathDir: ", fileDir)
			ioutil.WriteFile(filePath, data, 0777)

			fmt.Println("load a remote res file success !!!")
			outChan <- 1
			// out.Close()
			return nil
		}
	}
	fmt.Println("download fileTotalBytes: ", fileTotalBytes)

	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := os.Create(filePath + ".tmp")
	if err != nil {
		if outChan != nil {
			outChan <- 0
		}
		return err
	}

	// Create our progress reporter and pass it to be used alongside our writer
	counter := &WriteCounter{}
	counter.Reset()
	counter.OutChan = outChan
	counter.TaskID = taskID
	counter.TaskName = taskName
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		counter.ToFail()
		counter.Reset()
		out.Close()
		return err
	}

	// The progress use the same line so print a new line once it's finished downloading
	// fmt.Print("\n")

	// Close the file without defer so it can happen before Rename()
	out.Close()

	if err = os.Rename(filePath+".tmp", filePath); err != nil {
		counter.ToFail()
		counter.Reset()
		return err
	}
	return nil
}
