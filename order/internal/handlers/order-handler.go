package handlers

import (
	"e-commerce-backend/order/internal/services"
	"e-commerce-backend/payment/dbs"
	"e-commerce-backend/shared/middlewares"
	"github.com/gin-gonic/gin"
)

func OrderHandler(router *gin.RouterGroup) {
	orderServices := services.NewService(dbs.DB)
	router.GET("/", middlewares.GinAuthMiddleware(), orderServices.GetOrders)
	router.POST("/add", middlewares.GinAuthMiddleware(), orderServices.CreateOrder)
	router.GET("/:order_id", middlewares.GinAuthMiddleware(), orderServices.GetOrderById)
	router.POST("/checkout", middlewares.GinAuthMiddleware(), orderServices.Checkout)
}
