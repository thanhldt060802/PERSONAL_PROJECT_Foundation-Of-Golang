package app

import (
	"bytes"
	"encoding/json"
	"net/http"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type MyWebWorker struct {
	act.WebWorker
}

func FactoryMyWebWorker() gen.ProcessBehavior {
	return &MyWebWorker{}
}

func (myWebWorker *MyWebWorker) Init(args ...any) error {
	myWebWorker.Log().Info("Started process %v %v on %v", myWebWorker.PID(), myWebWorker.Name(), myWebWorker.Node().Name())
	return nil
}

func (myWebWorker *MyWebWorker) HandleGet(from gen.PID, writer http.ResponseWriter, request *http.Request) error {
	myWebWorker.Log().Info("Got HTTP request %s", request.URL.Path)

	path := request.URL.Path

	data := struct {
		Status  int
		Code    string
		Message string
	}{}

	switch path {
	case "/path1":
		{
			data.Status = http.StatusOK
			data.Code = "OK"
			data.Message = "Hello"
		}
	case "/path2":
		{
			data.Status = http.StatusOK
			data.Code = "OK"
			data.Message = "Goodbye"
		}
	default:
		{
			data.Status = http.StatusNotFound
			data.Code = "ERR_NOT_FOUND"
			data.Message = "Path is not valid"
		}
	}

	var buf bytes.Buffer
	writer.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.Encode(data)
	writer.Write(buf.Bytes())

	return nil
}
