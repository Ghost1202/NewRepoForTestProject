package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/turtlepavlo/stripe_integration/internal/domain"
	"github.com/turtlepavlo/stripe_integration/pkg/producer"
	stripe "github.com/turtlepavlo/stripe_integration/pkg/stripe"
)

type paymentRepoMock struct {
	mu sync.Mutex

	getPaymentByOrderIDFn func(ctx context.Context, orderID uuid.UUID) (domain.Payment, error)
	createPaymentFn       func(ctx context.Context, payment domain.Payment) error
	updatePaymentStatusFn func(ctx context.Context, payment domain.Payment) (int64, error)

	getPaymentByExternalIDFn        func(ctx context.Context, externalID string) (domain.Payment, error)
	updatePaymentStatusByExternalID func(ctx context.Context, externalID, status string) (int64, error)

	addToWaitlistFn  func(ctx context.Context, entry domain.ToWaitlist) error
	getNextWaiterFn  func(ctx context.Context, eventID int64) (domain.WaitlistEntry, error)
	markAsNotifiedFn func(ctx context.Context, id int64) error

	getPaymentCalls         int
	createPaymentCalls      int
	updatePaymentCalls      int
	getPaymentByExtCalls    int
	updatePaymentByExtCalls int
	addToWaitlistCalls      int
	getNextWaiterCalls      int
	markAsNotifiedCalls     int

	lastGetPayment struct {
		orderID uuid.UUID
	}
	lastCreatePayment struct {
		payment domain.Payment
	}
	lastUpdatePayment struct {
		payment domain.Payment
	}
	lastUpdateByExternal struct {
		externalID string
		status     string
	}
}

func (m *paymentRepoMock) GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (domain.Payment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.getPaymentCalls++
	m.lastGetPayment.orderID = orderID

	if m.getPaymentByOrderIDFn == nil {
		return domain.Payment{}, errors.New("paymentRepoMock.GetPaymentByOrderID not configured")
	}
	return m.getPaymentByOrderIDFn(ctx, orderID)
}

func (m *paymentRepoMock) CreatePayment(ctx context.Context, payment domain.Payment) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.createPaymentCalls++
	m.lastCreatePayment.payment = payment

	if m.createPaymentFn == nil {
		return errors.New("paymentRepoMock.CreatePayment not configured")
	}
	return m.createPaymentFn(ctx, payment)
}

func (m *paymentRepoMock) UpdatePaymentStatus(ctx context.Context, payment domain.Payment) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.updatePaymentCalls++
	m.lastUpdatePayment.payment = payment

	if m.updatePaymentStatusFn == nil {
		return 0, errors.New("paymentRepoMock.UpdatePaymentStatus not configured")
	}
	return m.updatePaymentStatusFn(ctx, payment)
}

func (m *paymentRepoMock) GetPaymentByExternalID(ctx context.Context, externalID string) (domain.Payment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.getPaymentByExtCalls++
	if m.getPaymentByExternalIDFn == nil {
		return domain.Payment{}, errors.New("paymentRepoMock.GetPaymentByExternalID not configured")
	}
	return m.getPaymentByExternalIDFn(ctx, externalID)
}

func (m *paymentRepoMock) UpdatePaymentStatusByExternalID(ctx context.Context, externalID, status string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.updatePaymentByExtCalls++
	m.lastUpdateByExternal.externalID = externalID
	m.lastUpdateByExternal.status = status
	if m.updatePaymentStatusByExternalID == nil {
		return 0, errors.New("paymentRepoMock.UpdatePaymentStatusByExternalID not configured")
	}
	return m.updatePaymentStatusByExternalID(ctx, externalID, status)
}

func (m *paymentRepoMock) AddToWaitlist(ctx context.Context, entry domain.ToWaitlist) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.addToWaitlistCalls++
	if m.addToWaitlistFn == nil {
		return errors.New("paymentRepoMock.AddToWaitlist not configured")
	}
	return m.addToWaitlistFn(ctx, entry)
}

func (m *paymentRepoMock) GetNextWaiter(ctx context.Context, eventID int64) (domain.WaitlistEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.getNextWaiterCalls++
	if m.getNextWaiterFn == nil {
		return domain.WaitlistEntry{}, errors.New("paymentRepoMock.GetNextWaiter not configured")
	}
	return m.getNextWaiterFn(ctx, eventID)
}

func (m *paymentRepoMock) MarkAsNotified(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.markAsNotifiedCalls++
	if m.markAsNotifiedFn == nil {
		return errors.New("paymentRepoMock.MarkAsNotified not configured")
	}
	return m.markAsNotifiedFn(ctx, id)
}

type providerMock struct {
	mu sync.Mutex

	createCheckoutSessionFn func(ctx context.Context, params stripe.CheckoutSessionParams) (*stripe.CheckoutSessionResult, error)
	createRefundFn          func(ctx context.Context, params stripe.RefundParams) error

	createSessionCalls int
	createRefundCalls  int
	lastParams         stripe.CheckoutSessionParams
}

func (m *providerMock) CreateCheckoutSession(ctx context.Context, params stripe.CheckoutSessionParams) (*stripe.CheckoutSessionResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.createSessionCalls++
	m.lastParams = params

	if m.createCheckoutSessionFn == nil {
		return nil, errors.New("providerMock.CreateCheckoutSession not configured")
	}
	return m.createCheckoutSessionFn(ctx, params)
}

func (m *providerMock) CreateRefund(ctx context.Context, params stripe.RefundParams) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.createRefundCalls++
	if m.createRefundFn == nil {
		return errors.New("providerMock.CreateRefund not configured")
	}
	return m.createRefundFn(ctx, params)
}

type producerMock struct {
	mu sync.Mutex

	publishFn       func(ctx context.Context, event producer.PaymentConfirmedEvent) error
	publishRefundFn func(ctx context.Context, event producer.TicketRefundedEvent) error
	publishStatusFn func(ctx context.Context, event producer.PaymentStatusChangedEvent) error

	publishCalls       int
	publishRefundCalls int
	publishStatusCalls int
	lastEvent          producer.PaymentConfirmedEvent
	lastStatusEvent    producer.PaymentStatusChangedEvent
}

func (p *producerMock) PublishPaymentEvent(ctx context.Context, event producer.PaymentConfirmedEvent) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.publishCalls++
	p.lastEvent = event

	if p.publishFn == nil {
		return errors.New("producerMock.PublishPaymentEvent not configured")
	}
	return p.publishFn(ctx, event)
}

func (p *producerMock) PublishRefundEvent(ctx context.Context, event producer.TicketRefundedEvent) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.publishRefundCalls++
	if p.publishRefundFn == nil {
		return errors.New("producerMock.PublishRefundEvent not configured")
	}
	return p.publishRefundFn(ctx, event)
}

func (p *producerMock) PublishPaymentStatusEvent(ctx context.Context, event producer.PaymentStatusChangedEvent) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.publishStatusCalls++
	p.lastStatusEvent = event

	if p.publishStatusFn == nil {
		return errors.New("producerMock.PublishPaymentStatusEvent not configured")
	}
	return p.publishStatusFn(ctx, event)
}

type emailProviderMock struct {
	mu sync.Mutex

	sendEmailFn func(ctx context.Context, msg domain.Notification) error
	sendCalls   int
}

func (e *emailProviderMock) SendEmail(ctx context.Context, msg domain.Notification) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.sendCalls++
	if e.sendEmailFn == nil {
		return nil
	}
	return e.sendEmailFn(ctx, msg)
}

func TestPaymentService_CreatePaymentLink_Table(t *testing.T) {
	type tc struct {
		name string

		payment domain.CreatePaymentInput

		repoSetup     func(r *paymentRepoMock)
		providerSetup func(p *providerMock)

		wantURL        string
		wantErr        bool
		wantErrMsg     string
		wantRepoCreate int
		wantProvider   int
	}

	testOrderID := uuid.New()

	tests := []tc{
		{
			name: "save error (generic) -> returns error",
			payment: domain.CreatePaymentInput{
				OrderID:   testOrderID,
				Amount:    100,
				Currency:  "USD",
				TicketIDs: []int64{1},
			},
			repoSetup: func(r *paymentRepoMock) {
				r.getPaymentByOrderIDFn = func(ctx context.Context, orderID uuid.UUID) (domain.Payment, error) {
					return domain.Payment{}, pgx.ErrNoRows
				}
				r.createPaymentFn = func(ctx context.Context, payment domain.Payment) error {
					return errors.New("db connection failed")
				}
			},
			providerSetup:  func(p *providerMock) {},
			wantErr:        true,
			wantErrMsg:     "db connection failed",
			wantRepoCreate: 1,
			wantProvider:   0,
		},
		{
			name: "stripe error -> updates to FAILED and returns error",
			payment: domain.CreatePaymentInput{
				OrderID:   testOrderID,
				TicketIDs: []int64{1},
			},
			repoSetup: func(r *paymentRepoMock) {
				r.getPaymentByOrderIDFn = func(ctx context.Context, orderID uuid.UUID) (domain.Payment, error) {
					return domain.Payment{}, pgx.ErrNoRows
				}
				r.createPaymentFn = func(ctx context.Context, payment domain.Payment) error {
					return nil
				}
				r.updatePaymentStatusFn = func(ctx context.Context, payment domain.Payment) (int64, error) {
					if payment.Status == StatusFailed {
						return 1, nil
					}
					return 0, errors.New("unexpected update call")
				}
			},
			providerSetup: func(p *providerMock) {
				p.createCheckoutSessionFn = func(ctx context.Context, params stripe.CheckoutSessionParams) (*stripe.CheckoutSessionResult, error) {
					return nil, errors.New("stripe api down")
				}
			},
			wantErr:        true,
			wantErrMsg:     "stripe api down",
			wantRepoCreate: 1,
			wantProvider:   1,
		},
		{
			name: "success -> saves PENDING/COMPLETE and returns URL",
			payment: domain.CreatePaymentInput{
				OrderID:   testOrderID,
				TokenTTL:  time.Hour,
				TicketIDs: []int64{1},
			},
			repoSetup: func(r *paymentRepoMock) {
				r.getPaymentByOrderIDFn = func(ctx context.Context, orderID uuid.UUID) (domain.Payment, error) {
					return domain.Payment{}, pgx.ErrNoRows
				}
				r.createPaymentFn = func(ctx context.Context, payment domain.Payment) error {
					return nil
				}
				r.updatePaymentStatusFn = func(ctx context.Context, payment domain.Payment) (int64, error) {
					if payment.Status == StatusPending && payment.ExternalID == "sess_ok" {
						return 1, nil
					}
					if payment.Status == StatusFailed {
						return 0, errors.New("should not fail")
					}
					return 0, nil
				}
			},
			providerSetup: func(p *providerMock) {
				p.createCheckoutSessionFn = func(ctx context.Context, params stripe.CheckoutSessionParams) (*stripe.CheckoutSessionResult, error) {
					return &stripe.CheckoutSessionResult{
						URL: "http://success.url",
						ID:  "sess_ok",
					}, nil
				}
			},
			wantURL:        "http://success.url",
			wantErr:        false,
			wantRepoCreate: 1,
			wantProvider:   1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := &paymentRepoMock{}
			prov := &providerMock{}
			prod := &producerMock{}
			email := &emailProviderMock{}

			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			if tt.providerSetup != nil {
				tt.providerSetup(prov)
			}

			svc := NewPaymentService(Config{DefaultTokenTTL: 30 * time.Minute}, repo, prov, prod, email, zap.NewNop())

			paymentURL, _, err := svc.CreatePaymentLink(context.Background(), tt.payment)

			require.Equal(t, tt.wantRepoCreate, repo.createPaymentCalls)
			require.Equal(t, tt.wantProvider, prov.createSessionCalls)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMsg != "" {
					require.Contains(t, err.Error(), tt.wantErrMsg)
				}
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantURL, paymentURL)
		})
	}
}

func TestPaymentService_ConfirmPayment_Table(t *testing.T) {
	type tc struct {
		name string

		orderID   uuid.UUID
		sessionID string

		repoSetup     func(r *paymentRepoMock)
		producerSetup func(p *producerMock)

		wantErr bool

		wantGetCalls     int
		wantUpdateCalls  int
		wantPublishCalls int
	}

	testOrderID := uuid.New()
	testSessionID := "sess_confirmation"

	tests := []tc{
		{
			name:      "get payment error -> returns error",
			orderID:   testOrderID,
			sessionID: testSessionID,
			repoSetup: func(r *paymentRepoMock) {
				r.getPaymentByOrderIDFn = func(ctx context.Context, orderID uuid.UUID) (domain.Payment, error) {
					return domain.Payment{}, errors.New("db fail")
				}
			},
			producerSetup:    func(p *producerMock) {},
			wantErr:          true,
			wantGetCalls:     1,
			wantUpdateCalls:  0,
			wantPublishCalls: 0,
		},
		{
			name:      "already paid -> returns nil (idempotent), no updates/publish",
			orderID:   testOrderID,
			sessionID: testSessionID,
			repoSetup: func(r *paymentRepoMock) {
				r.getPaymentByOrderIDFn = func(ctx context.Context, orderID uuid.UUID) (domain.Payment, error) {
					return domain.Payment{ID: 1, Status: StatusPaid}, nil
				}
			},
			producerSetup:    func(p *producerMock) {},
			wantErr:          false,
			wantGetCalls:     1,
			wantUpdateCalls:  0,
			wantPublishCalls: 0,
		},
		{
			name:      "update to PAID error -> returns error",
			orderID:   testOrderID,
			sessionID: testSessionID,
			repoSetup: func(r *paymentRepoMock) {
				r.getPaymentByOrderIDFn = func(ctx context.Context, orderID uuid.UUID) (domain.Payment, error) {
					return domain.Payment{ID: 1, Status: StatusPending}, nil
				}
				r.updatePaymentStatusFn = func(ctx context.Context, payment domain.Payment) (int64, error) {
					if payment.Status == StatusPaid {
						return 0, errors.New("update failed")
					}
					return 0, nil
				}
			},
			producerSetup:    func(p *producerMock) {},
			wantErr:          true,
			wantGetCalls:     1,
			wantUpdateCalls:  1,
			wantPublishCalls: 0,
		},
		{
			name:      "publish error -> rollbacks status to FAILED and returns error",
			orderID:   testOrderID,
			sessionID: testSessionID,
			repoSetup: func(r *paymentRepoMock) {
				r.getPaymentByOrderIDFn = func(ctx context.Context, orderID uuid.UUID) (domain.Payment, error) {
					return domain.Payment{ID: 1, Status: StatusPending}, nil
				}
				r.updatePaymentStatusFn = func(ctx context.Context, payment domain.Payment) (int64, error) {
					return 1, nil
				}
			},
			producerSetup: func(p *producerMock) {
				p.publishFn = func(ctx context.Context, event producer.PaymentConfirmedEvent) error {
					return errors.New("kafka unreachable")
				}
			},
			wantErr:          true,
			wantGetCalls:     1,
			wantUpdateCalls:  2,
			wantPublishCalls: 1,
		},
		{
			name:      "success -> updates to PAID and publishes event",
			orderID:   testOrderID,
			sessionID: testSessionID,
			repoSetup: func(r *paymentRepoMock) {
				r.getPaymentByOrderIDFn = func(ctx context.Context, orderID uuid.UUID) (domain.Payment, error) {
					return domain.Payment{ID: 1, Status: StatusPending, UserID: 100}, nil
				}
				r.updatePaymentStatusFn = func(ctx context.Context, payment domain.Payment) (int64, error) {
					if payment.Status == StatusPaid && payment.ExternalID == testSessionID {
						return 1, nil
					}
					return 0, errors.New("unexpected status update params")
				}
			},
			producerSetup: func(p *producerMock) {
				p.publishFn = func(ctx context.Context, event producer.PaymentConfirmedEvent) error {
					if event.Status != StatusPaid {
						return errors.New("wrong event status")
					}
					return nil
				}
			},
			wantErr:          false,
			wantGetCalls:     1,
			wantUpdateCalls:  1,
			wantPublishCalls: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := &paymentRepoMock{}
			prov := &providerMock{}
			prod := &producerMock{}
			email := &emailProviderMock{}

			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			if tt.producerSetup != nil {
				tt.producerSetup(prod)
			}

			svc := NewPaymentService(Config{}, repo, prov, prod, email, zap.NewNop())

			err := svc.ConfirmPayment(context.Background(), tt.orderID, tt.sessionID)

			require.Equal(t, tt.wantGetCalls, repo.getPaymentCalls)
			if tt.wantGetCalls > 0 {
				require.Equal(t, tt.orderID, repo.lastGetPayment.orderID)
			}

			require.Equal(t, tt.wantUpdateCalls, repo.updatePaymentCalls)
			require.Equal(t, tt.wantPublishCalls, prod.publishCalls)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestPaymentService_CreateRefund_Table(t *testing.T) {
	type tc struct {
		name          string
		refund        domain.Refund
		providerSetup func(p *providerMock)
		wantErr       bool
		wantErrMsg    string
		wantCalls     int
	}

	tests := []tc{
		{
			name: "stripe error -> returns error",
			refund: domain.Refund{
				UserID: 1, EventID: 10, TicketID: 100, Amount: 500, Description: "duplicate",
			},
			providerSetup: func(p *providerMock) {
				p.createRefundFn = func(ctx context.Context, params stripe.RefundParams) error {
					return errors.New("stripe rejected refund")
				}
			},
			wantErr:    true,
			wantErrMsg: "stripe rejected refund",
			wantCalls:  1,
		},
		{
			name: "success -> calls provider and returns nil",
			refund: domain.Refund{
				UserID: 1, EventID: 10, TicketID: 100, Amount: 500,
			},
			providerSetup: func(p *providerMock) {
				p.createRefundFn = func(ctx context.Context, params stripe.RefundParams) error {
					if params.Amount != 500 || params.TicketID != 100 {
						return errors.New("wrong params passed to stripe")
					}
					return nil
				}
			},
			wantErr:   false,
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prov := &providerMock{}
			if tt.providerSetup != nil {
				tt.providerSetup(prov)
			}
			repo := &paymentRepoMock{}
			prod := &producerMock{}
			email := &emailProviderMock{}

			svc := NewPaymentService(Config{}, repo, prov, prod, email, zap.NewNop())

			err := svc.CreateRefund(context.Background(), tt.refund)

			require.Equal(t, tt.wantCalls, prov.createRefundCalls)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMsg != "" {
					require.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPaymentService_ConfirmRefund_Table(t *testing.T) {
	type tc struct {
		name string

		externalID string
		refund     domain.Refund

		repoSetup     func(r *paymentRepoMock)
		producerSetup func(p *producerMock)

		wantErr          bool
		wantGetCalls     int
		wantUpdateCalls  int
		wantPublishCalls int
	}

	testExtID := "re_12345"

	tests := []tc{
		{
			name:       "DB get error -> returns error",
			externalID: testExtID,
			refund:     domain.Refund{TicketID: 1},
			repoSetup: func(r *paymentRepoMock) {
				r.getPaymentByExternalIDFn = func(ctx context.Context, externalID string) (domain.Payment, error) {
					return domain.Payment{}, errors.New("db connection fail")
				}
			},
			producerSetup:    func(p *producerMock) {},
			wantErr:          true,
			wantGetCalls:     1,
			wantUpdateCalls:  0,
			wantPublishCalls: 0,
		},
		{
			name:       "Payment not found (ErrNoRows) -> skips update, publishes event (SUCCESS)",
			externalID: testExtID,
			refund:     domain.Refund{TicketID: 1},
			repoSetup: func(r *paymentRepoMock) {
				r.getPaymentByExternalIDFn = func(ctx context.Context, externalID string) (domain.Payment, error) {
					return domain.Payment{}, pgx.ErrNoRows
				}

				r.getNextWaiterFn = func(ctx context.Context, eventID int64) (domain.WaitlistEntry, error) {
					return domain.WaitlistEntry{}, pgx.ErrNoRows
				}
			},
			producerSetup: func(p *producerMock) {
				p.publishRefundFn = func(ctx context.Context, event producer.TicketRefundedEvent) error {
					return nil
				}
			},
			wantErr:          false,
			wantGetCalls:     1,
			wantUpdateCalls:  0,
			wantPublishCalls: 1,
		},
		{
			name:       "DB update error -> returns error",
			externalID: testExtID,
			refund:     domain.Refund{TicketID: 1},
			repoSetup: func(r *paymentRepoMock) {
				r.getPaymentByExternalIDFn = func(ctx context.Context, externalID string) (domain.Payment, error) {
					return domain.Payment{ID: 1}, nil
				}
				r.updatePaymentStatusByExternalID = func(ctx context.Context, externalID, status string) (int64, error) {
					return 0, errors.New("update failed")
				}
			},
			producerSetup:    func(p *producerMock) {},
			wantErr:          true,
			wantGetCalls:     1,
			wantUpdateCalls:  1,
			wantPublishCalls: 0,
		},
		{
			name:       "Kafka publish error -> returns error",
			externalID: testExtID,
			refund:     domain.Refund{TicketID: 1},
			repoSetup: func(r *paymentRepoMock) {
				r.getPaymentByExternalIDFn = func(ctx context.Context, externalID string) (domain.Payment, error) {
					return domain.Payment{ID: 1}, nil
				}
				r.updatePaymentStatusByExternalID = func(ctx context.Context, externalID, status string) (int64, error) {
					return 1, nil
				}
			},
			producerSetup: func(p *producerMock) {
				p.publishRefundFn = func(ctx context.Context, event producer.TicketRefundedEvent) error {
					return errors.New("kafka error")
				}
			},
			wantErr:          true,
			wantGetCalls:     1,
			wantUpdateCalls:  1,
			wantPublishCalls: 1,
		},
		{
			name:       "Success -> updates DB and publishes event",
			externalID: testExtID,
			refund:     domain.Refund{TicketID: 1, Amount: 100},
			repoSetup: func(r *paymentRepoMock) {
				r.getPaymentByExternalIDFn = func(ctx context.Context, externalID string) (domain.Payment, error) {
					if externalID != testExtID {
						return domain.Payment{}, pgx.ErrNoRows
					}
					return domain.Payment{ID: 1}, nil
				}
				r.updatePaymentStatusByExternalID = func(ctx context.Context, externalID, status string) (int64, error) {
					if status != StatusRefunded {
						return 0, errors.New("wrong status")
					}
					return 1, nil
				}
				r.getNextWaiterFn = func(ctx context.Context, eventID int64) (domain.WaitlistEntry, error) {
					return domain.WaitlistEntry{}, pgx.ErrNoRows
				}
			},
			producerSetup: func(p *producerMock) {
				p.publishRefundFn = func(ctx context.Context, event producer.TicketRefundedEvent) error {
					if event.Status != StatusRefunded {
						return errors.New("wrong event status")
					}
					return nil
				}
			},
			wantErr:          false,
			wantGetCalls:     1,
			wantUpdateCalls:  1,
			wantPublishCalls: 1,
		},
		{
			name:       "Empty ExternalID -> skips DB, publishes event",
			externalID: "",
			refund:     domain.Refund{TicketID: 1},
			repoSetup: func(r *paymentRepoMock) {
				r.getNextWaiterFn = func(ctx context.Context, eventID int64) (domain.WaitlistEntry, error) {
					return domain.WaitlistEntry{}, pgx.ErrNoRows
				}
			},
			producerSetup: func(p *producerMock) {
				p.publishRefundFn = func(ctx context.Context, event producer.TicketRefundedEvent) error {
					return nil
				}
			},
			wantErr:          false,
			wantGetCalls:     0,
			wantUpdateCalls:  0,
			wantPublishCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &paymentRepoMock{}
			prod := &producerMock{}
			prov := &providerMock{}
			email := &emailProviderMock{}

			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			if tt.producerSetup != nil {
				tt.producerSetup(prod)
			}

			if repo.getNextWaiterFn == nil {
				repo.getNextWaiterFn = func(ctx context.Context, eventID int64) (domain.WaitlistEntry, error) {
					return domain.WaitlistEntry{}, pgx.ErrNoRows
				}
			}

			svc := NewPaymentService(Config{}, repo, prov, prod, email, zap.NewNop())

			err := svc.ConfirmRefund(context.Background(), tt.externalID, tt.refund)

			require.Equal(t, tt.wantGetCalls, repo.getPaymentByExtCalls)
			require.Equal(t, tt.wantUpdateCalls, repo.updatePaymentByExtCalls)
			require.Equal(t, tt.wantPublishCalls, prod.publishRefundCalls)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
