package router

import (
	"github.com/Doratou/Backend/internal/presentation/v1/handler"
	"github.com/gin-gonic/gin"
)

func RegisterHamyarRoutes(rg *gin.RouterGroup, h *handler.HamyarHandler, auth gin.HandlerFunc) {
	if auth == nil {
		panic("hamyar routes require authentication middleware")
	}
	hamyar := rg.Group("/groups/:groupID/hamyar", auth)
	{
		hamyar.GET("/trip-suggestions", h.RequireGroupMember(), h.ListSuggestions)
	}

	ownerHamyar := rg.Group("/groups/:groupID/hamyar", auth, h.RequireGroupOwner())
	{
		ownerHamyar.POST("/trip-suggestions", h.Suggest)
		ownerHamyar.GET("/weather", h.Forecast)
		ownerHamyar.GET("/cities", h.PopularCities)
	}
}
