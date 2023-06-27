package svr

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

func uploadExample() {
	fileDir, _ := os.Getwd()
	fileName := "rendering.jpg"
	filePath := path.Join(fileDir, fileName)

	file, _ := os.Open(filePath)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
	io.Copy(part, file)
	writer.Close()

	r, _ := http.NewRequest("POST", "http://test.com/upload", body)
	r.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	client.Do(r)
}
func postFileToResSvr(filename string, svrUrl string, phase string, taskID int64, taskName string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile("file", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	// open file handle
	fh, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file")
		return err
	}
	defer fh.Close()

	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	url := svrUrl
	if taskID > 0 {
		taskIDStr := strconv.FormatInt(taskID, 10)
		url += "?srcType=renderer&phase=" + phase + "&taskid=" + taskIDStr + "&taskname=" + taskName
	}

	resp, err := http.Post(url, contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println("upload resp status: ", resp.Status)
	fmt.Println("upload resp body: ", string(resp_body))
	return nil
}

func postFilesToResSvr(filePaths []string, svrUrl string, phase string, taskID int64, taskName string) error {

	// filename := ""
	// bodyBuf := &bytes.Buffer{}
	// bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	// fileWriter, err := bodyWriter.CreateFormFile("file", filename)
	// if err != nil {
	// 	fmt.Println("error writing to buffer")
	// 	return err
	// }

	// // open file handle
	// fh, err := os.Open(filename)
	// if err != nil {
	// 	fmt.Println("error opening file")
	// 	return err
	// }
	// defer fh.Close()

	// _, err = io.Copy(fileWriter, fh)
	// if err != nil {
	// 	return err
	// }

	// contentType := bodyWriter.FormDataContentType()
	// bodyWriter.Close()

	r, w := io.Pipe()
	m := multipart.NewWriter(w)

	go func() {
		defer func() {
			// m.Close() is important so the requset knows the boundary
			m.Close()
			w.Close()
		}()
		for _, path := range filePaths {
			f, err := os.Open(path)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer f.Close()
			fmt.Println("postFilesToResSvr(), f.Name(): ", f.Name())
			if fw, err := m.CreateFormFile("files", f.Name()); err != nil {
				return
			} else {
				if _, err = io.Copy(fw, f); err != nil {
					return
				}
			}
		}
	}()

	contentType := m.FormDataContentType()

	url := svrUrl
	if taskID > 0 {
		taskIDStr := strconv.FormatInt(taskID, 10)
		url += "?srcType=renderer&phase=" + phase + "&taskid=" + taskIDStr + "&taskname=" + taskName
		url += "&total" + strconv.Itoa(len(filePaths))
	}

	resp, err := http.Post(url, contentType, r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println("upload resp status: ", resp.Status)
	fmt.Println("upload resp body: ", string(resp_body))
	return nil
}
