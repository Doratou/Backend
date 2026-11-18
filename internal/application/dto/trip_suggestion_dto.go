package dto

// تاریخ‌ها با فرمت YYYY-MM-DD ارسال می‌شوند.
type TripSuggestionRequest struct {
	Destination   string   `json:"destination" binding:"required_without=Latitude,max=100"`
	Latitude      *float64 `json:"lat" binding:"required_with=Longitude,omitempty,min=-90,max=90"`
	Longitude     *float64 `json:"lng" binding:"required_with=Latitude,omitempty,min=-180,max=180"`
	StartDate     string   `json:"start_date" binding:"required"`
	EndDate       string   `json:"end_date" binding:"required"`
	People        int      `json:"people" binding:"required,min=1,max=100"`
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
	Name     string `json:"name"`
	Category string `json:"category"`
	Quantity int    `json:"quantity"`
	Priority string `json:"priority"` // ضروری | پیشنهادی | اختیاری
	Reason   string `json:"reason"`
}

type TripSuggestionResponse struct {
	Summary string               `json:"summary"`
	Items   []TripSuggestionItem `json:"items"`
}
