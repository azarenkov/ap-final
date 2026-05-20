package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewTrain_Validates(t *testing.T) {
	now := time.Now()
	dep := now.Add(time.Hour)
	arr := dep.Add(2 * time.Hour)

	cases := []struct {
		name       string
		seats      int32
		priceCents int64
		dep, arr   time.Time
		wantErr    error
	}{
		{"happy", 100, 1000, dep, arr, nil},
		{"zero seats", 0, 1000, dep, arr, ErrInvalidSeats},
		{"negative seats", -1, 1000, dep, arr, ErrInvalidSeats},
		{"zero price", 100, 0, dep, arr, ErrInvalidPrice},
		{"arrival before departure", 100, 1000, arr, dep, ErrInvalidTimes},
		{"same instant", 100, 1000, dep, dep, ErrInvalidTimes},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewTrain("id", "code", "name", "route", tc.dep, tc.arr, tc.seats, tc.priceCents)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got err %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestApplySeatDelta(t *testing.T) {
	mk := func(total, available int32) *Train {
		return &Train{TotalSeats: total, AvailableSeats: available}
	}

	t.Run("reserve ok", func(t *testing.T) {
		x := mk(10, 10)
		if err := x.ApplySeatDelta(-3); err != nil {
			t.Fatal(err)
		}
		if x.AvailableSeats != 7 {
			t.Fatalf("got %d", x.AvailableSeats)
		}
	})

	t.Run("release ok", func(t *testing.T) {
		x := mk(10, 5)
		if err := x.ApplySeatDelta(2); err != nil {
			t.Fatal(err)
		}
		if x.AvailableSeats != 7 {
			t.Fatalf("got %d", x.AvailableSeats)
		}
	})

	t.Run("over-reserve rejects", func(t *testing.T) {
		x := mk(10, 5)
		if err := x.ApplySeatDelta(-6); !errors.Is(err, ErrNotEnoughSeats) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("over-release rejects", func(t *testing.T) {
		x := mk(10, 9)
		if err := x.ApplySeatDelta(2); !errors.Is(err, ErrSeatDeltaTooLarge) {
			t.Fatalf("got %v", err)
		}
	})
}

func TestIsValidStatus(t *testing.T) {
	cases := []struct {
		status string
		ok     bool
	}{
		{TrainStatusScheduled, true},
		{TrainStatusDelayed, true},
		{TrainStatusCancelled, true},
		{TrainStatusDeparted, true},
		{TrainStatusArrived, true},
		{"BOGUS", false},
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.status, func(t *testing.T) {
			if got := IsValidStatus(tc.status); got != tc.ok {
				t.Fatalf("got %v, want %v", got, tc.ok)
			}
		})
	}
}

func TestNewRoute_Validates(t *testing.T) {
	cases := []struct {
		name             string
		origin           string
		destination      string
		distanceKm       int32
		estimatedMinutes int32
		wantErr          error
	}{
		{"happy", "A", "B", 100, 60, nil},
		{"empty origin", "", "B", 100, 60, ErrInvalidRouteFields},
		{"zero distance", "A", "B", 0, 60, ErrInvalidRouteFields},
		{"zero minutes", "A", "B", 100, 0, ErrInvalidRouteFields},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewRoute("id", tc.origin, tc.destination, tc.distanceKm, tc.estimatedMinutes)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}
