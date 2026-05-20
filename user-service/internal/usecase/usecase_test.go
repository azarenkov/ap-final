package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"user-service/internal/domain"
	jwtpkg "user-service/internal/jwt"
)

type fakeRepo struct {
	mu      sync.Mutex
	byID    map[string]*domain.User
	byEmail map[string]*domain.User
	revoked map[string]time.Time
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		byID:    map[string]*domain.User{},
		byEmail: map[string]*domain.User{},
		revoked: map[string]time.Time{},
	}
}

func (r *fakeRepo) Insert(_ context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, taken := r.byEmail[u.Email]; taken {
		return domain.ErrEmailTaken
	}
	r.byID[u.ID] = u
	r.byEmail[u.Email] = u
	return nil
}

func (r *fakeRepo) GetByID(_ context.Context, id string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	cp := *u
	return &cp, nil
}

func (r *fakeRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.byEmail[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	cp := *u
	return &cp, nil
}

func (r *fakeRepo) UpdateName(_ context.Context, id, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.byID[id]
	if !ok {
		return domain.ErrUserNotFound
	}
	u.FullName = name
	return nil
}

func (r *fakeRepo) UpdatePassword(_ context.Context, id, hash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.byID[id]
	if !ok {
		return domain.ErrUserNotFound
	}
	u.PasswordHash = hash
	return nil
}

func (r *fakeRepo) MarkVerified(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.byID[id]
	if !ok {
		return domain.ErrUserNotFound
	}
	u.Verified = true
	return nil
}

func (r *fakeRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.byID[id]
	if !ok {
		return domain.ErrUserNotFound
	}
	delete(r.byID, id)
	delete(r.byEmail, u.Email)
	return nil
}

func (r *fakeRepo) Exists(_ context.Context, email string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.byEmail[email]
	return ok, nil
}

func (r *fakeRepo) RevokeToken(_ context.Context, jti string, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.revoked[jti] = expiresAt
	return nil
}

func (r *fakeRepo) IsTokenRevoked(_ context.Context, jti string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.revoked[jti]
	return ok, nil
}

func newUC(t *testing.T) (*UserUseCase, *fakeRepo) {
	t.Helper()
	r := newFakeRepo()
	uc := New(r, jwtpkg.NewIssuer("test-secret", time.Hour))
	return uc, r
}

func TestCreate_HappyPath(t *testing.T) {
	uc, _ := newUC(t)
	u, err := uc.Create(context.Background(), "alice@example.com", "longenoughpw", "Alice")
	if err != nil {
		t.Fatal(err)
	}
	if u.Email != "alice@example.com" {
		t.Fatalf("email not normalised: %q", u.Email)
	}
	if u.PasswordHash == "longenoughpw" {
		t.Fatalf("password stored in plaintext")
	}
}

func TestCreate_Validates(t *testing.T) {
	uc, _ := newUC(t)
	if _, err := uc.Create(context.Background(), "not-an-email", "longenoughpw", "x"); !errors.Is(err, domain.ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
	if _, err := uc.Create(context.Background(), "ok@a.com", "short", "x"); !errors.Is(err, domain.ErrWeakPassword) {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}
	if _, err := uc.Create(context.Background(), "ok@a.com", "longenoughpw", ""); !errors.Is(err, domain.ErrEmptyName) {
		t.Fatalf("expected ErrEmptyName, got %v", err)
	}
}

func TestLogin_RejectsWrongPassword(t *testing.T) {
	uc, _ := newUC(t)
	if _, err := uc.Create(context.Background(), "alice@example.com", "longenoughpw", "Alice"); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.Login(context.Background(), "alice@example.com", "wrong"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("got %v", err)
	}
	res, err := uc.Login(context.Background(), "alice@example.com", "longenoughpw")
	if err != nil {
		t.Fatalf("login should pass with correct pwd: %v", err)
	}
	if res.Token == "" || res.User == nil {
		t.Fatalf("bad login response: %+v", res)
	}
}

func TestChangePassword_RequiresOld(t *testing.T) {
	uc, _ := newUC(t)
	uc.Create(context.Background(), "a@b.com", "longenoughpw", "A")
	res, _ := uc.Login(context.Background(), "a@b.com", "longenoughpw")
	if err := uc.ChangePassword(context.Background(), res.Token, "wrong", "newpassword1"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("got %v, want invalid creds", err)
	}
	if err := uc.ChangePassword(context.Background(), res.Token, "longenoughpw", "newpassword1"); err != nil {
		t.Fatalf("happy path: %v", err)
	}

	if _, err := uc.Login(context.Background(), "a@b.com", "longenoughpw"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("old pwd should fail, got %v", err)
	}
}

func TestLogout_RevokesToken(t *testing.T) {
	uc, _ := newUC(t)
	uc.Create(context.Background(), "a@b.com", "longenoughpw", "A")
	res, _ := uc.Login(context.Background(), "a@b.com", "longenoughpw")
	if _, err := uc.GetProfile(context.Background(), res.Token); err != nil {
		t.Fatalf("profile should succeed before logout: %v", err)
	}
	if err := uc.Logout(context.Background(), res.Token); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.GetProfile(context.Background(), res.Token); !errors.Is(err, domain.ErrTokenRevoked) {
		t.Fatalf("revoked token should fail, got %v", err)
	}
}

func TestExists(t *testing.T) {
	uc, _ := newUC(t)
	uc.Create(context.Background(), "alice@example.com", "longenoughpw", "Alice")
	if ok, _ := uc.Exists(context.Background(), "ALICE@example.COM"); !ok {
		t.Fatal("Exists should be case-insensitive")
	}
	if ok, _ := uc.Exists(context.Background(), "nope@example.com"); ok {
		t.Fatal("non-existent should report false")
	}
}
