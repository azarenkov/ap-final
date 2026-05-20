package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"api-gateway/internal/middleware"
	bookingv1 "github.com/azarenkov/ap2-final-gen/booking/v1"
)

type BookingHandler struct {
	client bookingv1.BookingServiceClient
}

func NewBookingHandler(c bookingv1.BookingServiceClient) *BookingHandler {
	return &BookingHandler{client: c}
}

func (h *BookingHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/bookings", h.create)
	rg.GET("/bookings/me", h.mine)
	rg.GET("/bookings/:id", h.get)
	rg.POST("/bookings/:id/cancel", h.cancel)
	rg.POST("/bookings/:id/confirm", h.confirm)
	rg.GET("/bookings/:id/status", h.statusFor)
	rg.POST("/bookings/:id/pay", h.pay)
	rg.POST("/bookings/:id/ticket", h.ticket)
	rg.POST("/bookings/:id/refund", h.refund)
}

func (h *BookingHandler) pay(c *gin.Context) {
	ctx, cancel := withTimeout(c, 5*time.Second)
	defer cancel()
	out, err := h.client.ProcessPaymentMock(ctx, &bookingv1.ProcessPaymentMockRequest{BookingId: c.Param("id")})
	respond(c, out, err, http.StatusOK)
}

func (h *BookingHandler) ticket(c *gin.Context) {
	ctx, cancel := withTimeout(c, 5*time.Second)
	defer cancel()
	out, err := h.client.GenerateTicket(ctx, &bookingv1.GenerateTicketRequest{BookingId: c.Param("id")})
	respond(c, out, err, http.StatusOK)
}

func (h *BookingHandler) refund(c *gin.Context) {
	ctx, cancel := withTimeout(c, 5*time.Second)
	defer cancel()
	out, err := h.client.RefundBooking(ctx, &bookingv1.RefundBookingRequest{Id: c.Param("id")})
	respond(c, out, err, http.StatusOK)
}

type createBookingBody struct {
	TrainID   string `json:"train_id" binding:"required"`
	SeatCount int32  `json:"seat_count" binding:"required,min=1"`
}

func (h *BookingHandler) create(c *gin.Context) {
	var body createBookingBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get(middleware.ContextUserID)
	ctx, cancel := withTimeout(c, 5*time.Second)
	defer cancel()
	out, err := h.client.CreateBooking(ctx, &bookingv1.CreateBookingRequest{
		UserId:    userID.(string),
		TrainId:   body.TrainID,
		SeatCount: body.SeatCount,
	})
	respond(c, out, err, http.StatusCreated)
}

func (h *BookingHandler) get(c *gin.Context) {
	ctx, cancel := withTimeout(c, 5*time.Second)
	defer cancel()
	out, err := h.client.GetBookingById(ctx, &bookingv1.GetBookingByIdRequest{Id: c.Param("id")})
	respond(c, out, err, http.StatusOK)
}

func (h *BookingHandler) mine(c *gin.Context) {
	userID, _ := c.Get(middleware.ContextUserID)
	ctx, cancel := withTimeout(c, 5*time.Second)
	defer cancel()
	out, err := h.client.GetUserBookings(ctx, &bookingv1.GetUserBookingsRequest{UserId: userID.(string)})
	respond(c, out, err, http.StatusOK)
}

func (h *BookingHandler) cancel(c *gin.Context) {
	ctx, cancel := withTimeout(c, 5*time.Second)
	defer cancel()
	out, err := h.client.CancelBooking(ctx, &bookingv1.CancelBookingRequest{Id: c.Param("id")})
	respond(c, out, err, http.StatusOK)
}

func (h *BookingHandler) confirm(c *gin.Context) {
	ctx, cancel := withTimeout(c, 5*time.Second)
	defer cancel()
	out, err := h.client.ConfirmBooking(ctx, &bookingv1.ConfirmBookingRequest{Id: c.Param("id")})
	respond(c, out, err, http.StatusOK)
}

func (h *BookingHandler) statusFor(c *gin.Context) {
	ctx, cancel := withTimeout(c, 5*time.Second)
	defer cancel()
	out, err := h.client.CheckBookingStatus(ctx, &bookingv1.CheckBookingStatusRequest{Id: c.Param("id")})
	respond(c, out, err, http.StatusOK)
}

func withTimeout(c *gin.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), d)
}

func bearerToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}
