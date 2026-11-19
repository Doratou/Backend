package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
)

const forecastURL = "https://api.open-meteo.com/v1/forecast"

type OpenMeteoClient struct {
	httpClient *http.Client
}

func NewOpenMeteoClient() adapter.WeatherProvider {
	return &OpenMeteoClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *OpenMeteoClient) Forecast(
	ctx context.Context,
	lat, lng float64,
	startDate, endDate string,
) (*dto.WeatherForecastResponse, error) {

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
	query.Set(
		"daily",
		"temperature_2m_max,temperature_2m_min,precipitation_probability_max,precipitation_sum,weather_code",
	)

	requestURL := forecastURL + "?" + query.Encode()

	log.Printf("[OPEN-METEO] REQUEST URL: %s", requestURL)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		requestURL,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create Open-Meteo request: %w", err)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[OPEN-METEO] REQUEST ERROR: %v", err)
		return nil, fmt.Errorf("request Open-Meteo forecast: %w", err)
	}
	defer res.Body.Close()

	// Read the complete response body so we can log the API error.
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("[OPEN-METEO] READ RESPONSE ERROR: %v", err)
		return nil, fmt.Errorf("read Open-Meteo response: %w", err)
	}

	log.Printf(
		"[OPEN-METEO] RESPONSE STATUS: %d | CONTENT-TYPE: %s",
		res.StatusCode,
		res.Header.Get("Content-Type"),
	)

	// Log the actual response body when the API returns an error.
	if res.StatusCode != http.StatusOK {
		log.Printf(
			"[OPEN-METEO] ERROR RESPONSE BODY: %s",
			string(body),
		)

		return nil, fmt.Errorf(
			"Open-Meteo returned status %d: %s",
			res.StatusCode,
			string(body),
		)
	}

	var forecast dto.WeatherForecastResponse

	if err := json.Unmarshal(body, &forecast); err != nil {
		log.Printf("[OPEN-METEO] JSON DECODE ERROR: %v", err)
		log.Printf("[OPEN-METEO] RESPONSE BODY: %s", string(body))

		return nil, fmt.Errorf(
			"decode Open-Meteo response: %w",
			err,
		)
	}

	log.Printf(
		"[OPEN-METEO] SUCCESS: received forecast for %d days",
		len(forecast.Daily.Time),
	)

	return &forecast, nil
}
