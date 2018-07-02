package frontend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"pythia"
)

//Execute the
func Execute(rw http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	fmt.Println("in ex")
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
	}

	var taskEx taskExecute
	if err := json.Unmarshal([]byte(body), &taskEx); err != nil {
		Error422(rw, err)
		return
	}
	fmt.Println(taskEx.Language + ", " + taskEx.Input)
	//To be changed

	json.NewEncoder(rw).Encode("code: " + taskEx.Input + ", lang: " + taskEx.Language)
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
		Error422(rw, err)
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
	Error422(rw, err)

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
		Error422(rw, err)
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

//Error422 response
func Error422(w http.ResponseWriter, err error) {
	//Unprocessable Entity if can't convert to struct
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(422)
	w.Write([]byte("Error 422: Unprocessable Entity "))
	if err := json.NewEncoder(w).Encode(err); err != nil {
		panic(err)
	}
}
