package apis

import (
	"awesomeProject1/gin-client-go/pkg/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetNamespaces(c *gin.Context)  {
	namespaces, err := service.GetNamespace()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
	}
	c.JSON(http.StatusOK, namespaces)
}