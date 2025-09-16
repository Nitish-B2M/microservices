package handlers

import (
	"e-commerce-backend/order/dbs"
	"e-commerce-backend/order/internal/services"
	"e-commerce-backend/shared/middlewares"
	"github.com/gin-gonic/gin"
)

func OrderHandler(router *gin.RouterGroup) {
	orderServices := services.NewService(dbs.DB)
	router.GET("/order/", middlewares.GinAuthMiddleware(), orderServices.GetOrders)
	router.GET("/orders", middlewares.GinAuthMiddleware(), orderServices.GetUserOrders)
	router.POST("/order/create", middlewares.GinAuthMiddleware(), orderServices.CreateOrder)
	router.GET("/order/:order_id", middlewares.GinAuthMiddleware(), orderServices.GetOrderById)
	router.POST("/order/checkout", middlewares.GinAuthMiddleware(), orderServices.Checkout)

	//0.	/orders/checkout
	//2.	/orders/update_status
	//	4. /orders/cancel
}
