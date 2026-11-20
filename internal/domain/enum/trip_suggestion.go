package enum

// TripType enum values: "city_tour", "beach", "camping", "mountain_hiking", "picnic", "skiing", "desert", "forest_nature", "cultural_historical", "road_trip", "other".
type TripType string

const (
	TripTypeCityTour           TripType = "city_tour"
	TripTypeBeach              TripType = "beach"
	TripTypeCamping            TripType = "camping"
	TripTypeMountainHiking     TripType = "mountain_hiking"
	TripTypePicnic             TripType = "picnic"
	TripTypeSkiing             TripType = "skiing"
	TripTypeDesert             TripType = "desert"
	TripTypeForestNature       TripType = "forest_nature"
	TripTypeCulturalHistorical TripType = "cultural_historical"
	TripTypeRoadTrip           TripType = "road_trip"
	TripTypeOther              TripType = "other"
)

func (t TripType) IsValid() bool {
	switch t {
	case TripTypeCityTour, TripTypeBeach, TripTypeCamping, TripTypeMountainHiking, TripTypePicnic, TripTypeSkiing, TripTypeDesert, TripTypeForestNature, TripTypeCulturalHistorical, TripTypeRoadTrip, TripTypeOther:
		return true
	default:
		return false
	}
}

// Transport enum values: "private_car", "bus", "train", "plane", "motorcycle", "bicycle", "on_foot", "mixed".
type Transport string

const (
	TransportPrivateCar Transport = "private_car"
	TransportBus        Transport = "bus"
	TransportTrain      Transport = "train"
	TransportPlane      Transport = "plane"
	TransportMotorcycle Transport = "motorcycle"
	TransportBicycle    Transport = "bicycle"
	TransportOnFoot     Transport = "on_foot"
	TransportMixed      Transport = "mixed"
)

func (t Transport) IsValid() bool {
	switch t {
	case TransportPrivateCar, TransportBus, TransportTrain, TransportPlane, TransportMotorcycle, TransportBicycle, TransportOnFoot, TransportMixed:
		return true
	default:
		return false
	}
}

// Accommodation enum values: "tent", "hotel", "villa", "guesthouse", "friends_family", "none_day_trip".
type Accommodation string

const (
	AccommodationTent          Accommodation = "tent"
	AccommodationHotel         Accommodation = "hotel"
	AccommodationVilla         Accommodation = "villa"
	AccommodationGuesthouse    Accommodation = "guesthouse"
	AccommodationFriendsFamily Accommodation = "friends_family"
	AccommodationNoneDayTrip   Accommodation = "none_day_trip"
)

func (a Accommodation) IsValid() bool {
	switch a {
	case AccommodationTent, AccommodationHotel, AccommodationVilla, AccommodationGuesthouse, AccommodationFriendsFamily, AccommodationNoneDayTrip:
		return true
	default:
		return false
	}
}

// MealPlan enum values: "self_cooking", "restaurant", "bbq", "mixed".
type MealPlan string

const (
	MealPlanSelfCooking MealPlan = "self_cooking"
	MealPlanRestaurant  MealPlan = "restaurant"
	MealPlanBBQ         MealPlan = "bbq"
	MealPlanMixed       MealPlan = "mixed"
)

func (m MealPlan) IsValid() bool {
	switch m {
	case MealPlanSelfCooking, MealPlanRestaurant, MealPlanBBQ, MealPlanMixed:
		return true
	default:
		return false
	}
}

// TimeOfDay enum values: "daytime", "evening_night", "overnight".
type TimeOfDay string

const (
	TimeOfDayDaytime      TimeOfDay = "daytime"
	TimeOfDayEveningNight TimeOfDay = "evening_night"
	TimeOfDayOvernight    TimeOfDay = "overnight"
)

func (t TimeOfDay) IsValid() bool {
	switch t {
	case TimeOfDayDaytime, TimeOfDayEveningNight, TimeOfDayOvernight:
		return true
	default:
		return false
	}
}

// Companion enum values: "children", "infant", "elderly", "pet", "person_with_disability", "none".
type Companion string

const (
	CompanionChildren             Companion = "children"
	CompanionInfant               Companion = "infant"
	CompanionElderly              Companion = "elderly"
	CompanionPet                  Companion = "pet"
	CompanionPersonWithDisability Companion = "person_with_disability"
	CompanionNone                 Companion = "none"
)

func (c Companion) IsValid() bool {
	switch c {
	case CompanionChildren, CompanionInfant, CompanionElderly, CompanionPet, CompanionPersonWithDisability, CompanionNone:
		return true
	default:
		return false
	}
}

// Activity enum values: "swimming", "hiking", "photography", "fishing", "bbq", "camp_cooking", "skiing", "water_sports", "cycling", "sightseeing", "shopping", "night_party".
type Activity string

const (
	ActivitySwimming    Activity = "swimming"
	ActivityHiking      Activity = "hiking"
	ActivityPhotography Activity = "photography"
	ActivityFishing     Activity = "fishing"
	ActivityBBQ         Activity = "bbq"
	ActivityCampCooking Activity = "camp_cooking"
	ActivitySkiing      Activity = "skiing"
	ActivityWaterSports Activity = "water_sports"
	ActivityCycling     Activity = "cycling"
	ActivitySightseeing Activity = "sightseeing"
	ActivityShopping    Activity = "shopping"
	ActivityNightParty  Activity = "night_party"
)

func (a Activity) IsValid() bool {
	switch a {
	case ActivitySwimming, ActivityHiking, ActivityPhotography, ActivityFishing, ActivityBBQ, ActivityCampCooking, ActivitySkiing, ActivityWaterSports, ActivityCycling, ActivitySightseeing, ActivityShopping, ActivityNightParty:
		return true
	default:
		return false
	}
}
