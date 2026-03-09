package service

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/turtlepavlo/stripe_integration/internal/domain"
	"github.com/turtlepavlo/stripe_integration/pkg/producer"
	stripe "github.com/turtlepavlo/stripe_integration/pkg/stripe"
)

const (
	StatusProcessing    = "PROCESSING"
	StatusPending       = "PENDING"
	StatusPaid          = "PAID"
	StatusFailed        = "FAILED"
	StatusRefundPending = "REFUND_PENDING"
	StatusRefunded      = "REFUND"
	StatusSucceeded     = "SUCCEEDED"
	DefaultCurrencyUSD  = "USD"
)

type PaymentRepository interface {
	GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (domain.Payment, error)
	CreatePayment(ctx context.Context, payment domain.Payment) error
	UpdatePaymentStatus(ctx context.Context, payment domain.Payment) (int64, error)

	GetPaymentByExternalID(ctx context.Context, externalID string) (domain.Payment, error)
	UpdatePaymentStatusByExternalID(ctx context.Context, externalID, status string) (int64, error)

	AddToWaitlist(ctx context.Context, entry domain.ToWaitlist) error
	GetNextWaiter(ctx context.Context, eventID int64) (domain.WaitlistEntry, error)
	MarkAsNotified(ctx context.Context, id int64) error
}

type PaymentProvider interface {
	CreateCheckoutSession(ctx context.Context, params stripe.CheckoutSessionParams) (*stripe.CheckoutSessionResult, error)
	CreateRefund(ctx context.Context, params stripe.RefundParams) error
}

type Producer interface {
	PublishPaymentEvent(ctx context.Context, event producer.PaymentConfirmedEvent) error
	PublishRefundEvent(ctx context.Context, event producer.TicketRefundedEvent) error
	PublishPaymentStatusEvent(ctx context.Context, event producer.PaymentStatusChangedEvent) error
}

type EmailProvider interface {
	SendEmail(ctx context.Context, msg domain.Notification) error
}

type PaymentService struct {
	cfg       Config
	repo      PaymentRepository
	provider  PaymentProvider
	producer  Producer
	log       *zap.Logger
	toStorage *ToStorageConvert
	toKafka   *ToKafkaConvert
	toStripe  *ToStripeConvert
	email     EmailProvider
}

func NewPaymentService(cfg Config, repo PaymentRepository, provider PaymentProvider, producer Producer, email EmailProvider, log *zap.Logger) *PaymentService {
	return &PaymentService{
		cfg:       cfg,
		repo:      repo,
		provider:  provider,
		producer:  producer,
		email:     email,
		log:       log,
		toStorage: NewToStorageConvert(),
		toKafka:   NewToKafkaConvert(),
		toStripe:  NewToStripeConvert(),
	}
}
