package temp

import (
	cm "project/common"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Pong(c *gin.Context) {
	c.JSON(http.StatusOK, &cm.EmptyResponse{Result: true})
}
