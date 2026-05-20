package usecase

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"

	"booking-service/internal/domain"
	"booking-service/internal/events"
	trainv1 "github.com/azarenkov/ap2-final-gen/train/v1"
)

type Repository interface {
	Insert(ctx context.Context, b *domain.Booking) error
	GetByID(ctx context.Context, id string) (*domain.Booking, error)
	UpdateStatus(ctx context.Context, id, status string) error
	UpdateAmount(ctx context.Context, id string, amount int64) error
	ListByUser(ctx context.Context, userID string) ([]*domain.Booking, error)
	ListPage(ctx context.Context, page, size int32) ([]*domain.Booking, error)
	InsertTicket(ctx context.Context, ticketID, bookingID, code string) (time.Time, error)
	RefundInTx(ctx context.Context, id, actor string) (prevStatus string, amount int64, err error)
}

type Publisher interface {
	Publish(ctx context.Context, e events.BookingEvent) error
}

type BookingUseCase struct {
	repo  Repository
	train trainv1.TrainServiceClient
	pub   Publisher
	rng   *rand.Rand
}

func New(repo Repository, train trainv1.TrainServiceClient, pub Publisher) *BookingUseCase {
	return &BookingUseCase{repo: repo, train: train, pub: pub, rng: rand.New(rand.NewPCG(1, 2))}
}

func (u *BookingUseCase) CreateBooking(ctx context.Context, userID, trainID string, seatCount int32) (*domain.Booking, error) {
	if u.train == nil {
		return nil, fmt.Errorf("train client unavailable")
	}
	tr, err := u.train.GetTrainById(ctx, &trainv1.GetTrainByIdRequest{Id: trainID})
	if err != nil {
		return nil, err
	}
	if _, err := u.train.UpdateSeatAvailability(ctx, &trainv1.UpdateSeatAvailabilityRequest{
		TrainId: trainID, Delta: -seatCount,
	}); err != nil {
		return nil, domain.ErrTrainUnavailable
	}
	b, err := domain.New(uuid.NewString(), userID, trainID, seatCount)
	if err != nil {

		_, _ = u.train.UpdateSeatAvailability(ctx, &trainv1.UpdateSeatAvailabilityRequest{TrainId: trainID, Delta: seatCount})
		return nil, err
	}
	b.AmountCents = tr.PriceCents * int64(seatCount)
	if err := u.repo.Insert(ctx, b); err != nil {
		_, _ = u.train.UpdateSeatAvailability(ctx, &trainv1.UpdateSeatAvailabilityRequest{TrainId: trainID, Delta: seatCount})
		return nil, err
	}
	_ = u.pub.Publish(ctx, events.BookingEvent{
		Type: events.BookingCreated, BookingID: b.ID, UserID: userID, TrainID: trainID,
		SeatCount: seatCount, AmountCents: b.AmountCents, At: time.Now().UTC(),
	})
	return b, nil
}

func (u *BookingUseCase) Get(ctx context.Context, id string) (*domain.Booking, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *BookingUseCase) ListByUser(ctx context.Context, userID string) ([]*domain.Booking, error) {
	return u.repo.ListByUser(ctx, userID)
}

func (u *BookingUseCase) ListPage(ctx context.Context, page, size int32) ([]*domain.Booking, error) {
	return u.repo.ListPage(ctx, page, size)
}

func (u *BookingUseCase) Cancel(ctx context.Context, id string) (*domain.Booking, error) {
	b, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := b.Cancel(); err != nil {
		return nil, err
	}
	if u.train != nil {
		_, _ = u.train.UpdateSeatAvailability(ctx, &trainv1.UpdateSeatAvailabilityRequest{TrainId: b.TrainID, Delta: b.SeatCount})
	}
	if err := u.repo.UpdateStatus(ctx, id, b.Status); err != nil {
		return nil, err
	}
	_ = u.pub.Publish(ctx, events.BookingEvent{
		Type: events.BookingCancelled, BookingID: b.ID, UserID: b.UserID, TrainID: b.TrainID,
		SeatCount: b.SeatCount, AmountCents: b.AmountCents, At: time.Now().UTC(),
	})
	return b, nil
}

func (u *BookingUseCase) Confirm(ctx context.Context, id string) (*domain.Booking, error) {
	b, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := b.Confirm(); err != nil {
		return nil, err
	}
	if err := u.repo.UpdateStatus(ctx, id, b.Status); err != nil {
		return nil, err
	}
	_ = u.pub.Publish(ctx, events.BookingEvent{
		Type: events.BookingConfirmed, BookingID: b.ID, UserID: b.UserID, TrainID: b.TrainID,
		SeatCount: b.SeatCount, AmountCents: b.AmountCents, At: time.Now().UTC(),
	})
	return b, nil
}

func (u *BookingUseCase) Status(ctx context.Context, id string) (string, error) {
	b, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	return b.Status, nil
}

func (u *BookingUseCase) ReserveSeat(ctx context.Context, userID, trainID string, seatCount int32) (*domain.Booking, error) {
	return u.CreateBooking(ctx, userID, trainID, seatCount)
}

func (u *BookingUseCase) ReleaseSeat(ctx context.Context, bookingID string) (*domain.Booking, error) {
	return u.Cancel(ctx, bookingID)
}

func (u *BookingUseCase) ProcessPaymentMock(ctx context.Context, bookingID string) (bool, string, error) {
	b, err := u.repo.GetByID(ctx, bookingID)
	if err != nil {
		return false, "", err
	}
	success := u.rng.Float64() < 0.8
	evt := events.BookingEvent{
		BookingID: bookingID, UserID: b.UserID, TrainID: b.TrainID,
		SeatCount: b.SeatCount, AmountCents: b.AmountCents, At: time.Now().UTC(),
	}
	if success {
		_ = b.Confirm()
		_ = u.repo.UpdateStatus(ctx, b.ID, b.Status)
		evt.Type = events.PaymentSucceeded
		_ = u.pub.Publish(ctx, evt)
		return true, "ok", nil
	}
	evt.Type = events.PaymentFailed
	_ = u.pub.Publish(ctx, evt)
	return false, "card declined (mock)", nil
}

func (u *BookingUseCase) GenerateTicket(ctx context.Context, bookingID string) (string, string, time.Time, error) {
	b, err := u.repo.GetByID(ctx, bookingID)
	if err != nil {
		return "", "", time.Time{}, err
	}
	if b.Status != domain.StatusConfirmed {
		return "", "", time.Time{}, domain.ErrIllegalTransition
	}
	ticketID := uuid.NewString()
	code := fmt.Sprintf("TKT-%s", uuid.NewString()[:8])
	issuedAt, err := u.repo.InsertTicket(ctx, ticketID, bookingID, code)
	return ticketID, code, issuedAt, err
}

func (u *BookingUseCase) Refund(ctx context.Context, id string) (int64, error) {
	b, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return 0, err
	}
	if _, amount, err := u.repo.RefundInTx(ctx, id, "system"); err != nil {
		return 0, err
	} else {
		b.AmountCents = amount
	}
	_ = u.pub.Publish(ctx, events.BookingEvent{
		Type: events.BookingRefunded, BookingID: id, UserID: b.UserID, TrainID: b.TrainID,
		SeatCount: b.SeatCount, AmountCents: b.AmountCents, At: time.Now().UTC(),
	})
	return b.AmountCents, nil
}
