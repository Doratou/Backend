package enum

type TripType string
type Transport string
type Accommodation string
type MealPlan string
type TimeOfDay string
type Companion string
type Activity string

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

const (
	AccommodationTent          Accommodation = "tent"
	AccommodationHotel         Accommodation = "hotel"
	AccommodationVilla         Accommodation = "villa"
	AccommodationGuesthouse    Accommodation = "guesthouse"
	AccommodationFriendsFamily Accommodation = "friends_family"
	AccommodationNoneDayTrip   Accommodation = "none_day_trip"
)

const (
	MealPlanSelfCooking MealPlan = "self_cooking"
	MealPlanRestaurant  MealPlan = "restaurant"
	MealPlanBBQ         MealPlan = "bbq"
	MealPlanMixed       MealPlan = "mixed"
)

const (
	TimeOfDayDaytime      TimeOfDay = "daytime"
	TimeOfDayEveningNight TimeOfDay = "evening_night"
	TimeOfDayOvernight    TimeOfDay = "overnight"
)

const (
	CompanionChildren             Companion = "children"
	CompanionInfant               Companion = "infant"
	CompanionElderly              Companion = "elderly"
	CompanionPet                  Companion = "pet"
	CompanionPersonWithDisability Companion = "person_with_disability"
	CompanionNone                 Companion = "none"
)

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
