package service

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/internal/storage"
	"github.com/turtlepavlo/event-booking/pkg/producer"
)

const (
	StatusPaid       = "PAID"
	StatusCreated    = "CREATED"
	CurrencyUSD      = "USD"
	PromoTypePercent = "PERCENT"
	PromoTypeFixed   = "FIXED"
)

const (
	DiscountTypePromo  = "PROMO"
	DiscountTypeEarly  = "EARLY"
	DiscountTypeBundle = "BUNDLE"
	DiscountTypeNone   = "NONE"
)

type TicketRepository interface {
	GetTicketByID(ctx context.Context, ticketID int64) (storage.TicketModel, error)
	GetTicketsByEvent(ctx context.Context, eventID int64) ([]storage.TicketModel, error)
	GetTicketsByIDs(ctx context.Context, ticketIDs []int64) ([]storage.TicketModel, error)
	GetTicketsByUserID(ctx context.Context, userID int64) ([]storage.TicketModel, error)
	GetPromoByCode(ctx context.Context, eventID int64, code string) (storage.Promo, error)
	GetEarlyByEvent(ctx context.Context, eventID int64) ([]storage.Early, error)
	GetBundleByEvent(ctx context.Context, eventID int64) ([]storage.Bundle, error)
	GetEarlyByID(ctx context.Context, earlyID int64) (storage.Early, error)
	GetBundleByID(ctx context.Context, bundleID int64) (storage.Bundle, error)
}

type CacheRepository interface {
	Lock(ctx context.Context, ticketID int64, userID int64, ttl time.Duration) (bool, error)
	MGet(ctx context.Context, keys []string) ([]interface{}, error)
	Unlock(ctx context.Context, ticketIDs []int64) error
}

type PaymentClient interface {
	GetPaymentLink(ctx context.Context, booking domain.PaymentInput) (string, error)
	CreateRefund(ctx context.Context, refund domain.RefundPayment) error
	AddToWaitlist(ctx context.Context, item domain.Waitlist) error
}

type WalletClient interface {
	ChargeWallet(ctx context.Context, input domain.WalletCharge) (domain.WalletChargeStatus, error)
}

type Producer interface {
	TicketTransfer(ctx context.Context, event producer.TicketTransfer) error
}

type BookingService struct {
	log         *zap.Logger
	tickets     TicketRepository
	cache       CacheRepository
	payment     PaymentClient
	wallet      WalletClient
	fromStorage *FromStorageConvert
	toStorage   *ToStorageConvert
	toProducer  *ToProducerConvert
	toPayment   *ToPaymentConvert
	toWallet    *ToWalletConvert
	cfg         Config
	producer    Producer
}

func New(
	tickets TicketRepository,
	cache CacheRepository,
	payment PaymentClient,
	wallet WalletClient,
	log *zap.Logger,
	cfg Config,
	producer Producer,
) *BookingService {
	return &BookingService{
		log:         log,
		tickets:     tickets,
		cache:       cache,
		payment:     payment,
		wallet:      wallet,
		fromStorage: NewFromStorageConvert(),
		toStorage:   NewToStorageConvert(),
		toProducer:  NewToProducerConvert(),
		toPayment:   NewToPaymentConvert(),
		toWallet:    NewToWalletConvert(),
		cfg:         cfg,
		producer:    producer,
	}
}
