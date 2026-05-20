package domain

import (
	"errors"
	"testing"
)

func TestNew_ValidatesSeatCount(t *testing.T) {
	if _, err := New("id", "u", "tr", 0); !errors.Is(err, ErrInvalidSeatCount) {
		t.Fatalf("zero seats should fail, got %v", err)
	}
	b, err := New("id", "u", "tr", 2)
	if err != nil {
		t.Fatal(err)
	}
	if b.Status != StatusPending {
		t.Fatalf("expected PENDING, got %s", b.Status)
	}
}

func TestStatusFSM(t *testing.T) {
	t.Run("pending → confirmed", func(t *testing.T) {
		b, _ := New("id", "u", "tr", 1)
		if err := b.Confirm(); err != nil {
			t.Fatal(err)
		}
		if b.Status != StatusConfirmed {
			t.Fatalf("got %s", b.Status)
		}
	})

	t.Run("confirmed cannot be confirmed twice", func(t *testing.T) {
		b, _ := New("id", "u", "tr", 1)
		_ = b.Confirm()
		if err := b.Confirm(); !errors.Is(err, ErrIllegalTransition) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("pending → cancelled", func(t *testing.T) {
		b, _ := New("id", "u", "tr", 1)
		if err := b.Cancel(); err != nil {
			t.Fatal(err)
		}
		if b.Status != StatusCancelled {
			t.Fatalf("got %s", b.Status)
		}
	})

	t.Run("refunded cannot be re-cancelled", func(t *testing.T) {
		b := &Booking{Status: StatusRefunded}
		if err := b.Cancel(); !errors.Is(err, ErrIllegalTransition) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("confirmed → refunded", func(t *testing.T) {
		b, _ := New("id", "u", "tr", 1)
		_ = b.Confirm()
		if err := b.Refund(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("pending cannot be refunded", func(t *testing.T) {
		b, _ := New("id", "u", "tr", 1)
		if err := b.Refund(); !errors.Is(err, ErrIllegalTransition) {
			t.Fatalf("got %v", err)
		}
	})
}
