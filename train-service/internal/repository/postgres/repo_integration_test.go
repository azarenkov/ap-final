//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"train-service/internal/domain"
)

func setupDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()

	migration := mustReadMigration(t)
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("train_db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("conn str: %v", err)
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}
	if _, err := db.Exec(migration); err != nil {
		t.Fatalf("apply migration: %v", err)
	}
	return db
}

func mustReadMigration(t *testing.T) string {
	t.Helper()

	wd, _ := os.Getwd()
	path := filepath.Join(wd, "..", "..", "..", "migrations", "001_init.sql")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	return string(b)
}

func insertRoute(t *testing.T, repo *RouteRepo) *domain.Route {
	t.Helper()
	r := &domain.Route{
		ID:               uuid.NewString(),
		Origin:           "Astana",
		Destination:      "Almaty",
		DistanceKm:       1200,
		EstimatedMinutes: 960,
	}
	if err := repo.Insert(context.Background(), r); err != nil {
		t.Fatalf("insert route: %v", err)
	}
	return r
}

func TestTrainCRUD_RoundTrips(t *testing.T) {
	db := setupDB(t)
	routes := NewRouteRepo(db)
	trains := NewTrainRepo(db)
	ctx := context.Background()

	route := insertRoute(t, routes)
	dep := time.Now().UTC().Add(1 * time.Hour).Truncate(time.Second)
	arr := dep.Add(8 * time.Hour)
	tr, err := domain.NewTrain(uuid.NewString(), "IC-INT-01", "Integration Train", route.ID, dep, arr, 200, 100_000)
	if err != nil {
		t.Fatal(err)
	}

	if err := trains.Insert(ctx, tr); err != nil {
		t.Fatalf("insert train: %v", err)
	}

	got, err := trains.GetByID(ctx, tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Code != tr.Code || got.TotalSeats != tr.TotalSeats {
		t.Fatalf("round-trip mismatch: %+v", got)
	}

	if err := trains.Delete(ctx, tr.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := trains.GetByID(ctx, tr.ID); !errors.Is(err, domain.ErrTrainNotFound) {
		t.Fatalf("expected ErrTrainNotFound after delete, got %v", err)
	}
}

func TestUpdateAvailableSeats_AtomicAndBounded(t *testing.T) {
	db := setupDB(t)
	routes := NewRouteRepo(db)
	trains := NewTrainRepo(db)
	ctx := context.Background()

	route := insertRoute(t, routes)
	dep := time.Now().UTC().Add(1 * time.Hour).Truncate(time.Second)
	tr, _ := domain.NewTrain(uuid.NewString(), "IC-SEAT", "Seats Test", route.ID, dep, dep.Add(time.Hour), 10, 100)
	if err := trains.Insert(ctx, tr); err != nil {
		t.Fatal(err)
	}

	available, err := trains.UpdateAvailableSeats(ctx, tr.ID, -4)
	if err != nil {
		t.Fatal(err)
	}
	if available != 6 {
		t.Fatalf("expected 6 available, got %d", available)
	}

	if _, err := trains.UpdateAvailableSeats(ctx, tr.ID, -20); err == nil {
		t.Fatal("over-reservation should be rejected, got nil")
	}

	available, err = trains.UpdateAvailableSeats(ctx, tr.ID, 4)
	if err != nil {
		t.Fatal(err)
	}
	if available != 10 {
		t.Fatalf("expected 10 available, got %d", available)
	}
}

func TestSearch_FiltersByRoute(t *testing.T) {
	db := setupDB(t)
	routes := NewRouteRepo(db)
	trains := NewTrainRepo(db)
	ctx := context.Background()

	r1 := insertRoute(t, routes)

	r2 := &domain.Route{ID: uuid.NewString(), Origin: "Atyrau", Destination: "Aktobe", DistanceKm: 500, EstimatedMinutes: 360}
	_ = routes.Insert(ctx, r2)

	dep := time.Now().UTC().Truncate(time.Second).Add(time.Hour)
	for _, route := range []*domain.Route{r1, r1, r2} {
		tr, _ := domain.NewTrain(uuid.NewString(), uuid.NewString()[:8], "X", route.ID, dep, dep.Add(time.Hour), 100, 100)
		_ = trains.Insert(ctx, tr)
	}

	got, total, err := trains.Search(ctx, &domain.SearchFilter{Origin: "Astana", Destination: "Almaty"})
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(got) != 2 {
		t.Fatalf("expected 2 matches for Astana→Almaty, got %d (total=%d)", len(got), total)
	}
}
