package apis
import (
	"awesomeProject1/gin-client-go/pkg/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Getpods(c *gin.Context)  {
	pods, err := service.GetPod()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
	}
	c.JSON(http.StatusOK, pods)
}