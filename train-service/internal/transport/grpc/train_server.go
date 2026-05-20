package grpc

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	trainv1 "github.com/azarenkov/ap2-final-gen/train/v1"
	"train-service/internal/domain"
	"train-service/internal/usecase"
)

type TrainServer struct {
	trainv1.UnimplementedTrainServiceServer
	uc *usecase.TrainUseCase
}

func NewTrainServer(uc *usecase.TrainUseCase) *TrainServer {
	return &TrainServer{uc: uc}
}

func (s *TrainServer) CreateTrain(ctx context.Context, req *trainv1.CreateTrainRequest) (*trainv1.Train, error) {
	if req.Code == "" || req.RouteId == "" {
		return nil, status.Error(codes.InvalidArgument, "code and route_id are required")
	}
	t, err := s.uc.CreateTrain(ctx,
		req.Code, req.Name, req.RouteId,
		ts(req.DepartureTime), ts(req.ArrivalTime),
		req.TotalSeats, req.PriceCents,
	)
	if err != nil {
		return nil, mapErr(err)
	}
	return toProto(t), nil
}

func (s *TrainServer) GetTrainById(ctx context.Context, req *trainv1.GetTrainByIdRequest) (*trainv1.Train, error) {
	t, err := s.uc.GetTrainByID(ctx, req.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return toProto(t), nil
}

func (s *TrainServer) UpdateTrain(ctx context.Context, req *trainv1.UpdateTrainRequest) (*trainv1.Train, error) {
	t, err := s.uc.UpdateTrain(ctx, req.Id, req.Name,
		ts(req.DepartureTime), ts(req.ArrivalTime), req.PriceCents, req.Status,
	)
	if err != nil {
		return nil, mapErr(err)
	}
	return toProto(t), nil
}

func (s *TrainServer) DeleteTrain(ctx context.Context, req *trainv1.DeleteTrainRequest) (*emptypb.Empty, error) {
	if err := s.uc.DeleteTrain(ctx, req.Id); err != nil {
		return nil, mapErr(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *TrainServer) SearchTrains(ctx context.Context, req *trainv1.SearchTrainsRequest) (*trainv1.SearchTrainsResponse, error) {
	filter := &domain.SearchFilter{
		Origin:      req.Origin,
		Destination: req.Destination,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}
	if a := ts(req.DepartureAfter); !a.IsZero() {
		filter.DepartureAfter = &a
	}
	if b := ts(req.DepartureBefore); !b.IsZero() {
		filter.DepartureBefore = &b
	}
	trains, total, err := s.uc.SearchTrains(ctx, filter)
	if err != nil {
		return nil, mapErr(err)
	}
	out := &trainv1.SearchTrainsResponse{Total: total, Trains: make([]*trainv1.Train, 0, len(trains))}
	for _, t := range trains {
		out.Trains = append(out.Trains, toProto(t))
	}
	return out, nil
}

func (s *TrainServer) GetTrainSchedule(ctx context.Context, req *trainv1.GetTrainScheduleRequest) (*trainv1.TrainSchedule, error) {
	t, r, err := s.uc.GetTrainSchedule(ctx, req.TrainId)
	if err != nil {
		return nil, mapErr(err)
	}
	mid := t.DepartureTime.Add(t.ArrivalTime.Sub(t.DepartureTime) / 2)
	return &trainv1.TrainSchedule{
		TrainId: t.ID,
		Stops: []*trainv1.ScheduleStop{
			{Station: r.Origin, Departure: timestamppb.New(t.DepartureTime)},
			{Station: "transit", Arrival: timestamppb.New(mid), Departure: timestamppb.New(mid.Add(2 * time.Minute))},
			{Station: r.Destination, Arrival: timestamppb.New(t.ArrivalTime)},
		},
	}, nil
}

func (s *TrainServer) CreateRoute(ctx context.Context, req *trainv1.CreateRouteRequest) (*trainv1.Route, error) {
	r, err := s.uc.CreateRoute(ctx, req.Origin, req.Destination, req.DistanceKm, req.EstimatedMinutes)
	if err != nil {
		return nil, mapErr(err)
	}
	return routeToProto(r), nil
}

func (s *TrainServer) GetRouteById(ctx context.Context, req *trainv1.GetRouteByIdRequest) (*trainv1.Route, error) {
	r, err := s.uc.GetRouteByID(ctx, req.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return routeToProto(r), nil
}

func (s *TrainServer) UpdateRoute(ctx context.Context, req *trainv1.UpdateRouteRequest) (*trainv1.Route, error) {
	r, err := s.uc.UpdateRoute(ctx, req.Id, req.DistanceKm, req.EstimatedMinutes)
	if err != nil {
		return nil, mapErr(err)
	}
	return routeToProto(r), nil
}

func (s *TrainServer) DeleteRoute(ctx context.Context, req *trainv1.DeleteRouteRequest) (*emptypb.Empty, error) {
	if err := s.uc.DeleteRoute(ctx, req.Id); err != nil {
		return nil, mapErr(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *TrainServer) GetAvailableSeats(ctx context.Context, req *trainv1.GetAvailableSeatsRequest) (*trainv1.AvailableSeatsResponse, error) {
	total, available, err := s.uc.GetAvailableSeats(ctx, req.TrainId)
	if err != nil {
		return nil, mapErr(err)
	}
	return &trainv1.AvailableSeatsResponse{TrainId: req.TrainId, TotalSeats: total, AvailableSeats: available}, nil
}

func (s *TrainServer) UpdateSeatAvailability(ctx context.Context, req *trainv1.UpdateSeatAvailabilityRequest) (*trainv1.UpdateSeatAvailabilityResponse, error) {
	available, err := s.uc.UpdateSeatAvailability(ctx, req.TrainId, req.Delta)
	if err != nil {
		return nil, mapErr(err)
	}
	return &trainv1.UpdateSeatAvailabilityResponse{TrainId: req.TrainId, AvailableSeats: available}, nil
}

func toProto(t *domain.Train) *trainv1.Train {
	return &trainv1.Train{
		Id:             t.ID,
		Code:           t.Code,
		Name:           t.Name,
		RouteId:        t.RouteID,
		DepartureTime:  timestamppb.New(t.DepartureTime),
		ArrivalTime:    timestamppb.New(t.ArrivalTime),
		TotalSeats:     t.TotalSeats,
		AvailableSeats: t.AvailableSeats,
		PriceCents:     t.PriceCents,
		Status:         t.Status,
		CreatedAt:      timestamppb.New(t.CreatedAt),
		UpdatedAt:      timestamppb.New(t.UpdatedAt),
	}
}

func routeToProto(r *domain.Route) *trainv1.Route {
	return &trainv1.Route{
		Id:               r.ID,
		Origin:           r.Origin,
		Destination:      r.Destination,
		DistanceKm:       r.DistanceKm,
		EstimatedMinutes: r.EstimatedMinutes,
	}
}

func ts(t *timestamppb.Timestamp) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.AsTime().UTC()
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrTrainNotFound),
		errors.Is(err, domain.ErrRouteNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidSeats),
		errors.Is(err, domain.ErrInvalidPrice),
		errors.Is(err, domain.ErrInvalidTimes),
		errors.Is(err, domain.ErrInvalidStatus),
		errors.Is(err, domain.ErrInvalidRouteFields),
		errors.Is(err, domain.ErrSeatDeltaTooLarge):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrNotEnoughSeats):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}
