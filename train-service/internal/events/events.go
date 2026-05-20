package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	TrainCreated   = "train.created"
	TrainUpdated   = "train.updated"
	TrainDelayed   = "train.delayed"
	TrainCancelled = "train.cancelled"
)

type TrainEvent struct {
	Type    string    `json:"type"`
	TrainID string    `json:"train_id"`
	Status  string    `json:"status"`
	At      time.Time `json:"at"`
}

type NATSPublisher struct {
	nc *nats.Conn
}

func NewNATSPublisher(nc *nats.Conn) *NATSPublisher {
	return &NATSPublisher{nc: nc}
}

func (p *NATSPublisher) PublishTrainEvent(_ context.Context, evt TrainEvent) error {
	if p.nc == nil {
		return nil
	}
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return p.nc.Publish(evt.Type, body)
}
