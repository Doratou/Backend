package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

func RegisterHamyarRoutes(rg *gin.RouterGroup, h *handler.HamyarHandler, auth gin.HandlerFunc) {
	if auth == nil {
		panic("hamyar routes require authentication middleware")
	}
	hamyar := rg.Group("/groups/:groupID/hamyar", auth, h.RequireGroupOwner())
	{
		hamyar.POST("/trip-suggestions", h.Suggest)
		hamyar.GET("/weather", h.Forecast)
		hamyar.GET("/cities", h.PopularCities)
	}
}
