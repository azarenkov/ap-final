package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"user-service/internal/domain"
	jwtpkg "user-service/internal/jwt"
)

type Repository interface {
	Insert(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateName(ctx context.Context, id, fullName string) error
	UpdatePassword(ctx context.Context, id, hash string) error
	MarkVerified(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, email string) (bool, error)
	RevokeToken(ctx context.Context, jti string, expiresAt time.Time) error
	IsTokenRevoked(ctx context.Context, jti string) (bool, error)
}

type UserUseCase struct {
	repo   Repository
	issuer *jwtpkg.Issuer
}

func New(repo Repository, issuer *jwtpkg.Issuer) *UserUseCase {
	return &UserUseCase{repo: repo, issuer: issuer}
}

func (u *UserUseCase) Create(ctx context.Context, email, password, fullName string) (*domain.User, error) {
	em, err := domain.NormaliseEmail(email)
	if err != nil {
		return nil, err
	}
	if err := domain.ValidatePassword(password); err != nil {
		return nil, err
	}
	if err := domain.ValidateName(fullName); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &domain.User{
		ID:           uuid.NewString(),
		Email:        em,
		PasswordHash: string(hash),
		FullName:     fullName,
		CreatedAt:    time.Now().UTC(),
	}
	if err := u.repo.Insert(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

type LoginResult struct {
	Token     string
	ExpiresAt time.Time
	User      *domain.User
}

func (u *UserUseCase) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	em, err := domain.NormaliseEmail(email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	user, err := u.repo.GetByEmail(ctx, em)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, domain.ErrInvalidCredentials
	}
	tok, _, exp, err := u.issuer.Issue(user.ID)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Token: tok, ExpiresAt: exp, User: user}, nil
}

func (u *UserUseCase) Logout(ctx context.Context, token string) error {
	_, jti, exp, err := u.issuer.Parse(token)
	if err != nil {
		return domain.ErrTokenInvalid
	}
	return u.repo.RevokeToken(ctx, jti, exp)
}

func (u *UserUseCase) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *UserUseCase) GetProfile(ctx context.Context, token string) (*domain.User, error) {
	sub, jti, _, err := u.issuer.Parse(token)
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}
	if revoked, _ := u.repo.IsTokenRevoked(ctx, jti); revoked {
		return nil, domain.ErrTokenRevoked
	}
	return u.repo.GetByID(ctx, sub)
}

func (u *UserUseCase) UpdateProfile(ctx context.Context, token, fullName string) (*domain.User, error) {
	if err := domain.ValidateName(fullName); err != nil {
		return nil, err
	}
	user, err := u.GetProfile(ctx, token)
	if err != nil {
		return nil, err
	}
	if err := u.repo.UpdateName(ctx, user.ID, fullName); err != nil {
		return nil, err
	}
	user.FullName = fullName
	return user, nil
}

func (u *UserUseCase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}

func (u *UserUseCase) ChangePassword(ctx context.Context, token, oldPwd, newPwd string) error {
	if err := domain.ValidatePassword(newPwd); err != nil {
		return err
	}
	sub, _, _, err := u.issuer.Parse(token)
	if err != nil {
		return domain.ErrTokenInvalid
	}
	user, err := u.repo.GetByID(ctx, sub)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPwd)) != nil {
		return domain.ErrInvalidCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return u.repo.UpdatePassword(ctx, sub, string(hash))
}

func (u *UserUseCase) ResetPassword(ctx context.Context, email string) (string, error) {
	em, err := domain.NormaliseEmail(email)
	if err != nil {
		return "", err
	}
	user, err := u.repo.GetByEmail(ctx, em)
	if err != nil {
		return "", err
	}
	tok, _, _, err := u.issuer.Issue(user.ID + ":reset")
	return tok, err
}

func (u *UserUseCase) VerifyEmail(ctx context.Context, token string) error {
	sub, _, _, err := u.issuer.Parse(token)
	if err != nil {
		return domain.ErrTokenInvalid
	}
	return u.repo.MarkVerified(ctx, sub)
}

func (u *UserUseCase) Exists(ctx context.Context, email string) (bool, error) {
	em, err := domain.NormaliseEmail(email)
	if err != nil {
		return false, nil
	}
	return u.repo.Exists(ctx, em)
}
