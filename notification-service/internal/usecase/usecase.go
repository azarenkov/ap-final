package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"notification-service/internal/domain"
	"notification-service/internal/sender"
)

type Repository interface {
	Insert(ctx context.Context, n *domain.Notification) error
	GetByID(ctx context.Context, id string) (*domain.Notification, error)
	ListByUser(ctx context.Context, userID string) ([]*domain.Notification, error)
	MarkRead(ctx context.Context, id string) error
}

type NotificationUseCase struct {
	repo   Repository
	sender sender.EmailSender
}

func New(repo Repository, s sender.EmailSender) *NotificationUseCase {
	return &NotificationUseCase{repo: repo, sender: s}
}

func (u *NotificationUseCase) Send(ctx context.Context, to, userID, kind, subject, body string) (*domain.Notification, error) {
	n := &domain.Notification{
		ID:        uuid.NewString(),
		UserID:    userID,
		Kind:      kind,
		Subject:   subject,
		Body:      body,
		CreatedAt: time.Now().UTC(),
	}
	if err := u.repo.Insert(ctx, n); err != nil {
		return nil, err
	}
	if to != "" {
		_ = u.sender.Send(to, subject, body)
	}
	return n, nil
}

func (u *NotificationUseCase) SendBookingConfirmation(ctx context.Context, to, userID, bookingID string) error {
	_, err := u.Send(ctx, to, userID, domain.KindBookingConfirmation,
		"Booking confirmed",
		fmt.Sprintf("Your booking %s has been confirmed.", bookingID))
	return err
}

func (u *NotificationUseCase) SendBookingCancellation(ctx context.Context, to, userID, bookingID string) error {
	_, err := u.Send(ctx, to, userID, domain.KindBookingCancellation,
		"Booking cancelled",
		fmt.Sprintf("Your booking %s has been cancelled.", bookingID))
	return err
}

func (u *NotificationUseCase) SendPaymentSuccess(ctx context.Context, to, userID, bookingID string, amount int64) error {
	_, err := u.Send(ctx, to, userID, domain.KindPaymentSuccess,
		"Payment succeeded",
		fmt.Sprintf("Payment of %d cents for booking %s succeeded.", amount, bookingID))
	return err
}

func (u *NotificationUseCase) SendPaymentFailed(ctx context.Context, to, userID, bookingID, reason string) error {
	_, err := u.Send(ctx, to, userID, domain.KindPaymentFailed,
		"Payment failed",
		fmt.Sprintf("Payment for booking %s failed: %s", bookingID, reason))
	return err
}

func (u *NotificationUseCase) SendPasswordReset(ctx context.Context, to, userID, token string) error {
	_, err := u.Send(ctx, to, userID, domain.KindPasswordReset,
		"Password reset requested",
		fmt.Sprintf("Use the following token to reset your password: %s", token))
	return err
}

func (u *NotificationUseCase) SendEmailVerification(ctx context.Context, to, userID, token string) error {
	_, err := u.Send(ctx, to, userID, domain.KindEmailVerification,
		"Verify your email",
		fmt.Sprintf("Use the following token to verify your email: %s", token))
	return err
}

func (u *NotificationUseCase) SendTrainDelay(ctx context.Context, to, userID, trainID string, delayMinutes int32) error {
	_, err := u.Send(ctx, to, userID, domain.KindTrainDelay,
		"Train delayed",
		fmt.Sprintf("Train %s is delayed by %d minutes.", trainID, delayMinutes))
	return err
}

func (u *NotificationUseCase) SendTrainCancellation(ctx context.Context, to, userID, trainID string) error {
	_, err := u.Send(ctx, to, userID, domain.KindTrainCancellation,
		"Train cancelled",
		fmt.Sprintf("Train %s has been cancelled.", trainID))
	return err
}

func (u *NotificationUseCase) Get(ctx context.Context, id string) (*domain.Notification, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *NotificationUseCase) List(ctx context.Context, userID string) ([]*domain.Notification, error) {
	return u.repo.ListByUser(ctx, userID)
}

func (u *NotificationUseCase) MarkRead(ctx context.Context, id string) (*domain.Notification, error) {
	if err := u.repo.MarkRead(ctx, id); err != nil {
		return nil, err
	}
	return u.repo.GetByID(ctx, id)
}
