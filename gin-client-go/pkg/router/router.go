package router

import (
	"awesomeProject1/gin-client-go/pkg/apis"
	"awesomeProject1/gin-client-go/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine)  {
	middleware.InitMiddleware(r)
	r.GET("/ping", apis.Ping)
	r.GET("/namespaces", apis.GetNamespaces)
	r.GET("/pods", apis.Getpods)
	r.GET("/namespace/:namespaceName/pod/:podName/container/:containerName", apis.ExecContainer)
}
