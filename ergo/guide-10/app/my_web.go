package app

import (
	"net/http"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/meta"
	"ergo.services/meta/websocket"
)

type MyWeb struct {
	act.Actor
}

func FactoryMyWeb() gen.ProcessBehavior {
	return &MyWeb{}
}

func (myWeb *MyWeb) Init(args ...any) error {

	mux := http.NewServeMux()

	websocketOptions := websocket.HandlerOptions{
		ProcessPool: []gen.Atom{
			"worker_1",
			"worker_2",
			"worker_3",
		},
	}
	websocketHandler := websocket.CreateHandler(websocketOptions)

	websocketHandlerId, err := myWeb.SpawnMeta(websocketHandler, gen.MetaOptions{})
	if err != nil {
		myWeb.Log().Error("Spawn WebSocket WebHandler meta-process failed: %s", err.Error())
		return err
	}

	mux.Handle("/", websocketHandler)
	myWeb.Log().Info("Started Websocket Handler to serve '/' (meta-process: %s) successful", websocketHandlerId)

	serverOptions := meta.WebServerOptions{
		Port:    9090,
		Host:    "localhost",
		Handler: mux,
	}

	webserver, err := meta.CreateWebServer(serverOptions)
	if err != nil {
		myWeb.Log().Error("Create Web server meta-process failed: %s", err.Error())
		return err
	}
	webserverid, err := myWeb.SpawnMeta(webserver, gen.MetaOptions{})
	if err != nil {
		webserver.Terminate(err)
	}

	https := ""
	if serverOptions.CertManager != nil {
		https = "s"
	}
	myWeb.Log().Info("started Web server %s: ws%s://%s:%d/", webserverid, https, serverOptions.Host, serverOptions.Port)
	myWeb.Log().Info("you may check it with command below:")
	myWeb.Log().Info("   $ websocat -k ws%s://%s:%d", https, serverOptions.Host, serverOptions.Port)

	return nil
}
