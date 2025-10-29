package app

import (
	"net/http"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/meta"
)

type MyWeb struct {
	act.Actor
}

func FactoryMyWeb() gen.ProcessBehavior {
	return &MyWeb{}
}

func (myWeb *MyWeb) Init(args ...any) error {

	mux := http.NewServeMux()

	root := meta.CreateWebHandler(meta.WebHandlerOptions{
		Worker: "my_web_worker",
	})
	rootid, err := myWeb.SpawnMeta(root, gen.MetaOptions{})
	if err != nil {
		myWeb.Log().Error("Spawn WebHandler meta-process failed: %s", err.Error())
		return err
	}

	mux.Handle("/", root)
	myWeb.Log().Info("Started WebHandler to serve '/' (meta-process: %s) successful", rootid)

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

	https := "http"
	if serverOptions.CertManager != nil {
		https = "https"
	}
	myWeb.Log().Info("Started Web server %s: use %s://%s:%d/", webserverid, https, serverOptions.Host, serverOptions.Port)
	myWeb.Log().Info("you may check it with command below:")
	myWeb.Log().Info("   $ curl -k %s://%s:%d", https, serverOptions.Host, serverOptions.Port)

	return nil
}
