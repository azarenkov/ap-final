package events

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	userv1 "github.com/azarenkov/ap2-final-gen/user/v1"
	"notification-service/internal/usecase"
)

type trainEnvelope struct {
	Type    string `json:"type"`
	TrainID string `json:"train_id"`
	Status  string `json:"status"`
}

type bookingEnvelope struct {
	Type        string `json:"type"`
	BookingID   string `json:"booking_id"`
	UserID      string `json:"user_id"`
	TrainID     string `json:"train_id"`
	AmountCents int64  `json:"amount_cents"`
}

type EmailResolver interface {
	Resolve(ctx context.Context, userID string) string
}

type userServiceResolver struct {
	client userv1.UserServiceClient
}

func NewUserServiceResolver(c userv1.UserServiceClient) EmailResolver {
	return &userServiceResolver{client: c}
}

func (r *userServiceResolver) Resolve(ctx context.Context, userID string) string {
	if r == nil || r.client == nil || userID == "" || userID == "broadcast" || userID == "system" {
		return ""
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	u, err := r.client.GetUserById(ctx, &userv1.GetUserByIdRequest{Id: userID})
	if err != nil {
		log.Printf("email resolve failed user=%s err=%v", userID, err)
		return ""
	}
	return u.Email
}

type Subscriber struct {
	uc       *usecase.NotificationUseCase
	resolver EmailResolver
	subs     []*nats.Subscription
}

func NewSubscriber(uc *usecase.NotificationUseCase, resolver EmailResolver) *Subscriber {
	return &Subscriber{uc: uc, resolver: resolver}
}

func (s *Subscriber) Start(nc *nats.Conn) error {
	if nc == nil {
		return nil
	}
	type bind struct {
		subject string
		handler func(*nats.Msg)
	}
	binds := []bind{
		{"train.delayed", s.onTrainDelay},
		{"train.cancelled", s.onTrainCancel},
		{"booking.created", s.onBookingCreated},
		{"booking.confirmed", s.onBookingConfirmed},
		{"booking.cancelled", s.onBookingCancelled},
		{"payment.succeeded", s.onPaymentSucceeded},
		{"payment.failed", s.onPaymentFailed},
	}
	for _, b := range binds {
		sub, err := nc.Subscribe(b.subject, b.handler)
		if err != nil {
			return err
		}
		s.subs = append(s.subs, sub)
	}
	log.Printf("notification-service subscribed to %d NATS subjects", len(s.subs))
	return nil
}

func (s *Subscriber) Stop() {
	for _, sub := range s.subs {
		_ = sub.Unsubscribe()
	}
}

func (s *Subscriber) emailFor(userID string) string {
	if s.resolver == nil {
		return ""
	}
	return s.resolver.Resolve(context.Background(), userID)
}

func (s *Subscriber) onTrainDelay(m *nats.Msg) {
	var e trainEnvelope
	if json.Unmarshal(m.Data, &e) != nil {
		return
	}
	_ = s.uc.SendTrainDelay(context.Background(), "", "broadcast", e.TrainID, 0)
}

func (s *Subscriber) onTrainCancel(m *nats.Msg) {
	var e trainEnvelope
	if json.Unmarshal(m.Data, &e) != nil {
		return
	}
	_ = s.uc.SendTrainCancellation(context.Background(), "", "broadcast", e.TrainID)
}

func (s *Subscriber) onBookingCreated(m *nats.Msg) {
	var e bookingEnvelope
	if json.Unmarshal(m.Data, &e) != nil {
		return
	}
	to := s.emailFor(e.UserID)
	_, _ = s.uc.Send(context.Background(), to, e.UserID, "BOOKING_CREATED",
		"Booking received",
		"We've received your booking and are awaiting payment. Booking id: "+e.BookingID)
}

func (s *Subscriber) onBookingConfirmed(m *nats.Msg) {
	var e bookingEnvelope
	if json.Unmarshal(m.Data, &e) != nil {
		return
	}
	to := s.emailFor(e.UserID)
	_ = s.uc.SendBookingConfirmation(context.Background(), to, e.UserID, e.BookingID)
}

func (s *Subscriber) onBookingCancelled(m *nats.Msg) {
	var e bookingEnvelope
	if json.Unmarshal(m.Data, &e) != nil {
		return
	}
	to := s.emailFor(e.UserID)
	_ = s.uc.SendBookingCancellation(context.Background(), to, e.UserID, e.BookingID)
}

func (s *Subscriber) onPaymentSucceeded(m *nats.Msg) {
	var e bookingEnvelope
	if json.Unmarshal(m.Data, &e) != nil {
		return
	}
	to := s.emailFor(e.UserID)
	_ = s.uc.SendPaymentSuccess(context.Background(), to, e.UserID, e.BookingID, e.AmountCents)
}

func (s *Subscriber) onPaymentFailed(m *nats.Msg) {
	var e bookingEnvelope
	if json.Unmarshal(m.Data, &e) != nil {
		return
	}
	to := s.emailFor(e.UserID)
	_ = s.uc.SendPaymentFailed(context.Background(), to, e.UserID, e.BookingID, "see payment provider")
}
