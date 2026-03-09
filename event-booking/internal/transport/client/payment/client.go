package client

import (
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	paymentv1 "github.com/turtlepavlo/proto-contract/gen/go/payment/v1"
)

const (
	OrderDescriptionFormat = "Payment for order %s"
	RefundDescription      = "User requested refund"
)

type PaymentClient struct {
	api       paymentv1.PaymentServiceClient
	conn      *grpc.ClientConn
	log       *zap.Logger
	converter protoConverter
}

func NewPaymentClient(addr string, tokenTTL time.Duration, log *zap.Logger) (*PaymentClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	api := paymentv1.NewPaymentServiceClient(conn)

	return &PaymentClient{
		api:       api,
		conn:      conn,
		log:       log,
		converter: newProtoConverter(tokenTTL),
	}, nil
}

func (c *PaymentClient) Close() error {
	return c.conn.Close()
}
