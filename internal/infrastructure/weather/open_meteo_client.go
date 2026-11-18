package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
)

const forecastURL = "https://api.open-meteo.com/v1/forecast"

type OpenMeteoClient struct{ httpClient *http.Client }

func NewOpenMeteoClient() adapter.WeatherProvider {
	return &OpenMeteoClient{httpClient: &http.Client{Timeout: 10 * time.Second}}
}

func (c *OpenMeteoClient) Forecast(ctx context.Context, lat, lng float64, startDate, endDate string) (*dto.WeatherForecastResponse, error) {
	query := url.Values{}
	query.Set("latitude", strconv.FormatFloat(lat, 'f', 4, 64))
	query.Set("longitude", strconv.FormatFloat(lng, 'f', 4, 64))
	query.Set("timezone", "auto")
	if startDate != "" && endDate != "" {
		query.Set("start_date", startDate)
		query.Set("end_date", endDate)
	} else {
		query.Set("forecast_days", "7")
	}
	query.Set("current", "temperature_2m,weather_code,wind_speed_10m")
	query.Set("daily", "temperature_2m_max,temperature_2m_min,precipitation_probability_max,precipitation_sum,weather_code")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, forecastURL+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("create Open-Meteo request: %w", err)
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request Open-Meteo forecast: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Open-Meteo returned status %d", res.StatusCode)
	}
	var forecast dto.WeatherForecastResponse
	if err := json.NewDecoder(res.Body).Decode(&forecast); err != nil {
		return nil, fmt.Errorf("decode Open-Meteo response: %w", err)
	}
	return &forecast, nil
}
