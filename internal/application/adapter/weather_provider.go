package adapter

import (
	"context"

	"github.com/Doratou/Backend/internal/application/dto"
)

type WeatherProvider interface {
	Forecast(ctx context.Context, latitude, longitude float64, startDate, endDate string) (*dto.WeatherForecastResponse, error)
}
