package constants

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
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

func ValidateUserWithCtxUserId(c *gin.Context) error {
	ctxUserId, ok := c.Get("userID")
	if !ok {
		return fmt.Errorf("user id not found in context")
	}
	paramUserId, err := GetUserIdFromParams(c)
	if err != nil {
		return fmt.Errorf("user id not found in params")
	}

	ctxUserIdStr := fmt.Sprintf("%v", ctxUserId)
	if strings.Compare(ctxUserIdStr, fmt.Sprintf("%v", paramUserId)) != 0 {
		return fmt.Errorf("user id not match")
	}

	return nil
}
