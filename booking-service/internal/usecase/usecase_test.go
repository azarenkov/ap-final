package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"booking-service/internal/domain"
	"booking-service/internal/events"
	trainv1 "github.com/azarenkov/ap2-final-gen/train/v1"
)

type fakeRepo struct {
	mu      sync.Mutex
	byID    map[string]*domain.Booking
	tickets map[string]string
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byID: map[string]*domain.Booking{}, tickets: map[string]string{}}
}

func (r *fakeRepo) Insert(_ context.Context, b *domain.Booking) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[b.ID] = b
	return nil
}

func (r *fakeRepo) GetByID(_ context.Context, id string) (*domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrBookingNotFound
	}
	cp := *b
	return &cp, nil
}

func (r *fakeRepo) UpdateStatus(_ context.Context, id, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.byID[id]
	if !ok {
		return domain.ErrBookingNotFound
	}
	b.Status = status
	return nil
}

func (r *fakeRepo) UpdateAmount(_ context.Context, id string, amount int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.byID[id]
	if !ok {
		return domain.ErrBookingNotFound
	}
	b.AmountCents = amount
	return nil
}

func (r *fakeRepo) ListByUser(_ context.Context, userID string) ([]*domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*domain.Booking
	for _, b := range r.byID {
		if b.UserID == userID {
			cp := *b
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *fakeRepo) ListPage(_ context.Context, page, size int32) ([]*domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*domain.Booking, 0, len(r.byID))
	for _, b := range r.byID {
		cp := *b
		out = append(out, &cp)
	}
	return out, nil
}

func (r *fakeRepo) InsertTicket(_ context.Context, ticketID, bookingID, code string) (time.Time, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tickets[ticketID] = bookingID
	return time.Now(), nil
}

func (r *fakeRepo) RefundInTx(_ context.Context, id, _ string) (string, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.byID[id]
	if !ok {
		return "", 0, domain.ErrBookingNotFound
	}
	if b.Status != domain.StatusConfirmed && b.Status != domain.StatusCancelled {
		return b.Status, 0, domain.ErrIllegalTransition
	}
	prev := b.Status
	b.Status = domain.StatusRefunded
	return prev, b.AmountCents, nil
}

type fakeTrainClient struct {
	mu         sync.Mutex
	priceCents int64
	available  int32
	rejectAt   int32
	calls      []int32
}

func (c *fakeTrainClient) GetTrainById(_ context.Context, _ *trainv1.GetTrainByIdRequest, _ ...grpc.CallOption) (*trainv1.Train, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return &trainv1.Train{Id: "tr", PriceCents: c.priceCents, AvailableSeats: c.available, TotalSeats: 100}, nil
}

func (c *fakeTrainClient) UpdateSeatAvailability(_ context.Context, in *trainv1.UpdateSeatAvailabilityRequest, _ ...grpc.CallOption) (*trainv1.UpdateSeatAvailabilityResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, in.Delta)
	if c.available+in.Delta < c.rejectAt {
		return nil, errors.New("not enough seats")
	}
	c.available += in.Delta
	return &trainv1.UpdateSeatAvailabilityResponse{TrainId: in.TrainId, AvailableSeats: c.available}, nil
}

func (c *fakeTrainClient) GetAvailableSeats(_ context.Context, _ *trainv1.GetAvailableSeatsRequest, _ ...grpc.CallOption) (*trainv1.AvailableSeatsResponse, error) {
	return nil, nil
}

func (c *fakeTrainClient) CreateTrain(context.Context, *trainv1.CreateTrainRequest, ...grpc.CallOption) (*trainv1.Train, error) {
	return nil, nil
}
func (c *fakeTrainClient) UpdateTrain(context.Context, *trainv1.UpdateTrainRequest, ...grpc.CallOption) (*trainv1.Train, error) {
	return nil, nil
}
func (c *fakeTrainClient) DeleteTrain(context.Context, *trainv1.DeleteTrainRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}
func (c *fakeTrainClient) SearchTrains(context.Context, *trainv1.SearchTrainsRequest, ...grpc.CallOption) (*trainv1.SearchTrainsResponse, error) {
	return nil, nil
}
func (c *fakeTrainClient) GetTrainSchedule(context.Context, *trainv1.GetTrainScheduleRequest, ...grpc.CallOption) (*trainv1.TrainSchedule, error) {
	return nil, nil
}
func (c *fakeTrainClient) CreateRoute(context.Context, *trainv1.CreateRouteRequest, ...grpc.CallOption) (*trainv1.Route, error) {
	return nil, nil
}
func (c *fakeTrainClient) GetRouteById(context.Context, *trainv1.GetRouteByIdRequest, ...grpc.CallOption) (*trainv1.Route, error) {
	return nil, nil
}
func (c *fakeTrainClient) UpdateRoute(context.Context, *trainv1.UpdateRouteRequest, ...grpc.CallOption) (*trainv1.Route, error) {
	return nil, nil
}
func (c *fakeTrainClient) DeleteRoute(context.Context, *trainv1.DeleteRouteRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

type spyPublisher struct {
	mu     sync.Mutex
	events []events.BookingEvent
}

func (s *spyPublisher) Publish(_ context.Context, e events.BookingEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, e)
	return nil
}

func (s *spyPublisher) types() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.events))
	for i, e := range s.events {
		out[i] = e.Type
	}
	return out
}

func newUC(t *testing.T) (*BookingUseCase, *fakeRepo, *fakeTrainClient, *spyPublisher) {
	t.Helper()
	repo := newFakeRepo()
	train := &fakeTrainClient{priceCents: 100_000, available: 50}
	pub := &spyPublisher{}
	return New(repo, train, pub), repo, train, pub
}

func TestCreateBooking_HappyPath(t *testing.T) {
	uc, _, tc, pub := newUC(t)
	b, err := uc.CreateBooking(context.Background(), "alice", "tr", 2)
	if err != nil {
		t.Fatal(err)
	}
	if b.AmountCents != 200_000 {
		t.Fatalf("amount = price*seats: got %d", b.AmountCents)
	}
	if tc.available != 48 {
		t.Fatalf("seats not deducted: %d", tc.available)
	}
	if got := pub.types(); len(got) != 1 || got[0] != events.BookingCreated {
		t.Fatalf("expected booking.created, got %v", got)
	}
}

func TestCreateBooking_TrainRejects_RollsBackSeats(t *testing.T) {
	uc, _, tc, pub := newUC(t)
	tc.available = 1
	tc.rejectAt = 0
	if _, err := uc.CreateBooking(context.Background(), "alice", "tr", 5); !errors.Is(err, domain.ErrTrainUnavailable) {
		t.Fatalf("expected ErrTrainUnavailable, got %v", err)
	}
	if got := pub.types(); len(got) != 0 {
		t.Fatalf("no event should fire on reservation failure, got %v", got)
	}
}

func TestCancel_ReturnsSeatsAndPublishes(t *testing.T) {
	uc, _, tc, pub := newUC(t)
	b, _ := uc.CreateBooking(context.Background(), "alice", "tr", 3)
	before := tc.available
	if _, err := uc.Cancel(context.Background(), b.ID); err != nil {
		t.Fatal(err)
	}
	if tc.available != before+3 {
		t.Fatalf("seats not released: %d (was %d)", tc.available, before)
	}
	if last := pub.types(); last[len(last)-1] != events.BookingCancelled {
		t.Fatalf("expected booking.cancelled last, got %v", last)
	}
}

func TestConfirm_OnlyFromPending(t *testing.T) {
	uc, repo, _, _ := newUC(t)
	b, _ := uc.CreateBooking(context.Background(), "alice", "tr", 1)
	if _, err := uc.Confirm(context.Background(), b.ID); err != nil {
		t.Fatal(err)
	}
	if repo.byID[b.ID].Status != domain.StatusConfirmed {
		t.Fatalf("status not persisted: %s", repo.byID[b.ID].Status)
	}
	if _, err := uc.Confirm(context.Background(), b.ID); !errors.Is(err, domain.ErrIllegalTransition) {
		t.Fatalf("second confirm should fail, got %v", err)
	}
}

func TestProcessPayment_PublishesAppropriateEvent(t *testing.T) {
	uc, _, _, pub := newUC(t)
	b, _ := uc.CreateBooking(context.Background(), "alice", "tr", 1)
	pub.events = nil
	_, _, err := uc.ProcessPaymentMock(context.Background(), b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pub.events) != 1 {
		t.Fatalf("expected exactly one event, got %d", len(pub.events))
	}
	got := pub.events[0].Type
	if got != events.PaymentSucceeded && got != events.PaymentFailed {
		t.Fatalf("expected payment.{succeeded,failed}, got %s", got)
	}
}

func TestGenerateTicket_RequiresConfirmed(t *testing.T) {
	uc, _, _, _ := newUC(t)
	b, _ := uc.CreateBooking(context.Background(), "alice", "tr", 1)
	if _, _, _, err := uc.GenerateTicket(context.Background(), b.ID); !errors.Is(err, domain.ErrIllegalTransition) {
		t.Fatalf("pending should reject ticket, got %v", err)
	}
	_, _ = uc.Confirm(context.Background(), b.ID)
	_, code, _, err := uc.GenerateTicket(context.Background(), b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if code == "" {
		t.Fatal("empty ticket code")
	}
}

func TestRefund_RequiresConfirmedOrCancelled(t *testing.T) {
	uc, _, _, _ := newUC(t)
	b, _ := uc.CreateBooking(context.Background(), "alice", "tr", 1)
	if _, err := uc.Refund(context.Background(), b.ID); !errors.Is(err, domain.ErrIllegalTransition) {
		t.Fatalf("pending should reject refund, got %v", err)
	}
	_, _ = uc.Confirm(context.Background(), b.ID)
	if _, err := uc.Refund(context.Background(), b.ID); err != nil {
		t.Fatalf("confirmed should refund: %v", err)
	}
}
