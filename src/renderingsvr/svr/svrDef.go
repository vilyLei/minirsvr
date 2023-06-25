package svr

type RTaskJsonNode struct {
	Id            int64       `json:"id"`
	Name          string      `json:"name"`
	ResUrl        string      `json:"resUrl"`
	Resolution    [2]int      `json:"resolution"`
	Camdvs        [16]float64 `json:"camdvs"`
	BGTransparent int         `json:"bgTransparent"`
	Phase         string      `json:"phase"`
	Action        string      `json:"action"`
}
type RTasksJson struct {
	Tasks []RTaskJsonNode `json:"tasks"`
}
type RTaskJson struct {
	Phase  string        `json:"phase"`
	Task   RTaskJsonNode `json:"task"`
	Status int           `json:"status"`
}

var AutoCheckRTask = false
var RootDir = ""

// func postFileToResSvr(filename string, svrUrl string, phase string, taskID int64, taskName string) error {
// 	bodyBuf := &bytes.Buffer{}
// 	bodyWriter := multipart.NewWriter(bodyBuf)

// 	// this step is very important
// 	fileWriter, err := bodyWriter.CreateFormFile("file", filename)
// 	if err != nil {
// 		fmt.Println("error writing to buffer")
// 		return err
// 	}

// 	// open file handle
// 	fh, err := os.Open(filename)
// 	if err != nil {
// 		fmt.Println("error opening file")
// 		return err
// 	}
// 	defer fh.Close()

// 	_, err = io.Copy(fileWriter, fh)
// 	if err != nil {
// 		return err
// 	}

// 	contentType := bodyWriter.FormDataContentType()
// 	bodyWriter.Close()

// 	url := svrUrl
// 	if taskID > 0 {
// 		taskIDStr := strconv.FormatInt(taskID, 10)
// 		url += "?phase=" + phase + "&taskid=" + taskIDStr + "&taskname=" + taskName
// 	}

// 	resp, err := http.Post(url, contentType, bodyBuf)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()
// 	resp_body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("upload resp status: ", resp.Status)
// 	fmt.Println("upload resp body: ", string(resp_body))
// 	return nil
// }
