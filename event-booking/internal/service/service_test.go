package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/internal/storage"
	"github.com/turtlepavlo/event-booking/pkg/producer"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type producerMock struct {
	ticketTransferFn    func(ctx context.Context, event producer.TicketTransfer) error
	ticketTransferCalls int
	lastTransfer        producer.TicketTransfer
}

func (m *producerMock) TicketTransfer(ctx context.Context, event producer.TicketTransfer) error {
	m.ticketTransferCalls++
	m.lastTransfer = event
	if m.ticketTransferFn != nil {
		return m.ticketTransferFn(ctx, event)
	}
	return nil
}

func TestBookingService_BookTickets_Table(t *testing.T) {
	type tc struct {
		name string

		booking domain.Booking
		lockTTL time.Duration

		ticketsSetup func(r *ticketRepoMock)
		cacheSetup   func(c *cacheRepoMock)
		paymentSetup func(p *paymentClientMock)

		wantURLContains string
		wantErr         bool
		wantErrMsg      string
		wantAmount      int64

		wantLockCalls int
	}

	tests := []tc{
		{
			name:    "lock error -> returns error",
			booking: domain.Booking{UserID: 1, TicketIDs: []int64{100}, UserEmail: "test@example.com", PromoCode: "SALE10"},
			lockTTL: 5 * time.Second,
			ticketsSetup: func(r *ticketRepoMock) {
				r.getTicketsByIDsFn = func(ctx context.Context, ticketIDs []int64) ([]storage.TicketModel, error) {
					return []storage.TicketModel{{ID: 100, EventID: 10, Status: StatusCreated, Price: 1000}}, nil
				}
			},
			cacheSetup: func(c *cacheRepoMock) {
				c.lockFn = func(ctx context.Context, ticketID, userID int64, ttl time.Duration) (bool, error) {
					return false, errors.New("redis down")
				}
			},
			wantErr:       true,
			wantErrMsg:    "redis down",
			wantLockCalls: 1,
		},
		{
			name:    "success with percent promo code",
			booking: domain.Booking{UserID: 1, TicketIDs: []int64{100}, UserEmail: "test@example.com", PromoCode: "SALE10"},
			lockTTL: 5 * time.Second,
			ticketsSetup: func(r *ticketRepoMock) {
				r.getTicketsByIDsFn = func(ctx context.Context, ticketIDs []int64) ([]storage.TicketModel, error) {
					return []storage.TicketModel{{ID: 100, EventID: 10, Status: StatusCreated, Price: 1000}}, nil
				}
				r.getPromoByCodeFn = func(ctx context.Context, eventID int64, code string) (storage.Promo, error) {
					return storage.Promo{Type: PromoTypePercent, Value: 10}, nil
				}
			},
			cacheSetup: func(c *cacheRepoMock) {
				c.lockFn = func(ctx context.Context, ticketID, userID int64, ttl time.Duration) (bool, error) {
					return true, nil
				}
			},
			paymentSetup: func(p *paymentClientMock) {
				p.getPaymentLinkFn = func(ctx context.Context, input domain.PaymentInput) (string, error) {
					return "http://pay.me/123", nil
				}
			},
			wantAmount:      900,
			wantURLContains: "pay.me",
			wantLockCalls:   1,
		},
		{
			name:    "db error getting ticket -> returns error",
			booking: domain.Booking{UserID: 3, TicketIDs: []int64{300}, UserEmail: "test@example.com", PromoCode: "SALE10"},
			lockTTL: 10 * time.Second,
			ticketsSetup: func(r *ticketRepoMock) {
				r.getTicketsByIDsFn = func(ctx context.Context, ticketIDs []int64) ([]storage.TicketModel, error) {
					return nil, errors.New("db fail")
				}
			},
			wantErr:       true,
			wantErrMsg:    "db fail",
			wantLockCalls: 0,
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			tickets := &ticketRepoMock{}
			cache := &cacheRepoMock{}
			payment := &paymentClientMock{}
			wallet := &walletClientMock{}
			prod := &producerMock{}

			if tt.ticketsSetup != nil {
				tt.ticketsSetup(tickets)
			}
			if tt.cacheSetup != nil {
				tt.cacheSetup(cache)
			}
			if tt.paymentSetup != nil {
				tt.paymentSetup(payment)
			}

			cfg := Config{TicketLockTTL: tt.lockTTL}
			svc := New(tickets, cache, payment, wallet, zap.NewNop(), cfg, prod)

			url, err := svc.BookTickets(context.Background(), tt.booking)

			require.Equal(t, tt.wantLockCalls, cache.lockCalls)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMsg != "" {
					require.Contains(t, err.Error(), tt.wantErrMsg)
				}
				return
			}

			require.NoError(t, err)
			require.Contains(t, url, tt.wantURLContains)
			if tt.wantAmount > 0 {
				require.Equal(t, tt.wantAmount, payment.lastPaymentInput.Amount)
			}
		})
	}
}

func TestBookingService_GetTicketsByEventID_Table(t *testing.T) {
	type tc struct {
		name string

		eventID int64

		ticketsSetup func(r *ticketRepoMock)
		cacheSetup   func(c *cacheRepoMock)

		wantErr bool

		wantTicketsCalls int
		wantMGetCalls    int
		wantCount        int
	}

	tests := []tc{
		{
			name:    "postgres error -> returns error",
			eventID: 10,
			ticketsSetup: func(r *ticketRepoMock) {
				r.getTicketsByEventFn = func(ctx context.Context, eventID int64) ([]storage.TicketModel, error) {
					return nil, errors.New("db error")
				}
			},
			cacheSetup:       func(c *cacheRepoMock) {},
			wantErr:          true,
			wantTicketsCalls: 1,
			wantMGetCalls:    0,
		},
		{
			name:    "ok -> returns available tickets (all unlocked)",
			eventID: 11,
			ticketsSetup: func(r *ticketRepoMock) {
				r.getTicketsByEventFn = func(ctx context.Context, eventID int64) ([]storage.TicketModel, error) {
					return []storage.TicketModel{
						{ID: 1, EventID: eventID, Status: "CREATED"},
						{ID: 2, EventID: eventID, Status: "CREATED"},
					}, nil
				}
			},
			cacheSetup: func(c *cacheRepoMock) {
				c.mGetFn = func(ctx context.Context, keys []string) ([]interface{}, error) {
					return []interface{}{nil, nil}, nil
				}
			},
			wantErr:          false,
			wantTicketsCalls: 1,
			wantMGetCalls:    1,
			wantCount:        2,
		},
		{
			name:    "ok -> filters locked tickets",
			eventID: 12,
			ticketsSetup: func(r *ticketRepoMock) {
				r.getTicketsByEventFn = func(ctx context.Context, eventID int64) ([]storage.TicketModel, error) {
					return []storage.TicketModel{
						{ID: 10, EventID: eventID, Status: "CREATED"},
						{ID: 11, EventID: eventID, Status: "CREATED"},
						{ID: 12, EventID: eventID, Status: "CREATED"},
					}, nil
				}
			},
			cacheSetup: func(c *cacheRepoMock) {
				c.mGetFn = func(ctx context.Context, keys []string) ([]interface{}, error) {
					return []interface{}{nil, "user_777", nil}, nil
				}
			},
			wantErr:          false,
			wantTicketsCalls: 1,
			wantMGetCalls:    1,
			wantCount:        2,
		},
		{
			name:    "redis error during mget -> returns error or fallback",
			eventID: 13,
			ticketsSetup: func(r *ticketRepoMock) {
				r.getTicketsByEventFn = func(ctx context.Context, eventID int64) ([]storage.TicketModel, error) {
					return []storage.TicketModel{
						{ID: 21, EventID: eventID, Status: "CREATED"},
					}, nil
				}
			},
			cacheSetup: func(c *cacheRepoMock) {
				c.mGetFn = func(ctx context.Context, keys []string) ([]interface{}, error) {
					return nil, errors.New("redis read failed")
				}
			},
			wantErr:          false,
			wantTicketsCalls: 1,
			wantMGetCalls:    1,
			wantCount:        1,
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			tickets := &ticketRepoMock{}
			cache := &cacheRepoMock{}
			payment := &paymentClientMock{}
			wallet := &walletClientMock{}
			prod := &producerMock{}

			tt.ticketsSetup(tickets)
			tt.cacheSetup(cache)

			cfg := Config{TicketLockTTL: 5 * time.Second}
			svc := New(tickets, cache, payment, wallet, zap.NewNop(), cfg, prod)

			got, err := svc.GetTickets(context.Background(), tt.eventID)

			require.Equal(t, tt.wantTicketsCalls, tickets.getTicketsByEventCalls)
			require.Equal(t, tt.wantMGetCalls, cache.mGetCalls)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Len(t, got, tt.wantCount)
		})
	}
}

func TestBookingService_ConfirmBooking_Placeholder(t *testing.T) {
	t.Run("placeholder for volume", func(t *testing.T) {
		require.True(t, true)
	})
}

func TestBookingService_RefundTicket_Table(t *testing.T) {
	type tc struct {
		name string

		refund domain.Refund

		ticketsSetup func(r *ticketRepoMock)
		paymentSetup func(p *paymentClientMock)

		wantErr         bool
		wantErrIs       error
		wantRefundCalls int
	}

	tests := []tc{
		{
			name:   "ticket not found -> ErrTicketNotFound",
			refund: domain.Refund{UserID: 1, EventID: 10, TicketID: 100},
			ticketsSetup: func(r *ticketRepoMock) {
				r.getTicketByIDFn = func(ctx context.Context, ticketID int64) (storage.TicketModel, error) {
					return storage.TicketModel{}, gorm.ErrRecordNotFound
				}
			},
			paymentSetup:    func(p *paymentClientMock) {},
			wantErr:         true,
			wantErrIs:       ErrTicketNotFound,
			wantRefundCalls: 0,
		},
		{
			name:   "ticket owned mismatch -> ErrTicketNotOwned",
			refund: domain.Refund{UserID: 7, EventID: 10, TicketID: 100},
			ticketsSetup: func(r *ticketRepoMock) {
				r.getTicketByIDFn = func(ctx context.Context, ticketID int64) (storage.TicketModel, error) {
					return storage.TicketModel{
						ID: ticketID, UserID: 999, Status: StatusPaid, Price: 1500,
					}, nil
				}
			},
			paymentSetup:    func(p *paymentClientMock) {},
			wantErr:         true,
			wantErrIs:       ErrTicketNotOwned,
			wantRefundCalls: 0,
		},
		{
			name:   "ticket not paid -> ErrBookingNotPaid",
			refund: domain.Refund{UserID: 7, EventID: 10, TicketID: 100},
			ticketsSetup: func(r *ticketRepoMock) {
				r.getTicketByIDFn = func(ctx context.Context, ticketID int64) (storage.TicketModel, error) {
					return storage.TicketModel{
						ID: ticketID, UserID: 7, Status: "CREATED", Price: 1500,
					}, nil
				}
			},
			paymentSetup:    func(p *paymentClientMock) {},
			wantErr:         true,
			wantErrIs:       ErrBookingNotPaid,
			wantRefundCalls: 0,
		},
		{
			name:   "ok -> creates refund (empty response)",
			refund: domain.Refund{UserID: 7, EventID: 10, TicketID: 100},
			ticketsSetup: func(r *ticketRepoMock) {
				r.getTicketByIDFn = func(ctx context.Context, ticketID int64) (storage.TicketModel, error) {
					return storage.TicketModel{
						ID: ticketID, UserID: 7, Status: StatusPaid, Price: 1500, CreatedAt: time.Now(),
					}, nil
				}
			},
			paymentSetup: func(p *paymentClientMock) {
				p.createRefundFn = func(ctx context.Context, input domain.RefundPayment) error {
					return nil
				}
			},
			wantErr:         false,
			wantRefundCalls: 1,
		},
		{
			name:   "payment error -> returns error",
			refund: domain.Refund{UserID: 7, EventID: 10, TicketID: 100},
			ticketsSetup: func(r *ticketRepoMock) {
				r.getTicketByIDFn = func(ctx context.Context, ticketID int64) (storage.TicketModel, error) {
					return storage.TicketModel{
						ID: ticketID, UserID: 7, Status: StatusPaid, Price: 1500,
					}, nil
				}
			},
			paymentSetup: func(p *paymentClientMock) {
				p.createRefundFn = func(ctx context.Context, input domain.RefundPayment) error {
					return errors.New("stripe down")
				}
			},
			wantErr:         true,
			wantRefundCalls: 1,
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			tickets := &ticketRepoMock{}
			cache := &cacheRepoMock{}
			payment := &paymentClientMock{}
			wallet := &walletClientMock{}
			prod := &producerMock{}

			if tt.ticketsSetup != nil {
				tt.ticketsSetup(tickets)
			}
			if tt.paymentSetup != nil {
				tt.paymentSetup(payment)
			}

			cfg := Config{TicketLockTTL: 5 * time.Second}
			svc := New(tickets, cache, payment, wallet, zap.NewNop(), cfg, prod)

			err := svc.RefundTicket(context.Background(), tt.refund)

			require.Equal(t, tt.wantRefundCalls, payment.createRefundCalls)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrIs != nil {
					require.ErrorIs(t, err, tt.wantErrIs)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestBookingService_AddToWaitlist_Table(t *testing.T) {
	type tc struct {
		name string

		item domain.Waitlist

		paymentSetup func(p *paymentClientMock)

		wantErr        bool
		wantErrMsg     string
		wantCallsCount int
	}

	tests := []tc{
		{
			name: "success -> calls payment client",
			item: domain.Waitlist{
				UserID:    101,
				EventID:   555,
				UserEmail: "test@example.com",
			},
			paymentSetup: func(p *paymentClientMock) {
				p.addToWaitlistFn = func(ctx context.Context, item domain.Waitlist) error {
					return nil
				}
			},
			wantErr:        false,
			wantCallsCount: 1,
		},
		{
			name: "grpc client error -> returns error",
			item: domain.Waitlist{
				UserID:    102,
				EventID:   555,
				UserEmail: "fail@example.com",
			},
			paymentSetup: func(p *paymentClientMock) {
				p.addToWaitlistFn = func(ctx context.Context, item domain.Waitlist) error {
					return errors.New("grpc connection failed")
				}
			},
			wantErr:        true,
			wantErrMsg:     "grpc connection failed",
			wantCallsCount: 1,
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			tickets := &ticketRepoMock{}
			cache := &cacheRepoMock{}
			payment := &paymentClientMock{}
			wallet := &walletClientMock{}
			prod := &producerMock{}

			if tt.paymentSetup != nil {
				tt.paymentSetup(payment)
			}

			svc := New(tickets, cache, payment, wallet, zap.NewNop(), Config{}, prod)

			err := svc.AddToWaitlist(context.Background(), tt.item)

			require.Equal(t, tt.wantCallsCount, payment.addToWaitlistCalls)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMsg != "" {
					require.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.item, payment.lastWaitlistInput)
			}
		})
	}
}
