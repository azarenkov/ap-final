package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"train-service/internal/cache"
	"train-service/internal/domain"
	"train-service/internal/events"
)

type fakeTrainRepo struct {
	mu     sync.Mutex
	byID   map[string]*domain.Train
	failOn string
}

func newFakeTrainRepo() *fakeTrainRepo { return &fakeTrainRepo{byID: map[string]*domain.Train{}} }

func (r *fakeTrainRepo) Insert(_ context.Context, t *domain.Train) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[t.ID] = t
	return nil
}

func (r *fakeTrainRepo) GetByID(_ context.Context, id string) (*domain.Train, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrTrainNotFound
	}
	cp := *t
	return &cp, nil
}

func (r *fakeTrainRepo) Update(_ context.Context, t *domain.Train) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[t.ID]; !ok {
		return domain.ErrTrainNotFound
	}
	r.byID[t.ID] = t
	return nil
}

func (r *fakeTrainRepo) UpdateAvailableSeats(_ context.Context, id string, delta int32) (int32, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.byID[id]
	if !ok {
		return 0, domain.ErrTrainNotFound
	}
	next := t.AvailableSeats + delta
	if next < 0 {
		return 0, domain.ErrNotEnoughSeats
	}
	if next > t.TotalSeats {
		return 0, domain.ErrSeatDeltaTooLarge
	}
	t.AvailableSeats = next
	return t.AvailableSeats, nil
}

func (r *fakeTrainRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[id]; !ok {
		return domain.ErrTrainNotFound
	}
	delete(r.byID, id)
	return nil
}

func (r *fakeTrainRepo) Search(_ context.Context, f *domain.SearchFilter) ([]*domain.Train, int32, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*domain.Train, 0, len(r.byID))
	for _, t := range r.byID {
		cp := *t
		out = append(out, &cp)
	}
	return out, int32(len(out)), nil
}

type fakeRouteRepo struct {
	byID map[string]*domain.Route
}

func newFakeRouteRepo() *fakeRouteRepo { return &fakeRouteRepo{byID: map[string]*domain.Route{}} }

func (r *fakeRouteRepo) Insert(_ context.Context, x *domain.Route) error {
	r.byID[x.ID] = x
	return nil
}
func (r *fakeRouteRepo) GetByID(_ context.Context, id string) (*domain.Route, error) {
	x, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrRouteNotFound
	}
	cp := *x
	return &cp, nil
}
func (r *fakeRouteRepo) Update(_ context.Context, x *domain.Route) error {
	if _, ok := r.byID[x.ID]; !ok {
		return domain.ErrRouteNotFound
	}
	r.byID[x.ID] = x
	return nil
}
func (r *fakeRouteRepo) Delete(_ context.Context, id string) error {
	if _, ok := r.byID[id]; !ok {
		return domain.ErrRouteNotFound
	}
	delete(r.byID, id)
	return nil
}

type spyPublisher struct {
	mu     sync.Mutex
	events []events.TrainEvent
}

func (p *spyPublisher) PublishTrainEvent(_ context.Context, evt events.TrainEvent) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, evt)
	return nil
}

func newUseCase(t *testing.T) (*TrainUseCase, *fakeRouteRepo, *spyPublisher) {
	t.Helper()
	tr := newFakeTrainRepo()
	rr := newFakeRouteRepo()
	pub := &spyPublisher{}
	uc := New(tr, rr, cache.NewTrainCache(nil), pub)
	return uc, rr, pub
}

func newUseCaseWithSpyTrains(t *testing.T) (*TrainUseCase, *fakeTrainRepo, *fakeRouteRepo, *spyPublisher) {
	t.Helper()
	tr := newFakeTrainRepo()
	rr := newFakeRouteRepo()
	pub := &spyPublisher{}
	uc := New(tr, rr, cache.NewTrainCache(nil), pub)
	return uc, tr, rr, pub
}

func TestCreateTrain_RequiresExistingRoute(t *testing.T) {
	uc, _, _ := newUseCase(t)
	_, err := uc.CreateTrain(context.Background(), "C-1", "Train", "missing", time.Now(), time.Now().Add(time.Hour), 100, 1000)
	if !errors.Is(err, domain.ErrRouteNotFound) {
		t.Fatalf("got %v, want ErrRouteNotFound", err)
	}
}

func TestCreateTrain_HappyPath_PublishesEvent(t *testing.T) {
	uc, rr, pub := newUseCase(t)
	_ = rr.Insert(context.Background(), &domain.Route{ID: "r1", Origin: "A", Destination: "B", DistanceKm: 100, EstimatedMinutes: 60})

	dep := time.Now()
	arr := dep.Add(time.Hour)
	tr, err := uc.CreateTrain(context.Background(), "C-1", "Train", "r1", dep, arr, 100, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if tr.AvailableSeats != 100 {
		t.Fatalf("got available_seats=%d", tr.AvailableSeats)
	}
	if len(pub.events) != 1 || pub.events[0].Type != events.TrainCreated {
		t.Fatalf("expected one TrainCreated event, got %+v", pub.events)
	}
}

func TestUpdateTrain_StatusTransition_PublishesDelayed(t *testing.T) {
	uc, tr, rr, pub := newUseCaseWithSpyTrains(t)
	_ = rr.Insert(context.Background(), &domain.Route{ID: "r1", Origin: "A", Destination: "B", DistanceKm: 100, EstimatedMinutes: 60})

	dep := time.Now()
	arr := dep.Add(time.Hour)
	created, _ := uc.CreateTrain(context.Background(), "C-1", "Train", "r1", dep, arr, 100, 1000)
	pub.events = nil

	_, err := uc.UpdateTrain(context.Background(), created.ID, "", time.Time{}, time.Time{}, 0, domain.TrainStatusDelayed)
	if err != nil {
		t.Fatal(err)
	}
	if got := tr.byID[created.ID].Status; got != domain.TrainStatusDelayed {
		t.Fatalf("status not persisted, got %q", got)
	}
	if len(pub.events) != 1 || pub.events[0].Type != events.TrainDelayed {
		t.Fatalf("expected one TrainDelayed event, got %+v", pub.events)
	}
}

func TestUpdateTrain_RejectsUnknownStatus(t *testing.T) {
	uc, _, rr, _ := newUseCaseWithSpyTrains(t)
	_ = rr.Insert(context.Background(), &domain.Route{ID: "r1", Origin: "A", Destination: "B", DistanceKm: 100, EstimatedMinutes: 60})
	dep := time.Now()
	arr := dep.Add(time.Hour)
	created, _ := uc.CreateTrain(context.Background(), "C-1", "Train", "r1", dep, arr, 100, 1000)

	_, err := uc.UpdateTrain(context.Background(), created.ID, "", time.Time{}, time.Time{}, 0, "BOGUS")
	if !errors.Is(err, domain.ErrInvalidStatus) {
		t.Fatalf("got %v, want ErrInvalidStatus", err)
	}
}

func TestUpdateSeatAvailability_HappyPath(t *testing.T) {
	uc, _, rr, _ := newUseCaseWithSpyTrains(t)
	_ = rr.Insert(context.Background(), &domain.Route{ID: "r1", Origin: "A", Destination: "B", DistanceKm: 100, EstimatedMinutes: 60})
	dep := time.Now()
	created, _ := uc.CreateTrain(context.Background(), "C-1", "Train", "r1", dep, dep.Add(time.Hour), 100, 1000)

	remaining, err := uc.UpdateSeatAvailability(context.Background(), created.ID, -5)
	if err != nil {
		t.Fatal(err)
	}
	if remaining != 95 {
		t.Fatalf("got %d", remaining)
	}

	remaining, err = uc.UpdateSeatAvailability(context.Background(), created.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if remaining != 95 {
		t.Fatalf("got %d", remaining)
	}
}

func TestUpdateSeatAvailability_OverReserveRejects(t *testing.T) {
	uc, _, rr, _ := newUseCaseWithSpyTrains(t)
	_ = rr.Insert(context.Background(), &domain.Route{ID: "r1", Origin: "A", Destination: "B", DistanceKm: 100, EstimatedMinutes: 60})
	dep := time.Now()
	created, _ := uc.CreateTrain(context.Background(), "C-1", "Train", "r1", dep, dep.Add(time.Hour), 10, 1000)

	_, err := uc.UpdateSeatAvailability(context.Background(), created.ID, -20)
	if !errors.Is(err, domain.ErrNotEnoughSeats) {
		t.Fatalf("got %v, want ErrNotEnoughSeats", err)
	}
}

func TestCreateRoute_Validates(t *testing.T) {
	uc, _, _ := newUseCase(t)
	_, err := uc.CreateRoute(context.Background(), "", "B", 100, 60)
	if !errors.Is(err, domain.ErrInvalidRouteFields) {
		t.Fatalf("got %v", err)
	}
	r, err := uc.CreateRoute(context.Background(), "A", "B", 100, 60)
	if err != nil {
		t.Fatal(err)
	}
	if r.Origin != "A" || r.DistanceKm != 100 {
		t.Fatalf("unexpected: %+v", r)
	}
}
