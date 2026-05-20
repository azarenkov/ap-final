package domain

import (
	"errors"
	"time"
)

const (
	StatusPending   = "PENDING"
	StatusConfirmed = "CONFIRMED"
	StatusCancelled = "CANCELLED"
	StatusRefunded  = "REFUNDED"
)

var (
	ErrBookingNotFound   = errors.New("booking not found")
	ErrInvalidSeatCount  = errors.New("seat_count must be > 0")
	ErrIllegalTransition = errors.New("illegal status transition")
	ErrTrainUnavailable  = errors.New("train unavailable or not enough seats")
)

type Booking struct {
	ID          string
	UserID      string
	TrainID     string
	SeatCount   int32
	AmountCents int64
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func New(id, userID, trainID string, seatCount int32) (*Booking, error) {
	if seatCount <= 0 {
		return nil, ErrInvalidSeatCount
	}
	now := time.Now().UTC()
	return &Booking{
		ID: id, UserID: userID, TrainID: trainID, SeatCount: seatCount,
		Status: StatusPending, CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (b *Booking) Confirm() error {
	if b.Status != StatusPending {
		return ErrIllegalTransition
	}
	b.Status = StatusConfirmed
	b.UpdatedAt = time.Now().UTC()
	return nil
}

func (b *Booking) Cancel() error {
	if b.Status != StatusPending && b.Status != StatusConfirmed {
		return ErrIllegalTransition
	}
	b.Status = StatusCancelled
	b.UpdatedAt = time.Now().UTC()
	return nil
}

func (b *Booking) Refund() error {
	if b.Status != StatusCancelled && b.Status != StatusConfirmed {
		return ErrIllegalTransition
	}
	b.Status = StatusRefunded
	b.UpdatedAt = time.Now().UTC()
	return nil
}
