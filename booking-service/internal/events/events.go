package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	BookingCreated   = "booking.created"
	BookingConfirmed = "booking.confirmed"
	BookingCancelled = "booking.cancelled"
	BookingRefunded  = "booking.refunded"
	PaymentSucceeded = "payment.succeeded"
	PaymentFailed    = "payment.failed"
)

type BookingEvent struct {
	Type        string    `json:"type"`
	BookingID   string    `json:"booking_id"`
	UserID      string    `json:"user_id"`
	TrainID     string    `json:"train_id"`
	SeatCount   int32     `json:"seat_count"`
	AmountCents int64     `json:"amount_cents"`
	At          time.Time `json:"at"`
}

type Publisher struct{ nc *nats.Conn }

func NewPublisher(nc *nats.Conn) *Publisher { return &Publisher{nc: nc} }

func (p *Publisher) Publish(_ context.Context, evt BookingEvent) error {
	if p == nil || p.nc == nil {
		return nil
	}
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return p.nc.Publish(evt.Type, body)
}
