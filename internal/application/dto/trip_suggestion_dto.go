package dto

import "time"

type TripSuggestionRequest struct {
	CityName      string   `json:"city_name" binding:"required,max=100"`
	Latitude      *float64 `json:"lat" binding:"required,min=-90,max=90"`
	Longitude     *float64 `json:"lng" binding:"required,min=-180,max=180"`
	StartDate     string   `json:"start_date" binding:"required"`
	EndDate       string   `json:"end_date" binding:"required"`
	PeopleCount   int      `json:"people_count" binding:"required,min=1,max=100"`
	TripType      string   `json:"trip_type" binding:"omitempty,oneof=city_tour beach camping mountain_hiking picnic skiing desert forest_nature cultural_historical road_trip other"`
	Transport     string   `json:"transport" binding:"omitempty,oneof=private_car bus train plane motorcycle bicycle on_foot mixed"`
	Accommodation string   `json:"accommodation" binding:"omitempty,oneof=tent hotel villa guesthouse friends_family none_day_trip"`
	MealPlan      string   `json:"meal_plan" binding:"omitempty,oneof=self_cooking restaurant bbq mixed"`
	TimeOfDay     string   `json:"time_of_day" binding:"omitempty,oneof=daytime evening_night overnight"`
	Companions    []string `json:"companions" binding:"omitempty,max=6,dive,oneof=children infant elderly pet person_with_disability none"`
	Activities    []string `json:"activities" binding:"omitempty,max=12,dive,oneof=swimming hiking photography fishing bbq camp_cooking skiing water_sports cycling sightseeing shopping night_party"`
	Notes         string   `json:"notes" binding:"omitempty,max=500"`
}

type TripSuggestionItem struct {
	Name     string                 `json:"name"`
	Priority TripSuggestionPriority `json:"priority"`
	Reason   string                 `json:"reason"`
}

type TripSuggestionPriority string

const (
	TripSuggestionPriorityEssential   TripSuggestionPriority = "ضروری"
	TripSuggestionPriorityRecommended TripSuggestionPriority = "پیشنهادی"
	TripSuggestionPriorityOptional    TripSuggestionPriority = "اختیاری"
)

type TripSuggestionResponse struct {
	Summary string               `json:"summary"`
	Items   []TripSuggestionItem `json:"items"`
}

type TripSuggestionHistoryItem struct {
	ID              int64                  `json:"id"`
	GroupID         int64                  `json:"group_id"`
	CreatedByUserID int64                  `json:"created_by_user_id"`
	Latitude        float64                `json:"latitude"`
	Longitude       float64                `json:"longitude"`
	StartDate       string                 `json:"start_date"`
	EndDate         string                 `json:"end_date"`
	PeopleCount     int16                  `json:"people_count"`
	Input           map[string]interface{} `json:"input"`
	WeatherSnapshot map[string]interface{} `json:"weather_snapshot,omitempty"`
	Summary         string                 `json:"summary"`
	Items           []TripSuggestionItem   `json:"items"`
	Model           string                 `json:"model"`
	PromptVersion   string                 `json:"prompt_version"`
	CreatedAt       time.Time              `json:"created_at"`
}

type TripSuggestionHistoryResponse struct {
	Items []TripSuggestionHistoryItem `json:"items"`
	Total int64                       `json:"total"`
	Page  int                         `json:"page"`
	Limit int                         `json:"limit"`
}
