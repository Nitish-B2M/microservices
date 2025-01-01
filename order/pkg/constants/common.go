package constants

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

func GetUserIdFromParams(c *gin.Context) (int, error) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return 0, err
	}
	return idInt, nil
}

func GetCartIdFromParams(c *gin.Context) (int, error) {
	id := c.Param("cart_id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return 0, err
	}
	return idInt, nil
}
