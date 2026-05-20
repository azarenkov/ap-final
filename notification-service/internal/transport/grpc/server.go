package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	notificationv1 "github.com/azarenkov/ap2-final-gen/notification/v1"
	"notification-service/internal/domain"
	"notification-service/internal/usecase"
)

type Server struct {
	notificationv1.UnimplementedNotificationServiceServer
	uc *usecase.NotificationUseCase
}

func NewServer(uc *usecase.NotificationUseCase) *Server { return &Server{uc: uc} }

func (s *Server) SendEmail(ctx context.Context, req *notificationv1.SendEmailRequest) (*emptypb.Empty, error) {
	kind := req.Kind
	if kind == "" {
		kind = domain.KindGeneric
	}
	if _, err := s.uc.Send(ctx, req.To, req.UserId, kind, req.Subject, req.Body); err != nil {
		return nil, mapErr(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) SendBookingConfirmation(ctx context.Context, req *notificationv1.SendBookingConfirmationRequest) (*emptypb.Empty, error) {
	return ackOrErr(s.uc.SendBookingConfirmation(ctx, req.To, req.UserId, req.BookingId))
}

func (s *Server) SendBookingCancellation(ctx context.Context, req *notificationv1.SendBookingCancellationRequest) (*emptypb.Empty, error) {
	return ackOrErr(s.uc.SendBookingCancellation(ctx, req.To, req.UserId, req.BookingId))
}

func (s *Server) SendPaymentSuccess(ctx context.Context, req *notificationv1.SendPaymentSuccessRequest) (*emptypb.Empty, error) {
	return ackOrErr(s.uc.SendPaymentSuccess(ctx, req.To, req.UserId, req.BookingId, req.AmountCents))
}

func (s *Server) SendPaymentFailed(ctx context.Context, req *notificationv1.SendPaymentFailedRequest) (*emptypb.Empty, error) {
	return ackOrErr(s.uc.SendPaymentFailed(ctx, req.To, req.UserId, req.BookingId, req.Reason))
}

func (s *Server) SendPasswordResetEmail(ctx context.Context, req *notificationv1.SendPasswordResetEmailRequest) (*emptypb.Empty, error) {
	return ackOrErr(s.uc.SendPasswordReset(ctx, req.To, req.UserId, req.ResetToken))
}

func (s *Server) SendEmailVerification(ctx context.Context, req *notificationv1.SendEmailVerificationRequest) (*emptypb.Empty, error) {
	return ackOrErr(s.uc.SendEmailVerification(ctx, req.To, req.UserId, req.VerifyToken))
}

func (s *Server) SendTrainDelayNotification(ctx context.Context, req *notificationv1.SendTrainDelayNotificationRequest) (*emptypb.Empty, error) {
	return ackOrErr(s.uc.SendTrainDelay(ctx, req.To, req.UserId, req.TrainId, req.DelayMinutes))
}

func (s *Server) SendTrainCancellationNotification(ctx context.Context, req *notificationv1.SendTrainCancellationNotificationRequest) (*emptypb.Empty, error) {
	return ackOrErr(s.uc.SendTrainCancellation(ctx, req.To, req.UserId, req.TrainId))
}

func (s *Server) GetNotificationById(ctx context.Context, req *notificationv1.GetNotificationByIdRequest) (*notificationv1.Notification, error) {
	n, err := s.uc.Get(ctx, req.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return toProto(n), nil
}

func (s *Server) ListUserNotifications(ctx context.Context, req *notificationv1.ListUserNotificationsRequest) (*notificationv1.ListUserNotificationsResponse, error) {
	list, err := s.uc.List(ctx, req.UserId)
	if err != nil {
		return nil, mapErr(err)
	}
	out := &notificationv1.ListUserNotificationsResponse{Items: make([]*notificationv1.Notification, 0, len(list))}
	for _, n := range list {
		out.Items = append(out.Items, toProto(n))
	}
	return out, nil
}

func (s *Server) MarkNotificationAsRead(ctx context.Context, req *notificationv1.MarkNotificationAsReadRequest) (*notificationv1.Notification, error) {
	n, err := s.uc.MarkRead(ctx, req.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return toProto(n), nil
}

func toProto(n *domain.Notification) *notificationv1.Notification {
	return &notificationv1.Notification{
		Id:        n.ID,
		UserId:    n.UserID,
		Kind:      n.Kind,
		Subject:   n.Subject,
		Body:      n.Body,
		Read:      n.Read,
		CreatedAt: timestamppb.New(n.CreatedAt),
	}
}

func ackOrErr(err error) (*emptypb.Empty, error) {
	if err != nil {
		return nil, mapErr(err)
	}
	return &emptypb.Empty{}, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}
