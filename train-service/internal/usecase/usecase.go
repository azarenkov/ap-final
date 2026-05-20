package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"train-service/internal/cache"
	"train-service/internal/domain"
	"train-service/internal/events"
)

type TrainRepository interface {
	Insert(ctx context.Context, t *domain.Train) error
	GetByID(ctx context.Context, id string) (*domain.Train, error)
	Update(ctx context.Context, t *domain.Train) error
	UpdateAvailableSeats(ctx context.Context, id string, delta int32) (int32, error)
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, f *domain.SearchFilter) ([]*domain.Train, int32, error)
}

type RouteRepository interface {
	Insert(ctx context.Context, r *domain.Route) error
	GetByID(ctx context.Context, id string) (*domain.Route, error)
	Update(ctx context.Context, r *domain.Route) error
	Delete(ctx context.Context, id string) error
}

type Publisher interface {
	PublishTrainEvent(ctx context.Context, evt events.TrainEvent) error
}

type TrainUseCase struct {
	trains    TrainRepository
	routes    RouteRepository
	cache     *cache.TrainCache
	publisher Publisher
	now       func() time.Time
	newID     func() string
}

func New(trains TrainRepository, routes RouteRepository, c *cache.TrainCache, pub Publisher) *TrainUseCase {
	return &TrainUseCase{
		trains:    trains,
		routes:    routes,
		cache:     c,
		publisher: pub,
		now:       func() time.Time { return time.Now().UTC() },
		newID:     func() string { return uuid.NewString() },
	}
}

func (u *TrainUseCase) CreateTrain(ctx context.Context, code, name, routeID string,
	dep, arr time.Time, totalSeats int32, priceCents int64,
) (*domain.Train, error) {
	if _, err := u.routes.GetByID(ctx, routeID); err != nil {
		return nil, err
	}
	t, err := domain.NewTrain(u.newID(), code, name, routeID, dep, arr, totalSeats, priceCents)
	if err != nil {
		return nil, err
	}
	if err := u.trains.Insert(ctx, t); err != nil {
		return nil, err
	}
	u.cache.InvalidateAll(ctx)
	_ = u.publisher.PublishTrainEvent(ctx, events.TrainEvent{
		Type: events.TrainCreated, TrainID: t.ID, Status: t.Status, At: u.now(),
	})
	return t, nil
}

func (u *TrainUseCase) GetTrainByID(ctx context.Context, id string) (*domain.Train, error) {
	return u.trains.GetByID(ctx, id)
}

func (u *TrainUseCase) UpdateTrain(ctx context.Context, id, name string,
	dep, arr time.Time, priceCents int64, status string,
) (*domain.Train, error) {
	t, err := u.trains.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		t.Name = name
	}
	if !dep.IsZero() {
		t.DepartureTime = dep
	}
	if !arr.IsZero() {
		t.ArrivalTime = arr
	}
	if !t.ArrivalTime.After(t.DepartureTime) {
		return nil, domain.ErrInvalidTimes
	}
	if priceCents > 0 {
		t.PriceCents = priceCents
	}
	prevStatus := t.Status
	if status != "" {
		if !domain.IsValidStatus(status) {
			return nil, domain.ErrInvalidStatus
		}
		t.Status = status
	}
	t.UpdatedAt = u.now()
	if err := u.trains.Update(ctx, t); err != nil {
		return nil, err
	}
	u.cache.InvalidateAll(ctx)
	if prevStatus != t.Status {
		evtType := events.TrainUpdated
		switch t.Status {
		case domain.TrainStatusDelayed:
			evtType = events.TrainDelayed
		case domain.TrainStatusCancelled:
			evtType = events.TrainCancelled
		}
		_ = u.publisher.PublishTrainEvent(ctx, events.TrainEvent{
			Type: evtType, TrainID: t.ID, Status: t.Status, At: u.now(),
		})
	}
	return t, nil
}

func (u *TrainUseCase) DeleteTrain(ctx context.Context, id string) error {
	if err := u.trains.Delete(ctx, id); err != nil {
		return err
	}
	u.cache.InvalidateAll(ctx)
	_ = u.publisher.PublishTrainEvent(ctx, events.TrainEvent{
		Type: events.TrainCancelled, TrainID: id, Status: domain.TrainStatusCancelled, At: u.now(),
	})
	return nil
}

func (u *TrainUseCase) SearchTrains(ctx context.Context, f *domain.SearchFilter) ([]*domain.Train, int32, error) {
	if cached, ok := u.cache.GetSearch(ctx, f); ok {
		return cached.Trains, cached.Total, nil
	}
	trains, total, err := u.trains.Search(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	u.cache.SetSearch(ctx, f, &cache.SearchResult{Trains: trains, Total: total})
	return trains, total, nil
}

func (u *TrainUseCase) GetTrainSchedule(ctx context.Context, trainID string) (*domain.Train, *domain.Route, error) {
	t, err := u.trains.GetByID(ctx, trainID)
	if err != nil {
		return nil, nil, err
	}
	r, err := u.routes.GetByID(ctx, t.RouteID)
	if err != nil {
		return nil, nil, err
	}
	return t, r, nil
}

func (u *TrainUseCase) CreateRoute(ctx context.Context, origin, destination string, distanceKm, estimatedMinutes int32) (*domain.Route, error) {
	r, err := domain.NewRoute(u.newID(), origin, destination, distanceKm, estimatedMinutes)
	if err != nil {
		return nil, err
	}
	if err := u.routes.Insert(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

func (u *TrainUseCase) GetRouteByID(ctx context.Context, id string) (*domain.Route, error) {
	return u.routes.GetByID(ctx, id)
}

func (u *TrainUseCase) UpdateRoute(ctx context.Context, id string, distanceKm, estimatedMinutes int32) (*domain.Route, error) {
	r, err := u.routes.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if distanceKm > 0 {
		r.DistanceKm = distanceKm
	}
	if estimatedMinutes > 0 {
		r.EstimatedMinutes = estimatedMinutes
	}
	if err := u.routes.Update(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

func (u *TrainUseCase) DeleteRoute(ctx context.Context, id string) error {
	return u.routes.Delete(ctx, id)
}

func (u *TrainUseCase) GetAvailableSeats(ctx context.Context, trainID string) (int32, int32, error) {
	t, err := u.trains.GetByID(ctx, trainID)
	if err != nil {
		return 0, 0, err
	}
	return t.TotalSeats, t.AvailableSeats, nil
}

func (u *TrainUseCase) UpdateSeatAvailability(ctx context.Context, trainID string, delta int32) (int32, error) {
	if delta == 0 {
		t, err := u.trains.GetByID(ctx, trainID)
		if err != nil {
			return 0, err
		}
		return t.AvailableSeats, nil
	}
	available, err := u.trains.UpdateAvailableSeats(ctx, trainID, delta)
	if err != nil {
		return 0, err
	}
	u.cache.InvalidateAll(ctx)
	return available, nil
}
