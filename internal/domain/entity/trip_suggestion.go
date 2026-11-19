package entity

type TripSuggestion struct {
	BaseEntity
	GroupID         int64                  `gorm:"column:group_id;not null;index" json:"group_id"`
	CreatedByUserID int64                  `gorm:"column:created_by_user_id;not null;index" json:"created_by_user_id"`
	Latitude        float64                `gorm:"column:latitude;not null" json:"latitude"`
	Longitude       float64                `gorm:"column:longitude;not null" json:"longitude"`
	StartDate       string                 `gorm:"column:start_date;type:date;not null" json:"start_date"`
	EndDate         string                 `gorm:"column:end_date;type:date;not null" json:"end_date"`
	PeopleCount     int16                  `gorm:"column:people_count;not null" json:"people_count"`
	Input           map[string]interface{} `gorm:"column:input;type:jsonb;serializer:json;not null" json:"input"`
	WeatherSnapshot map[string]interface{} `gorm:"column:weather_snapshot;type:jsonb;serializer:json" json:"weather_snapshot,omitempty"`
	Summary         string                 `gorm:"column:summary;type:text;not null;default:''" json:"summary"`
	Model           string                 `gorm:"column:model;not null" json:"model"`
	PromptVersion   string                 `gorm:"column:prompt_version;not null;default:v1" json:"prompt_version"`
	Items           []TripSuggestionItem   `gorm:"foreignKey:SuggestionID" json:"items,omitempty"`
}

func (TripSuggestion) TableName() string { return "trip_suggestions" }

type TripSuggestionItem struct {
	BaseEntity
	SuggestionID int64  `gorm:"column:suggestion_id;not null;index" json:"suggestion_id"`
	Name         string `gorm:"column:name;type:text;not null" json:"name"`
	Priority     string `gorm:"column:priority;not null" json:"priority"`
	Reason       string `gorm:"column:reason;type:text;not null;default:''" json:"reason"`
}

func (TripSuggestionItem) TableName() string { return "trip_suggestion_items" }
