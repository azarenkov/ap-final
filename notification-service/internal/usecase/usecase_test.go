package usecase

import (
	"context"
	"sync"
	"testing"

	"notification-service/internal/domain"
)

type fakeRepo struct {
	mu   sync.Mutex
	byID map[string]*domain.Notification
	list []*domain.Notification
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byID: map[string]*domain.Notification{}}
}

func (r *fakeRepo) Insert(_ context.Context, n *domain.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[n.ID] = n
	r.list = append(r.list, n)
	return nil
}

func (r *fakeRepo) GetByID(_ context.Context, id string) (*domain.Notification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *n
	return &cp, nil
}

func (r *fakeRepo) ListByUser(_ context.Context, userID string) ([]*domain.Notification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*domain.Notification
	for _, n := range r.list {
		if n.UserID == userID {
			cp := *n
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *fakeRepo) MarkRead(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.byID[id]
	if !ok {
		return domain.ErrNotFound
	}
	n.Read = true
	return nil
}

type spySender struct {
	mu    sync.Mutex
	calls []sentEmail
}

type sentEmail struct{ to, subject, body string }

func (s *spySender) Send(to, subject, body string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, sentEmail{to, subject, body})
	return nil
}

func newUC(t *testing.T) (*NotificationUseCase, *fakeRepo, *spySender) {
	t.Helper()
	repo := newFakeRepo()
	s := &spySender{}
	return New(repo, s), repo, s
}

func TestSend_PersistsAndEmails(t *testing.T) {
	uc, repo, s := newUC(t)
	n, err := uc.Send(context.Background(), "to@example.com", "u1", "GENERIC", "subj", "body")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := repo.byID[n.ID]; !ok {
		t.Fatal("notification not stored")
	}
	if len(s.calls) != 1 || s.calls[0].to != "to@example.com" {
		t.Fatalf("expected one email send, got %+v", s.calls)
	}
}

func TestSend_NoEmailWhenRecipientEmpty(t *testing.T) {
	uc, repo, s := newUC(t)
	n, err := uc.Send(context.Background(), "", "u1", "GENERIC", "subj", "body")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := repo.byID[n.ID]; !ok {
		t.Fatal("notification not stored")
	}
	if len(s.calls) != 0 {
		t.Fatalf("no email expected when to is empty, got %+v", s.calls)
	}
}

func TestHelpers_EmitCorrectKinds(t *testing.T) {
	uc, repo, _ := newUC(t)
	ctx := context.Background()

	_ = uc.SendBookingConfirmation(ctx, "u@x.com", "u", "b1")
	_ = uc.SendBookingCancellation(ctx, "u@x.com", "u", "b1")
	_ = uc.SendPaymentSuccess(ctx, "u@x.com", "u", "b1", 100)
	_ = uc.SendPaymentFailed(ctx, "u@x.com", "u", "b1", "card declined")
	_ = uc.SendPasswordReset(ctx, "u@x.com", "u", "token")
	_ = uc.SendEmailVerification(ctx, "u@x.com", "u", "token")
	_ = uc.SendTrainDelay(ctx, "u@x.com", "u", "tr1", 30)
	_ = uc.SendTrainCancellation(ctx, "u@x.com", "u", "tr1")

	expected := []string{
		domain.KindBookingConfirmation, domain.KindBookingCancellation,
		domain.KindPaymentSuccess, domain.KindPaymentFailed,
		domain.KindPasswordReset, domain.KindEmailVerification,
		domain.KindTrainDelay, domain.KindTrainCancellation,
	}
	got := make([]string, 0, len(repo.list))
	for _, n := range repo.list {
		got = append(got, n.Kind)
	}
	if len(got) != len(expected) {
		t.Fatalf("expected %d notifications, got %d (%v)", len(expected), len(got), got)
	}
	for i, want := range expected {
		if got[i] != want {
			t.Fatalf("at index %d: got %q, want %q", i, got[i], want)
		}
	}
}

func TestMarkRead(t *testing.T) {
	uc, _, _ := newUC(t)
	n, _ := uc.Send(context.Background(), "", "u1", "GENERIC", "s", "b")
	if n.Read {
		t.Fatal("notification should start unread")
	}
	updated, err := uc.MarkRead(context.Background(), n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Read {
		t.Fatal("MarkRead should flip the flag")
	}
}

func TestList(t *testing.T) {
	uc, _, _ := newUC(t)
	ctx := context.Background()
	_, _ = uc.Send(ctx, "", "alice", "GENERIC", "s1", "b1")
	_, _ = uc.Send(ctx, "", "bob", "GENERIC", "s2", "b2")
	_, _ = uc.Send(ctx, "", "alice", "GENERIC", "s3", "b3")

	aliceList, _ := uc.List(ctx, "alice")
	if len(aliceList) != 2 {
		t.Fatalf("alice should have 2 notifications, got %d", len(aliceList))
	}
}
