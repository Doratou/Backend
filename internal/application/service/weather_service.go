package service

import (
	"context"
	"fmt"
	"math"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/exception"
)

type WeatherService interface {
	Forecast(ctx context.Context, latitude, longitude float64) (*dto.WeatherForecastResponse, error)
	ForecastForDates(ctx context.Context, latitude, longitude float64, startDate, endDate string) (*dto.WeatherForecastResponse, error)
	PopularCities() []dto.PopularCity
}

type weatherService struct{ provider adapter.WeatherProvider }

func NewWeatherService(provider adapter.WeatherProvider) WeatherService {
	return &weatherService{provider: provider}
}

func (s *weatherService) Forecast(ctx context.Context, lat, lng float64) (*dto.WeatherForecastResponse, error) {
	return s.forecast(ctx, lat, lng, "", "")
}

func (s *weatherService) ForecastForDates(ctx context.Context, lat, lng float64, startDate, endDate string) (*dto.WeatherForecastResponse, error) {
	return s.forecast(ctx, lat, lng, startDate, endDate)
}

func (s *weatherService) forecast(ctx context.Context, lat, lng float64, startDate, endDate string) (*dto.WeatherForecastResponse, error) {
	if math.IsNaN(lat) || math.IsInf(lat, 0) || lat < -90 || lat > 90 || math.IsNaN(lng) || math.IsInf(lng, 0) || lng < -180 || lng > 180 {
		return nil, fmt.Errorf("%w: latitude must be -90..90 and longitude -180..180", exception.ErrInvalidWeatherCoordinates)
	}
	return s.provider.Forecast(ctx, lat, lng, startDate, endDate)
}

func (s *weatherService) PopularCities() []dto.PopularCity {
	return append([]dto.PopularCity(nil), popularIranianCities...)
}

var popularIranianCities = []dto.PopularCity{
	{Name: "تهران", Slug: "tehran", Province: "تهران", Latitude: 35.6892, Longitude: 51.3890},
	{Name: "مشهد", Slug: "mashhad", Province: "خراسان رضوی", Latitude: 36.2972, Longitude: 59.6067},
	{Name: "اصفهان", Slug: "isfahan", Province: "اصفهان", Latitude: 32.6546, Longitude: 51.6680},
	{Name: "شیراز", Slug: "shiraz", Province: "فارس", Latitude: 29.5918, Longitude: 52.5837},
	{Name: "تبریز", Slug: "tabriz", Province: "آذربایجان شرقی", Latitude: 38.0962, Longitude: 46.2738},
	{Name: "کرج", Slug: "karaj", Province: "البرز", Latitude: 35.8400, Longitude: 50.9391},
	{Name: "اهواز", Slug: "ahvaz", Province: "خوزستان", Latitude: 31.3183, Longitude: 48.6706},
	{Name: "قم", Slug: "qom", Province: "قم", Latitude: 34.6416, Longitude: 50.8746},
	{Name: "رشت", Slug: "rasht", Province: "گیلان", Latitude: 37.2808, Longitude: 49.5832},
	{Name: "کرمانشاه", Slug: "kermanshah", Province: "کرمانشاه", Latitude: 34.3142, Longitude: 47.0650},
	{Name: "ارومیه", Slug: "urmia", Province: "آذربایجان غربی", Latitude: 37.5527, Longitude: 45.0761},
	{Name: "یزد", Slug: "yazd", Province: "یزد", Latitude: 31.8974, Longitude: 54.3569},
	{Name: "بندرعباس", Slug: "bandar-abbas", Province: "هرمزگان", Latitude: 27.1832, Longitude: 56.2666},
	{Name: "کرمان", Slug: "kerman", Province: "کرمان", Latitude: 30.2832, Longitude: 57.0788},
	{Name: "همدان", Slug: "hamadan", Province: "همدان", Latitude: 34.7989, Longitude: 48.5150},
	{Name: "اردبیل", Slug: "ardabil", Province: "اردبیل", Latitude: 38.2498, Longitude: 48.2933},
	{Name: "قزوین", Slug: "qazvin", Province: "قزوین", Latitude: 36.2688, Longitude: 50.0041},
	{Name: "زنجان", Slug: "zanjan", Province: "زنجان", Latitude: 36.6736, Longitude: 48.4787},
	{Name: "گرگان", Slug: "gorgan", Province: "گلستان", Latitude: 36.8456, Longitude: 54.4393},
	{Name: "ساری", Slug: "sari", Province: "مازندران", Latitude: 36.5633, Longitude: 53.0601},
	{Name: "کاشان", Slug: "kashan", Province: "اصفهان", Latitude: 33.9850, Longitude: 51.4100},
	{Name: "کیش", Slug: "kish", Province: "هرمزگان", Latitude: 26.5320, Longitude: 53.9670},
	{Name: "قشم", Slug: "qeshm", Province: "هرمزگان", Latitude: 26.9581, Longitude: 56.2719},
	{Name: "سنندج", Slug: "sanandaj", Province: "کردستان", Latitude: 35.3149, Longitude: 46.9988},
	{Name: "خرم‌آباد", Slug: "khorramabad", Province: "لرستان", Latitude: 33.4878, Longitude: 48.3558},
	{Name: "بوشهر", Slug: "bushehr", Province: "بوشهر", Latitude: 28.9234, Longitude: 50.8203},
	{Name: "بابلسر", Slug: "babolsar", Province: "مازندران", Latitude: 36.7025, Longitude: 52.6576},
	{Name: "رامسر", Slug: "ramsar", Province: "مازندران", Latitude: 36.9034, Longitude: 50.6585},
}
