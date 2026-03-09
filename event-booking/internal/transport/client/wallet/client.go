package wallet

import (
	"net"
	"strconv"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	walletv1 "github.com/turtlepavlo/proto-contract/gen/go/wallet/v1"
)

type Client struct {
	api         walletv1.WalletServiceClient
	conn        *grpc.ClientConn
	log         *zap.Logger
	converter   protoConverter
	callTimeout time.Duration
}

func NewWalletClient(cfg Config, log *zap.Logger) (*Client, error) {
	addr := net.JoinHostPort(cfg.Host, strconv.FormatInt(int64(cfg.Port), 10))
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{MinConnectTimeout: cfg.DialTimeout}),
	)
	if err != nil {
		return nil, err
	}

	api := walletv1.NewWalletServiceClient(conn)

	return &Client{
		api:         api,
		conn:        conn,
		log:         log,
		converter:   newProtoConverter(),
		callTimeout: cfg.CallTimeout,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
