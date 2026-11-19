
package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	redisInfra "github.com/Doratou/Backend/internal/infrastructure/redis"
	"github.com/redis/go-redis/v9"
)

const forecastCacheTTL = 4 * time.Hour

type cachedWeatherProvider struct {
	provider adapter.WeatherProvider
	cache    redisInfra.Cache
}

func NewCachedWeatherProvider(
	provider adapter.WeatherProvider,
	cache redisInfra.Cache,
) adapter.WeatherProvider {
	return &cachedWeatherProvider{
		provider: provider,
		cache:    cache,
	}
}

func (p *cachedWeatherProvider) Forecast(
	ctx context.Context,
	latitude, longitude float64,
	startDate, endDate string,
) (*dto.WeatherForecastResponse, error) {
	if p.provider == nil {
		return nil, fmt.Errorf("weather provider is nil")
	}

	// Use the same rounded coordinates for both the cache key and API request.
	latitude = math.Round(latitude*100) / 100
	longitude = math.Round(longitude*100) / 100

	// v3 prevents old cache entries from being reused with the new validation.
	key := fmt.Sprintf(
		"doratou:weather:v3:%.2f:%.2f:%s:%s",
		latitude,
		longitude,
		startDate,
		endDate,
	)

	var redisClient redis.Cmdable

	if p.cache != nil {
		redisClient = p.cache.GetRDB()
	}

	// 1. Try Redis first.
	if redisClient != nil {
		cached, err := redisClient.Get(ctx, key).Result()

		switch {
		case err == nil:
			var forecast dto.WeatherForecastResponse

			if unmarshalErr := json.Unmarshal([]byte(cached), &forecast); unmarshalErr == nil &&
				isValidCachedForecast(&forecast) {
				log.Printf("[WEATHER CACHE] HIT key=%s", key)
				return &forecast, nil
			}

			log.Printf("[WEATHER CACHE] INVALID entry, deleting key=%s", key)

			if deleteErr := redisClient.Del(ctx, key).Err(); deleteErr != nil {
				log.Printf("[WEATHER CACHE] failed to delete invalid entry: %v", deleteErr)
			}

		case err == redis.Nil:
			log.Printf("[WEATHER CACHE] MISS key=%s", key)

		default:
			// Redis is an optimization. If it is unavailable, continue with API.
			log.Printf("[WEATHER CACHE] GET ERROR: %v", err)
		}
	} else {
		log.Printf("[WEATHER CACHE] Redis is not configured; fetching directly from API")
	}

	// 2. Cache miss: fetch from Open-Meteo.
	log.Printf(
		"[WEATHER API] requesting forecast lat=%.2f lng=%.2f start=%s end=%s",
		latitude,
		longitude,
		startDate,
		endDate,
	)

	forecast, err := p.provider.Forecast(
		ctx,
		latitude,
		longitude,
		startDate,
		endDate,
	)
	if err != nil {
		log.Printf("[WEATHER API] REQUEST ERROR: %v", err)
		return nil, err
	}

	if forecast == nil {
		return nil, fmt.Errorf("weather API returned nil forecast")
	}

	if !isValidCachedForecast(forecast) {
		return nil, fmt.Errorf("weather API returned incomplete or invalid forecast")
	}

	log.Printf(
		"[WEATHER API] SUCCESS: received %d forecast days",
		len(forecast.Daily.Time),
	)

	// 3. Save successful response in Redis.
	if redisClient != nil {
		encoded, marshalErr := json.Marshal(forecast)
		if marshalErr != nil {
			log.Printf("[WEATHER CACHE] failed to encode forecast: %v", marshalErr)
			return forecast, nil
		}

		if setErr := redisClient.Set(
			ctx,
			key,
			encoded,
			forecastCacheTTL,
		).Err(); setErr != nil {
			// Cache failure should not prevent returning a valid API response.
			log.Printf("[WEATHER CACHE] SET ERROR: %v", setErr)
		} else {
			log.Printf(
				"[WEATHER CACHE] SAVED key=%s ttl=%s",
				key,
				forecastCacheTTL,
			)
		}
	}

	return forecast, nil
}

// A forecast is valid when dates and minimum/maximum temperatures align.
// Other daily fields are optional, but must align with dates when present.
func isValidCachedForecast(forecast *dto.WeatherForecastResponse) bool {
	if forecast == nil {
		return false
	}

	daily := forecast.Daily
	count := len(daily.Time)

	if count == 0 {
		return false
	}

	if len(daily.TemperatureMax) != count ||
		len(daily.TemperatureMin) != count {
		return false
	}

	if len(daily.PrecipitationProbability) != 0 &&
		len(daily.PrecipitationProbability) != count {
		return false
	}

	if len(daily.PrecipitationSum) != 0 &&
		len(daily.PrecipitationSum) != count {
		return false
	}

	if len(daily.WeatherCode) != 0 &&
		len(daily.WeatherCode) != count {
		return false
	}

	return true
}
