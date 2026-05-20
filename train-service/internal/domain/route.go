package domain

type Route struct {
	ID               string
	Origin           string
	Destination      string
	DistanceKm       int32
	EstimatedMinutes int32
}

func NewRoute(id, origin, destination string, distanceKm, estimatedMinutes int32) (*Route, error) {
	if origin == "" || destination == "" || distanceKm <= 0 || estimatedMinutes <= 0 {
		return nil, ErrInvalidRouteFields
	}
	return &Route{
		ID:               id,
		Origin:           origin,
		Destination:      destination,
		DistanceKm:       distanceKm,
		EstimatedMinutes: estimatedMinutes,
	}, nil
}
