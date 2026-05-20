package domain

import (
	"errors"
	"time"
)

const (
	TrainStatusScheduled = "SCHEDULED"
	TrainStatusDelayed   = "DELAYED"
	TrainStatusCancelled = "CANCELLED"
	TrainStatusDeparted  = "DEPARTED"
	TrainStatusArrived   = "ARRIVED"
)

var (
	ErrTrainNotFound      = errors.New("train not found")
	ErrRouteNotFound      = errors.New("route not found")
	ErrInvalidSeats       = errors.New("total_seats must be > 0")
	ErrInvalidPrice       = errors.New("price must be > 0")
	ErrInvalidTimes       = errors.New("arrival_time must be after departure_time")
	ErrInvalidStatus      = errors.New("unknown train status")
	ErrNotEnoughSeats     = errors.New("not enough available seats")
	ErrSeatDeltaTooLarge  = errors.New("delta would exceed total seats")
	ErrInvalidRouteFields = errors.New("route fields invalid")
)

type Train struct {
	ID             string
	Code           string
	Name           string
	RouteID        string
	DepartureTime  time.Time
	ArrivalTime    time.Time
	TotalSeats     int32
	AvailableSeats int32
	PriceCents     int64
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewTrain(id, code, name, routeID string, dep, arr time.Time, totalSeats int32, priceCents int64) (*Train, error) {
	if totalSeats <= 0 {
		return nil, ErrInvalidSeats
	}
	if priceCents <= 0 {
		return nil, ErrInvalidPrice
	}
	if !arr.After(dep) {
		return nil, ErrInvalidTimes
	}
	now := time.Now().UTC()
	return &Train{
		ID:             id,
		Code:           code,
		Name:           name,
		RouteID:        routeID,
		DepartureTime:  dep,
		ArrivalTime:    arr,
		TotalSeats:     totalSeats,
		AvailableSeats: totalSeats,
		PriceCents:     priceCents,
		Status:         TrainStatusScheduled,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func IsValidStatus(s string) bool {
	switch s {
	case TrainStatusScheduled, TrainStatusDelayed, TrainStatusCancelled,
		TrainStatusDeparted, TrainStatusArrived:
		return true
	}
	return false
}

func (t *Train) ApplySeatDelta(delta int32) error {
	next := t.AvailableSeats + delta
	if next < 0 {
		return ErrNotEnoughSeats
	}
	if next > t.TotalSeats {
		return ErrSeatDeltaTooLarge
	}
	t.AvailableSeats = next
	t.UpdatedAt = time.Now().UTC()
	return nil
}
