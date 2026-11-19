package exception

import "errors"

var (
	ErrInvalidTripInput          = errors.New("invalid trip input")
	ErrAIUnavailable             = errors.New("ai service unavailable")
	ErrAIBadResponse             = errors.New("ai returned an invalid response")
	ErrInvalidWeatherCoordinates = errors.New("invalid weather coordinates")
)
