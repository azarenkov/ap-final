package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	trainv1 "github.com/azarenkov/ap2-final-gen/train/v1"
)

type TrainHandler struct {
	client trainv1.TrainServiceClient
}

func NewTrainHandler(c trainv1.TrainServiceClient) *TrainHandler {
	return &TrainHandler{client: c}
}

func (h *TrainHandler) RegisterPublic(rg *gin.RouterGroup) {
	rg.GET("/trains", h.search)
	rg.GET("/trains/:id", h.get)
	rg.GET("/trains/:id/schedule", h.schedule)
	rg.GET("/trains/:id/seats", h.seats)
	rg.GET("/routes/:id", h.getRoute)
}

func (h *TrainHandler) RegisterAuthenticated(rg *gin.RouterGroup) {
	rg.POST("/trains", h.create)
	rg.PATCH("/trains/:id", h.update)
	rg.DELETE("/trains/:id", h.delete)
	rg.POST("/trains/:id/seats", h.updateSeats)

	rg.POST("/routes", h.createRoute)
	rg.PATCH("/routes/:id", h.updateRoute)
	rg.DELETE("/routes/:id", h.deleteRoute)
}

type createTrainBody struct {
	Code          string    `json:"code" binding:"required"`
	Name          string    `json:"name" binding:"required"`
	RouteID       string    `json:"route_id" binding:"required"`
	DepartureTime time.Time `json:"departure_time" binding:"required"`
	ArrivalTime   time.Time `json:"arrival_time" binding:"required"`
	TotalSeats    int32     `json:"total_seats" binding:"required"`
	PriceCents    int64     `json:"price_cents" binding:"required"`
}

func (h *TrainHandler) create(c *gin.Context) {
	var body createTrainBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	out, err := h.client.CreateTrain(c, &trainv1.CreateTrainRequest{
		Code:          body.Code,
		Name:          body.Name,
		RouteId:       body.RouteID,
		DepartureTime: timestamppb.New(body.DepartureTime),
		ArrivalTime:   timestamppb.New(body.ArrivalTime),
		TotalSeats:    body.TotalSeats,
		PriceCents:    body.PriceCents,
	})
	respond(c, out, err, http.StatusCreated)
}

func (h *TrainHandler) get(c *gin.Context) {
	out, err := h.client.GetTrainById(c, &trainv1.GetTrainByIdRequest{Id: c.Param("id")})
	respond(c, out, err, http.StatusOK)
}

type updateTrainBody struct {
	Name          string    `json:"name"`
	DepartureTime time.Time `json:"departure_time"`
	ArrivalTime   time.Time `json:"arrival_time"`
	PriceCents    int64     `json:"price_cents"`
	Status        string    `json:"status"`
}

func (h *TrainHandler) update(c *gin.Context) {
	var body updateTrainBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req := &trainv1.UpdateTrainRequest{
		Id:         c.Param("id"),
		Name:       body.Name,
		PriceCents: body.PriceCents,
		Status:     body.Status,
	}
	if !body.DepartureTime.IsZero() {
		req.DepartureTime = timestamppb.New(body.DepartureTime)
	}
	if !body.ArrivalTime.IsZero() {
		req.ArrivalTime = timestamppb.New(body.ArrivalTime)
	}
	out, err := h.client.UpdateTrain(c, req)
	respond(c, out, err, http.StatusOK)
}

func (h *TrainHandler) delete(c *gin.Context) {
	_, err := h.client.DeleteTrain(c, &trainv1.DeleteTrainRequest{Id: c.Param("id")})
	if err != nil {
		writeGrpcErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *TrainHandler) search(c *gin.Context) {
	q := c.Request.URL.Query()
	req := &trainv1.SearchTrainsRequest{
		Origin:      q.Get("origin"),
		Destination: q.Get("destination"),
	}
	if v := q.Get("page"); v != "" {
		n, _ := strconv.Atoi(v)
		req.Page = int32(n)
	}
	if v := q.Get("page_size"); v != "" {
		n, _ := strconv.Atoi(v)
		req.PageSize = int32(n)
	}
	if v := q.Get("after"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			req.DepartureAfter = timestamppb.New(t)
		}
	}
	if v := q.Get("before"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			req.DepartureBefore = timestamppb.New(t)
		}
	}
	out, err := h.client.SearchTrains(c, req)
	respond(c, out, err, http.StatusOK)
}

func (h *TrainHandler) schedule(c *gin.Context) {
	out, err := h.client.GetTrainSchedule(c, &trainv1.GetTrainScheduleRequest{TrainId: c.Param("id")})
	respond(c, out, err, http.StatusOK)
}

func (h *TrainHandler) seats(c *gin.Context) {
	out, err := h.client.GetAvailableSeats(c, &trainv1.GetAvailableSeatsRequest{TrainId: c.Param("id")})
	respond(c, out, err, http.StatusOK)
}

type seatDeltaBody struct {
	Delta int32 `json:"delta" binding:"required"`
}

func (h *TrainHandler) updateSeats(c *gin.Context) {
	var body seatDeltaBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	out, err := h.client.UpdateSeatAvailability(c, &trainv1.UpdateSeatAvailabilityRequest{
		TrainId: c.Param("id"),
		Delta:   body.Delta,
	})
	respond(c, out, err, http.StatusOK)
}

type routeBody struct {
	Origin           string `json:"origin"`
	Destination      string `json:"destination"`
	DistanceKm       int32  `json:"distance_km"`
	EstimatedMinutes int32  `json:"estimated_minutes"`
}

func (h *TrainHandler) createRoute(c *gin.Context) {
	var body routeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	out, err := h.client.CreateRoute(c, &trainv1.CreateRouteRequest{
		Origin:           body.Origin,
		Destination:      body.Destination,
		DistanceKm:       body.DistanceKm,
		EstimatedMinutes: body.EstimatedMinutes,
	})
	respond(c, out, err, http.StatusCreated)
}

func (h *TrainHandler) getRoute(c *gin.Context) {
	out, err := h.client.GetRouteById(c, &trainv1.GetRouteByIdRequest{Id: c.Param("id")})
	respond(c, out, err, http.StatusOK)
}

func (h *TrainHandler) updateRoute(c *gin.Context) {
	var body routeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	out, err := h.client.UpdateRoute(c, &trainv1.UpdateRouteRequest{
		Id:               c.Param("id"),
		DistanceKm:       body.DistanceKm,
		EstimatedMinutes: body.EstimatedMinutes,
	})
	respond(c, out, err, http.StatusOK)
}

func (h *TrainHandler) deleteRoute(c *gin.Context) {
	_, err := h.client.DeleteRoute(c, &trainv1.DeleteRouteRequest{Id: c.Param("id")})
	if err != nil {
		writeGrpcErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func respond(c *gin.Context, payload any, err error, ok int) {
	if err != nil {
		writeGrpcErr(c, err)
		return
	}
	c.JSON(ok, payload)
}

func writeGrpcErr(c *gin.Context, err error) {
	s, ok := status.FromError(err)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	httpCode := http.StatusInternalServerError
	switch s.Code() {
	case codes.NotFound:
		httpCode = http.StatusNotFound
	case codes.InvalidArgument:
		httpCode = http.StatusBadRequest
	case codes.FailedPrecondition:
		httpCode = http.StatusConflict
	case codes.Unauthenticated:
		httpCode = http.StatusUnauthorized
	case codes.PermissionDenied:
		httpCode = http.StatusForbidden
	case codes.Unavailable:
		httpCode = http.StatusServiceUnavailable
	}
	c.JSON(httpCode, gin.H{"error": s.Message()})
}
