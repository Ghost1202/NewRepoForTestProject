package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/internal/storage"
)

type ticketRepoMock struct {
	mu sync.Mutex

	getTicketsByEventFn func(ctx context.Context, eventID int64) ([]storage.TicketModel, error)
	getTicketByIDFn     func(ctx context.Context, ticketID int64) (storage.TicketModel, error)
	getTicketsByIDsFn   func(ctx context.Context, ticketIDs []int64) ([]storage.TicketModel, error)
	getTicketsByUserFn  func(ctx context.Context, userID int64) ([]storage.TicketModel, error)
	getPromoByCodeFn    func(ctx context.Context, eventID int64, code string) (storage.Promo, error)
	getEarlyByEventFn   func(ctx context.Context, eventID int64) ([]storage.Early, error)
	getBundleByEventFn  func(ctx context.Context, eventID int64) ([]storage.Bundle, error)
	getEarlyByIDFn      func(ctx context.Context, earlyID int64) (storage.Early, error)
	getBundleByIDFn     func(ctx context.Context, bundleID int64) (storage.Bundle, error)

	getTicketsByEventCalls int
	getTicketByIDCalls     int
	getTicketsByIDsCalls   int
	getTicketsByUserCalls  int
	getPromoByCodeCalls    int
	getEarlyByEventCalls   int
	getBundleByEventCalls  int
	getEarlyByIDCalls      int
	getBundleByIDCalls     int

	lastGetTicketsByEvent struct {
		eventID int64
	}
	lastGetTicketByID struct {
		ticketID int64
	}
	lastGetTicketsByIDs struct {
		ticketIDs []int64
	}
	lastGetTicketsByUser struct {
		userID int64
	}
	lastGetPromoByCode struct {
		eventID int64
		code    string
	}
	lastGetEarlyByEvent struct {
		eventID int64
	}
	lastGetBundleByEvent struct {
		eventID int64
	}
	lastGetEarlyByID struct {
		earlyID int64
	}
	lastGetBundleByID struct {
		bundleID int64
	}
}

func (m *ticketRepoMock) GetTicketsByEvent(ctx context.Context, eventID int64) ([]storage.TicketModel, error) {
	m.mu.Lock()
	m.getTicketsByEventCalls++
	m.lastGetTicketsByEvent.eventID = eventID
	fn := m.getTicketsByEventFn
	m.mu.Unlock()

	if fn == nil {
		return nil, errors.New("ticketRepoMock.GetTicketsByEvent not configured")
	}
	return fn(ctx, eventID)
}

func (m *ticketRepoMock) GetTicketByID(ctx context.Context, ticketID int64) (storage.TicketModel, error) {
	m.mu.Lock()
	m.getTicketByIDCalls++
	m.lastGetTicketByID.ticketID = ticketID
	fn := m.getTicketByIDFn
	m.mu.Unlock()

	if fn == nil {
		return storage.TicketModel{}, errors.New("ticketRepoMock.GetTicketByID not configured")
	}
	return fn(ctx, ticketID)
}

func (m *ticketRepoMock) GetTicketsByIDs(ctx context.Context, ticketIDs []int64) ([]storage.TicketModel, error) {
	m.mu.Lock()
	m.getTicketsByIDsCalls++
	m.lastGetTicketsByIDs.ticketIDs = ticketIDs
	fn := m.getTicketsByIDsFn
	m.mu.Unlock()

	if fn == nil {
		return nil, errors.New("ticketRepoMock.GetTicketsByIDs not configured")
	}
	return fn(ctx, ticketIDs)
}

func (m *ticketRepoMock) GetTicketsByUserID(ctx context.Context, userID int64) ([]storage.TicketModel, error) {
	m.mu.Lock()
	m.getTicketsByUserCalls++
	m.lastGetTicketsByUser.userID = userID
	fn := m.getTicketsByUserFn
	m.mu.Unlock()

	if fn == nil {
		return nil, errors.New("ticketRepoMock.GetTicketsByUserID not configured")
	}
	return fn(ctx, userID)
}

func (m *ticketRepoMock) GetPromoByCode(ctx context.Context, eventID int64, code string) (storage.Promo, error) {
	m.mu.Lock()
	m.getPromoByCodeCalls++
	m.lastGetPromoByCode.eventID = eventID
	m.lastGetPromoByCode.code = code
	fn := m.getPromoByCodeFn
	m.mu.Unlock()

	if fn == nil {
		return storage.Promo{}, errors.New("ticketRepoMock.GetPromoByCode not configured")
	}
	return fn(ctx, eventID, code)
}

func (m *ticketRepoMock) GetEarlyByEvent(ctx context.Context, eventID int64) ([]storage.Early, error) {
	m.mu.Lock()
	m.getEarlyByEventCalls++
	m.lastGetEarlyByEvent.eventID = eventID
	fn := m.getEarlyByEventFn
	m.mu.Unlock()

	if fn == nil {
		return nil, errors.New("ticketRepoMock.GetEarlyByEvent not configured")
	}
	return fn(ctx, eventID)
}

func (m *ticketRepoMock) GetBundleByEvent(ctx context.Context, eventID int64) ([]storage.Bundle, error) {
	m.mu.Lock()
	m.getBundleByEventCalls++
	m.lastGetBundleByEvent.eventID = eventID
	fn := m.getBundleByEventFn
	m.mu.Unlock()

	if fn == nil {
		return nil, errors.New("ticketRepoMock.GetBundleByEvent not configured")
	}
	return fn(ctx, eventID)
}

func (m *ticketRepoMock) GetEarlyByID(ctx context.Context, earlyID int64) (storage.Early, error) {
	m.mu.Lock()
	m.getEarlyByIDCalls++
	m.lastGetEarlyByID.earlyID = earlyID
	fn := m.getEarlyByIDFn
	m.mu.Unlock()

	if fn == nil {
		return storage.Early{}, errors.New("ticketRepoMock.GetEarlyByID not configured")
	}
	return fn(ctx, earlyID)
}

func (m *ticketRepoMock) GetBundleByID(ctx context.Context, bundleID int64) (storage.Bundle, error) {
	m.mu.Lock()
	m.getBundleByIDCalls++
	m.lastGetBundleByID.bundleID = bundleID
	fn := m.getBundleByIDFn
	m.mu.Unlock()

	if fn == nil {
		return storage.Bundle{}, errors.New("ticketRepoMock.GetBundleByID not configured")
	}
	return fn(ctx, bundleID)
}

type cacheRepoMock struct {
	mu sync.Mutex

	lockFn   func(ctx context.Context, ticketID int64, userID int64, ttl time.Duration) (bool, error)
	mGetFn   func(ctx context.Context, keys []string) ([]interface{}, error)
	unlockFn func(ctx context.Context, ticketIDs []int64) error

	lockCalls   int
	mGetCalls   int
	unlockCalls int

	lastLock struct {
		ticketID int64
		userID   int64
		ttl      time.Duration
	}
	lastUnlock struct {
		ticketIDs []int64
	}
}

func (m *cacheRepoMock) Lock(ctx context.Context, ticketID, userID int64, ttl time.Duration) (bool, error) {
	m.mu.Lock()
	m.lockCalls++
	m.lastLock.ticketID = ticketID
	m.lastLock.userID = userID
	m.lastLock.ttl = ttl
	fn := m.lockFn
	m.mu.Unlock()

	if fn == nil {
		return false, errors.New("cacheRepoMock.Lock not configured")
	}
	return fn(ctx, ticketID, userID, ttl)
}

func (m *cacheRepoMock) MGet(ctx context.Context, keys []string) ([]interface{}, error) {
	m.mu.Lock()
	m.mGetCalls++
	fn := m.mGetFn
	m.mu.Unlock()

	if fn == nil {
		return nil, errors.New("cacheRepoMock.MGet not configured")
	}
	return fn(ctx, keys)
}

func (m *cacheRepoMock) Unlock(ctx context.Context, ticketIDs []int64) error {
	m.mu.Lock()
	m.unlockCalls++
	m.lastUnlock.ticketIDs = ticketIDs
	fn := m.unlockFn
	m.mu.Unlock()

	if fn == nil {
		return errors.New("cacheRepoMock.Unlock not configured")
	}
	return fn(ctx, ticketIDs)
}

type paymentClientMock struct {
	mu sync.Mutex

	getPaymentLinkFn func(ctx context.Context, input domain.PaymentInput) (string, error)
	createRefundFn   func(ctx context.Context, input domain.RefundPayment) error
	addToWaitlistFn  func(ctx context.Context, item domain.Waitlist) error

	getPaymentLinkCalls int
	createRefundCalls   int
	addToWaitlistCalls  int

	lastPaymentInput  domain.PaymentInput
	lastRefundInput   domain.RefundPayment
	lastWaitlistInput domain.Waitlist
}

func (m *paymentClientMock) GetPaymentLink(ctx context.Context, input domain.PaymentInput) (string, error) {
	m.mu.Lock()
	m.getPaymentLinkCalls++
	m.lastPaymentInput = input
	fn := m.getPaymentLinkFn
	m.mu.Unlock()

	if fn == nil {
		return "", errors.New("paymentClientMock.GetPaymentLink not configured")
	}
	return fn(ctx, input)
}

func (m *paymentClientMock) CreateRefund(ctx context.Context, input domain.RefundPayment) error {
	m.mu.Lock()
	m.createRefundCalls++
	m.lastRefundInput = input
	fn := m.createRefundFn
	m.mu.Unlock()

	if fn == nil {
		return errors.New("paymentClientMock.CreateRefund not configured")
	}
	return fn(ctx, input)
}

func (m *paymentClientMock) AddToWaitlist(ctx context.Context, item domain.Waitlist) error {
	m.mu.Lock()
	m.addToWaitlistCalls++
	m.lastWaitlistInput = item
	fn := m.addToWaitlistFn
	m.mu.Unlock()

	if fn == nil {
		return errors.New("paymentClientMock.AddToWaitlist not configured")
	}
	return fn(ctx, item)
}

type walletClientMock struct {
	mu sync.Mutex

	chargeWalletFn    func(ctx context.Context, input domain.WalletCharge) (domain.WalletChargeStatus, error)
	chargeWalletCalls int

	lastChargeInput domain.WalletCharge
}

func (m *walletClientMock) ChargeWallet(ctx context.Context, input domain.WalletCharge) (domain.WalletChargeStatus, error) {
	m.mu.Lock()
	m.chargeWalletCalls++
	m.lastChargeInput = input
	fn := m.chargeWalletFn
	m.mu.Unlock()

	if fn == nil {
		return domain.WalletChargeStatusUnspecified, errors.New("walletClientMock.ChargeWallet not configured")
	}
	return fn(ctx, input)
}
