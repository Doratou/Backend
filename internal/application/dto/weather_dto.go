package dto

type WeatherForecastResponse struct {
	Latitude     float64           `json:"latitude"`
	Longitude    float64           `json:"longitude"`
	Timezone     string            `json:"timezone"`
	TimezoneAbbr string            `json:"timezone_abbreviation"`
	Current      WeatherCurrent    `json:"current"`
	CurrentUnits map[string]string `json:"current_units"`
	Daily        WeatherDaily      `json:"daily"`
	DailyUnits   map[string]string `json:"daily_units"`
}

type WeatherCurrent struct {
	Time        string  `json:"time"`
	Temperature float64 `json:"temperature_2m"`
	WeatherCode int     `json:"weather_code"`
	WindSpeed   float64 `json:"wind_speed_10m"`
}

type WeatherDaily struct {
	Time                     []string  `json:"time"`
	TemperatureMax           []float64 `json:"temperature_2m_max"`
	TemperatureMin           []float64 `json:"temperature_2m_min"`
	PrecipitationProbability []float64 `json:"precipitation_probability_max"`
	PrecipitationSum         []float64 `json:"precipitation_sum"`
	WeatherCode              []int     `json:"weather_code"`
}

type PopularCity struct {
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Province  string  `json:"province"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
