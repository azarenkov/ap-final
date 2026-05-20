package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	bookingv1 "github.com/azarenkov/ap2-final-gen/booking/v1"
	userv1 "github.com/azarenkov/ap2-final-gen/user/v1"
	"user-service/internal/domain"
	"user-service/internal/usecase"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	uc      *usecase.UserUseCase
	booking bookingv1.BookingServiceClient
}

func NewServer(uc *usecase.UserUseCase, bookingClient bookingv1.BookingServiceClient) *Server {
	return &Server{uc: uc, booking: bookingClient}
}

func (s *Server) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.User, error) {
	u, err := s.uc.Create(ctx, req.Email, req.Password, req.FullName)
	if err != nil {
		return nil, mapErr(err)
	}
	return toProto(u), nil
}

func (s *Server) LoginUser(ctx context.Context, req *userv1.LoginUserRequest) (*userv1.LoginUserResponse, error) {
	res, err := s.uc.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, mapErr(err)
	}
	return &userv1.LoginUserResponse{
		AccessToken: res.Token,
		ExpiresAt:   timestamppb.New(res.ExpiresAt),
		User:        toProto(res.User),
	}, nil
}

func (s *Server) LogoutUser(ctx context.Context, req *userv1.LogoutUserRequest) (*emptypb.Empty, error) {
	if err := s.uc.Logout(ctx, req.AccessToken); err != nil {
		return nil, mapErr(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetUserById(ctx context.Context, req *userv1.GetUserByIdRequest) (*userv1.User, error) {
	u, err := s.uc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return toProto(u), nil
}

func (s *Server) GetUserProfile(ctx context.Context, req *userv1.GetUserProfileRequest) (*userv1.User, error) {
	u, err := s.uc.GetProfile(ctx, req.AccessToken)
	if err != nil {
		return nil, mapErr(err)
	}
	return toProto(u), nil
}

func (s *Server) UpdateUserProfile(ctx context.Context, req *userv1.UpdateUserProfileRequest) (*userv1.User, error) {
	u, err := s.uc.UpdateProfile(ctx, req.AccessToken, req.FullName)
	if err != nil {
		return nil, mapErr(err)
	}
	return toProto(u), nil
}

func (s *Server) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*emptypb.Empty, error) {
	if err := s.uc.Delete(ctx, req.Id); err != nil {
		return nil, mapErr(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ChangePassword(ctx context.Context, req *userv1.ChangePasswordRequest) (*emptypb.Empty, error) {
	if err := s.uc.ChangePassword(ctx, req.AccessToken, req.OldPassword, req.NewPassword); err != nil {
		return nil, mapErr(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ResetPassword(ctx context.Context, req *userv1.ResetPasswordRequest) (*emptypb.Empty, error) {
	if _, err := s.uc.ResetPassword(ctx, req.Email); err != nil {
		return nil, mapErr(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) VerifyEmail(ctx context.Context, req *userv1.VerifyEmailRequest) (*emptypb.Empty, error) {
	if err := s.uc.VerifyEmail(ctx, req.Token); err != nil {
		return nil, mapErr(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetUserBookings(ctx context.Context, req *userv1.GetUserBookingsRequest) (*userv1.UserBookingsResponse, error) {
	if s.booking == nil {
		return &userv1.UserBookingsResponse{}, nil
	}
	out, err := s.booking.GetUserBookings(ctx, &bookingv1.GetUserBookingsRequest{UserId: req.UserId})
	if err != nil {

		return &userv1.UserBookingsResponse{}, nil
	}
	ids := make([]string, 0, len(out.Bookings))
	for _, b := range out.Bookings {
		ids = append(ids, b.Id)
	}
	return &userv1.UserBookingsResponse{BookingIds: ids}, nil
}

func (s *Server) CheckUserExists(ctx context.Context, req *userv1.CheckUserExistsRequest) (*userv1.CheckUserExistsResponse, error) {
	ok, err := s.uc.Exists(ctx, req.Email)
	if err != nil {
		return nil, mapErr(err)
	}
	return &userv1.CheckUserExistsResponse{Exists: ok}, nil
}

func toProto(u *domain.User) *userv1.User {
	return &userv1.User{
		Id:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		Verified:  u.Verified,
		CreatedAt: timestamppb.New(u.CreatedAt),
	}
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrEmailTaken):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrWeakPassword),
		errors.Is(err, domain.ErrEmptyName):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrTokenInvalid),
		errors.Is(err, domain.ErrTokenRevoked):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}
