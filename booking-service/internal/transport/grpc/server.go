package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"booking-service/internal/domain"
	"booking-service/internal/usecase"
	bookingv1 "github.com/azarenkov/ap2-final-gen/booking/v1"
)

type Server struct {
	bookingv1.UnimplementedBookingServiceServer
	uc *usecase.BookingUseCase
}

func NewServer(uc *usecase.BookingUseCase) *Server { return &Server{uc: uc} }

func (s *Server) CreateBooking(ctx context.Context, req *bookingv1.CreateBookingRequest) (*bookingv1.Booking, error) {
	b, err := s.uc.CreateBooking(ctx, req.UserId, req.TrainId, req.SeatCount)
	if err != nil {
		return nil, mapErr(err)
	}
	return toProto(b), nil
}

func (s *Server) GetBookingById(ctx context.Context, req *bookingv1.GetBookingByIdRequest) (*bookingv1.Booking, error) {
	b, err := s.uc.Get(ctx, req.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return toProto(b), nil
}

func (s *Server) GetUserBookings(ctx context.Context, req *bookingv1.GetUserBookingsRequest) (*bookingv1.UserBookings, error) {
	list, err := s.uc.ListByUser(ctx, req.UserId)
	if err != nil {
		return nil, mapErr(err)
	}
	return toListProto(list), nil
}

func (s *Server) ListBookings(ctx context.Context, req *bookingv1.ListBookingsRequest) (*bookingv1.UserBookings, error) {
	list, err := s.uc.ListPage(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, mapErr(err)
	}
	return toListProto(list), nil
}

func (s *Server) CancelBooking(ctx context.Context, req *bookingv1.CancelBookingRequest) (*bookingv1.Booking, error) {
	b, err := s.uc.Cancel(ctx, req.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return toProto(b), nil
}

func (s *Server) ConfirmBooking(ctx context.Context, req *bookingv1.ConfirmBookingRequest) (*bookingv1.Booking, error) {
	b, err := s.uc.Confirm(ctx, req.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return toProto(b), nil
}

func (s *Server) CheckBookingStatus(ctx context.Context, req *bookingv1.CheckBookingStatusRequest) (*bookingv1.BookingStatus, error) {
	st, err := s.uc.Status(ctx, req.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return &bookingv1.BookingStatus{Id: req.Id, Status: st}, nil
}

func (s *Server) ReserveSeat(ctx context.Context, req *bookingv1.ReserveSeatRequest) (*bookingv1.Booking, error) {

	b, err := s.uc.ReserveSeat(ctx, "system", req.TrainId, req.SeatCount)
	if err != nil {
		return nil, mapErr(err)
	}
	return toProto(b), nil
}

func (s *Server) ReleaseSeat(ctx context.Context, req *bookingv1.ReleaseSeatRequest) (*bookingv1.Booking, error) {
	b, err := s.uc.ReleaseSeat(ctx, req.BookingId)
	if err != nil {
		return nil, mapErr(err)
	}
	return toProto(b), nil
}

func (s *Server) ProcessPaymentMock(ctx context.Context, req *bookingv1.ProcessPaymentMockRequest) (*bookingv1.PaymentResult, error) {
	ok, msg, err := s.uc.ProcessPaymentMock(ctx, req.BookingId)
	if err != nil {
		return nil, mapErr(err)
	}
	return &bookingv1.PaymentResult{BookingId: req.BookingId, Success: ok, Message: msg}, nil
}

func (s *Server) GenerateTicket(ctx context.Context, req *bookingv1.GenerateTicketRequest) (*bookingv1.Ticket, error) {
	id, code, issued, err := s.uc.GenerateTicket(ctx, req.BookingId)
	if err != nil {
		return nil, mapErr(err)
	}
	return &bookingv1.Ticket{Id: id, BookingId: req.BookingId, Code: code, IssuedAt: timestamppb.New(issued)}, nil
}

func (s *Server) RefundBooking(ctx context.Context, req *bookingv1.RefundBookingRequest) (*bookingv1.RefundResult, error) {
	amount, err := s.uc.Refund(ctx, req.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return &bookingv1.RefundResult{BookingId: req.Id, AmountCents: amount}, nil
}

func toProto(b *domain.Booking) *bookingv1.Booking {
	return &bookingv1.Booking{
		Id:          b.ID,
		UserId:      b.UserID,
		TrainId:     b.TrainID,
		SeatCount:   b.SeatCount,
		AmountCents: b.AmountCents,
		Status:      b.Status,
		CreatedAt:   timestamppb.New(b.CreatedAt),
		UpdatedAt:   timestamppb.New(b.UpdatedAt),
	}
}

func toListProto(list []*domain.Booking) *bookingv1.UserBookings {
	out := &bookingv1.UserBookings{Bookings: make([]*bookingv1.Booking, 0, len(list))}
	for _, b := range list {
		out.Bookings = append(out.Bookings, toProto(b))
	}
	return out
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrBookingNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidSeatCount):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrIllegalTransition):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrTrainUnavailable):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}
