package clients

import (
	"context"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	bookingv1 "github.com/azarenkov/ap2-final-gen/booking/v1"
	notificationv1 "github.com/azarenkov/ap2-final-gen/notification/v1"
	trainv1 "github.com/azarenkov/ap2-final-gen/train/v1"
	userv1 "github.com/azarenkov/ap2-final-gen/user/v1"
)

type Clients struct {
	Train        trainv1.TrainServiceClient
	User         userv1.UserServiceClient
	Booking      bookingv1.BookingServiceClient
	Notification notificationv1.NotificationServiceClient

	conns []*grpc.ClientConn
}

func Dial(ctx context.Context, trainAddr, userAddr, bookingAddr, notificationAddr string) (*Clients, error) {
	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	traceHandler := otelgrpc.NewClientHandler()
	tConn, err := grpc.DialContext(dialCtx, trainAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithStatsHandler(traceHandler),
	)
	if err != nil {
		return nil, err
	}

	uConn, _ := grpc.NewClient(userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithStatsHandler(traceHandler))
	bConn, _ := grpc.NewClient(bookingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithStatsHandler(traceHandler))
	nConn, _ := grpc.NewClient(notificationAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithStatsHandler(traceHandler))

	c := &Clients{
		Train:        trainv1.NewTrainServiceClient(tConn),
		User:         userv1.NewUserServiceClient(uConn),
		Booking:      bookingv1.NewBookingServiceClient(bConn),
		Notification: notificationv1.NewNotificationServiceClient(nConn),
		conns:        []*grpc.ClientConn{tConn, uConn, bConn, nConn},
	}
	return c, nil
}

func (c *Clients) Close() {
	for _, conn := range c.conns {
		if conn != nil {
			_ = conn.Close()
		}
	}
}
