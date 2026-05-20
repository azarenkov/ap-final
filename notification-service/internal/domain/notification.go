package domain

import (
	"errors"
	"time"
)

const (
	KindGeneric             = "GENERIC"
	KindBookingConfirmation = "BOOKING_CONFIRMATION"
	KindBookingCancellation = "BOOKING_CANCELLATION"
	KindPaymentSuccess      = "PAYMENT_SUCCESS"
	KindPaymentFailed       = "PAYMENT_FAILED"
	KindPasswordReset       = "PASSWORD_RESET"
	KindEmailVerification   = "EMAIL_VERIFICATION"
	KindTrainDelay          = "TRAIN_DELAY"
	KindTrainCancellation   = "TRAIN_CANCELLATION"
)

var ErrNotFound = errors.New("notification not found")

type Notification struct {
	ID        string
	UserID    string
	Kind      string
	Subject   string
	Body      string
	Read      bool
	CreatedAt time.Time
}
