package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

var (
	APP_NAME    string
	APP_VERSION string
	APP_HOST    string
	APP_PORT    int
)

func NewHTTPServer() *gin.Engine {
	engine := gin.New()
	engine.Use(otelgin.Middleware(APP_NAME))
	engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service-name": APP_NAME,
			"version":      APP_VERSION,
			"time":         time.Now().Unix(),
		})
	})

	return engine
}

func Start(server *gin.Engine) {
	exit := make(chan struct{})
	go func() {
		if err := server.Run(fmt.Sprintf(":%v", APP_PORT)); err != nil {
			log.Errorf("Start service %v failed: %v", APP_NAME, err.Error())
			close(exit)
		}
	}()
	log.Infof("Service %v listening on port %v", APP_NAME, APP_PORT)
	<-exit
}
