package main

import (
	"awesomeProject1/gin-client-go/pkg/config"
	"awesomeProject1/gin-client-go/pkg/router"
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)
func main() {
	engine := gin.Default()
	gin.SetMode(gin.DebugMode)
	router.InitRouter(engine)
	err := engine.Run(fmt.Sprintf("%s:%d", config.GetString(config.ServerHost), config.GetInt(config.SercerPort)))
	if err != nil {
		klog.Fatal(err)
		return
	}
}
