package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/application/service"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/tokencontext"
	"github.com/gin-gonic/gin"
)

type HamyarHandler struct {
	tripSuggestion adapter.TripSuggestionService
	weather        service.WeatherService
	groups         adapter.GroupService
}

func NewHamyarHandler(tripSuggestion adapter.TripSuggestionService, weather service.WeatherService, groups adapter.GroupService) *HamyarHandler {
	return &HamyarHandler{tripSuggestion: tripSuggestion, weather: weather, groups: groups}
}

func (h *HamyarHandler) RequireGroupOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		groupID, err := parseIDParam(c, "groupID")
		if err != nil || groupID <= 0 {
			failLogged(c, exception.ErrGroupInvalidID)
			c.Abort()
			return
		}
		userID, ok := tokencontext.GetUserID(c)
		role, roleOK := tokencontext.GetUserRole(c)
		if !ok || !roleOK {
			failLogged(c, exception.ErrMissingToken)
			c.Abort()
			return
		}
		group, err := h.groups.Get(c.Request.Context(), userID, role, groupID)
		if err != nil {
			failLogged(c, err)
			c.Abort()
			return
		}
		if group.CreatedByUserID != userID {
			failLogged(c, exception.ErrGroupOwnerRequired)
			c.Abort()
			return
		}
		c.Next()
	}
}

func (h *HamyarHandler) Suggest(c *gin.Context) {
	var req dto.TripSuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}
	res, err := h.tripSuggestion.SuggestItems(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, exception.ErrInvalidTripInput):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, exception.ErrAIUnavailable), errors.Is(err, exception.ErrAIBadResponse):
			log.Printf("trip suggestion ai error: %v", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "ai service is currently unavailable, please try again"})
		default:
			log.Printf("trip suggestion error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *HamyarHandler) Forecast(c *gin.Context) {
	lat, errLat := strconv.ParseFloat(c.Query("lat"), 64)
	lng, errLng := strconv.ParseFloat(c.Query("lng"), 64)
	if errLat != nil || errLng != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat and lng query parameters are required numeric coordinates"})
		return
	}
	forecast, err := h.weather.Forecast(c.Request.Context(), lat, lng)
	if err != nil {
		if errors.Is(err, exception.ErrInvalidWeatherCoordinates) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Printf("weather forecast error: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "weather service is currently unavailable"})
		return
	}
	c.JSON(http.StatusOK, forecast)
}

func (h *HamyarHandler) PopularCities(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"cities": h.weather.PopularCities()})
}
