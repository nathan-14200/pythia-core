package frontend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"pythia"
)

//Execute
func Execute(rw http.ResponseWriter, r *http.Request) {

	fmt.Println("in ex")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	var taskEx taskExecute
	if err := json.Unmarshal([]byte(body), &taskEx); err != nil {
		fmt.Println(body)
		fmt.Println(err)
		Error422(rw)
		return
	}

	fmt.Println(taskEx.Language + ", " + taskEx.Input)

	var taskFile string
	lang := taskEx.Language

	switch lang {
	case "python":
		taskFile = "execute-python.sfs"
	case "java":
		taskFile = "execute-java.sfs"
	default:
		Error(rw, "Wrong language syntax")
		return
	}

	// Connection to the pool and execution of the task
	conn := pythia.DialRetry(pythia.QueueAddr)
	defer conn.Close()

	fmt.Println("connected")

	task := pythia.Task{
		Environment: lang,
		TaskFS:      taskFile,
		Limits: struct {
			Time   int `json:"time"`
			Memory int `json:"memory"`
			Disk   int `json:"disk"`
			Output int `json:"output"`
		}{
			Time:   60,
			Memory: 32,
			Disk:   50,
			Output: 1024,
		},
	}

	code := taskEx.Input
	print("code: " + code)

	conn.Send(pythia.Message{
		Message: pythia.LaunchMsg,
		Id:      "test",
		Task:    &task,
		Input:   code,
	})
	fmt.Println("sent")
	//var msg pythia.Message

	if msg, ok := <-conn.Receive(); ok {
		switch msg.Status {
		case "success":
			fmt.Println("success")
			//Sending back Message struct in string
			json.NewEncoder(rw).Encode(msg.Output)
			return
		}

	}
	//If not ok
	//TODO
	return
}

//Echo the given message in a JSON Message struct format
func Echo(rw http.ResponseWriter, r *http.Request) {
	var message map[string]string
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
	}
	if err := json.Unmarshal(body, &message); err != nil {
		Error422(rw)
		return
	}
	for key := range message {
		if key == "text" {
			if err := json.NewEncoder(rw).Encode("Reply: " + message["text"]); err != nil {
				panic(err)
			}
			return
		}
	}
	Error422(rw)

}

// Task function for the server.
func Task(rw http.ResponseWriter, req *http.Request) {
	log.Println("Client connected: ", req.URL)
	if req.Method != "POST" {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Reading the task request
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	var taskReq taskRequest
	if err := json.Unmarshal([]byte(body), &taskReq); err != nil {
		fmt.Println("could not unmarsh taskReq")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	// Connection to the pool and execution of the task
	conn := pythia.DialRetry(pythia.QueueAddr)
	defer conn.Close()
	content, err := ioutil.ReadFile("tasks/" + taskReq.Tid + ".task")
	if err != nil {
		fmt.Println(err)
		fmt.Println("could not read file")
		Error422(rw)
		return
	}
	var task pythia.Task
	if err := json.Unmarshal([]byte(content), &task); err != nil {
		fmt.Println(err)
		fmt.Println("could not unmarsh task")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	conn.Send(pythia.Message{
		Message: pythia.LaunchMsg,
		Id:      "test",
		Task:    &task,
		Input:   taskReq.Response,
	})
	var msg pythia.Message

	if msg, ok := <-conn.Receive(); ok {
		switch msg.Status {
		case "success":
			fmt.Println("success")
			fmt.Fprintf(rw, msg.Output)
		}
		return
	}
	fmt.Println("task status is no success")
	fmt.Println(msg.Status)
	rw.WriteHeader(http.StatusInternalServerError)
}

//Error422 responseused when the data can't be converted to a struct
func Error422(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
	rw.WriteHeader(422)
	message := make(map[string]string)
	message["message"] = "Unprocessable Entity"
	json.NewEncoder(rw).Encode(message)
}

//Error sends which variable was not congruent
func Error(rw http.ResponseWriter, msg string) {
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
	rw.WriteHeader(400)
	message := make(map[string]string)
	message["message"] = msg
	json.NewEncoder(rw).Encode(message)
}
