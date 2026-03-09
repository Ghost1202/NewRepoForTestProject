package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/turtlepavlo/event-service/internal/domain"
	"github.com/turtlepavlo/event-service/internal/storage"
)

type MockEventRepo struct {
	mock.Mock
}

func (m *MockEventRepo) CreateEvent(ctx context.Context, e storage.Events, t []storage.Tickets, p []storage.Promo, early []storage.Early, bundles []storage.Bundle) (int64, error) {
	args := m.Called(ctx, e, t, p, early, bundles)
	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	if val, ok := args.Get(0).(int); ok {
		return int64(val), args.Error(1)
	}
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockEventRepo) UpdateEvent(ctx context.Context, e storage.Events) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *MockEventRepo) GetEventPerformerID(ctx context.Context, id int64) (int64, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	if val, ok := args.Get(0).(int); ok {
		return int64(val), args.Error(1)
	}
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockEventRepo) UpdateTicketStatus(ctx context.Context, ticketID int64, fromStatus, toStatus string) (int64, error) {
	args := m.Called(ctx, ticketID, fromStatus, toStatus)
	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	if val, ok := args.Get(0).(int); ok {
		return int64(val), args.Error(1)
	}
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockEventRepo) UpdateTicketOwner(ctx context.Context, ticketID, fromUserID, toUserID int64) (int64, error) {
	args := m.Called(ctx, ticketID, fromUserID, toUserID)
	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	if val, ok := args.Get(0).(int); ok {
		return int64(val), args.Error(1)
	}
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockEventRepo) UpdateTicketStatusAndOwner(ctx context.Context, ticketID int64, status string, userID int64) (int64, error) {
	args := m.Called(ctx, ticketID, status, userID)
	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	if val, ok := args.Get(0).(int); ok {
		return int64(val), args.Error(1)
	}
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockEventRepo) GetTicketMeta(ctx context.Context, ticketID int64) (storage.TicketMeta, bool, error) {
	args := m.Called(ctx, ticketID)

	var meta storage.TicketMeta
	if args.Get(0) != nil {
		meta = args.Get(0).(storage.TicketMeta)
	}

	found := false
	if args.Get(1) != nil {
		found = args.Get(1).(bool)
	}

	return meta, found, args.Error(2)
}

func (m *MockEventRepo) AddPromo(ctx context.Context, promos []storage.Promo) ([]int64, error) {
	args := m.Called(ctx, promos)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockEventRepo) AddEarly(ctx context.Context, early []storage.Early) ([]int64, error) {
	args := m.Called(ctx, early)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockEventRepo) AddBundle(ctx context.Context, bundles []storage.Bundle) ([]int64, error) {
	args := m.Called(ctx, bundles)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockEventRepo) GetPromo(ctx context.Context, eventID, limit, offset int64) ([]storage.Promo, error) {
	args := m.Called(ctx, eventID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]storage.Promo), args.Error(1)
}

func (m *MockEventRepo) GetEarly(ctx context.Context, eventID, limit, offset int64) ([]storage.Early, error) {
	args := m.Called(ctx, eventID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]storage.Early), args.Error(1)
}

func (m *MockEventRepo) GetBundle(ctx context.Context, eventID, limit, offset int64) ([]storage.Bundle, error) {
	args := m.Called(ctx, eventID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]storage.Bundle), args.Error(1)
}

func TestEventService_CreateEvent(t *testing.T) {
	now := time.Now()

	makeSectors := func() []domain.TicketSector {
		return []domain.TicketSector{
			{
				Name:        "A",
				Type:        "standard",
				Price:       1000,
				RowsCount:   2,
				SeatsPerRow: 3,
			},
			{
				Name:        "VIP",
				Type:        "vip",
				Price:       2500,
				RowsCount:   1,
				SeatsPerRow: 2,
			},
		}
	}

	tests := []struct {
		name      string
		actor     domain.Performer
		event     *domain.Event
		sectors   []domain.TicketSector
		mockSetup func(m *MockEventRepo)
		wantErr   bool
		wantPanic bool
	}{
		{
			name:    "success basic (event + tickets)",
			actor:   domain.Performer{ID: 1, Role: Performer},
			event:   &domain.Event{VenueID: 20, Name: "Rock Fest", StartDate: now.Add(24 * time.Hour)},
			sectors: makeSectors(),
			mockSetup: func(m *MockEventRepo) {
				expectedTickets := int(2*3 + 1*2)

				m.On(
					"CreateEvent",
					mock.Anything,
					mock.MatchedBy(func(e storage.Events) bool {
						return e.Name == "Rock Fest" &&
							e.VenueID == 20 &&
							e.PerformerID == 1
					}),
					mock.MatchedBy(func(t []storage.Tickets) bool {
						if len(t) != expectedTickets {
							return false
						}
						if len(t) > 0 && t[0].VenueID != 20 {
							return false
						}
						return true
					}),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(int64(100), nil).Once()
			},
			wantErr: false,
		},
		{
			name:      "non performer forbidden",
			actor:     domain.Performer{ID: 2, Role: "user"},
			event:     &domain.Event{VenueID: 20, Name: "Jazz Night", StartDate: now.Add(24 * time.Hour)},
			sectors:   makeSectors(),
			mockSetup: func(m *MockEventRepo) {},
			wantErr:   true,
		},
		{
			name:      "role lowercase performer (forbidden, case-sensitive)",
			actor:     domain.Performer{ID: 3, Role: "performer"},
			event:     &domain.Event{VenueID: 20, Name: "Case Sensitivity", StartDate: now.Add(24 * time.Hour)},
			sectors:   makeSectors(),
			mockSetup: func(m *MockEventRepo) {},
			wantErr:   true,
		},
		{
			name:      "nil event",
			actor:     domain.Performer{ID: 4, Role: Performer},
			event:     nil,
			sectors:   makeSectors(),
			mockSetup: func(m *MockEventRepo) {},
			wantErr:   true,
		},
		{
			name:      "past date",
			actor:     domain.Performer{ID: 5, Role: Performer},
			event:     &domain.Event{VenueID: 20, Name: "Retro Party", StartDate: now.Add(-48 * time.Hour)},
			sectors:   makeSectors(),
			mockSetup: func(m *MockEventRepo) {},
			wantErr:   true,
		},
		{
			name:    "repo failure",
			actor:   domain.Performer{ID: 6, Role: Performer},
			event:   &domain.Event{VenueID: 20, Name: "Indie Show", StartDate: now.Add(4 * time.Hour)},
			sectors: makeSectors(),
			mockSetup: func(m *MockEventRepo) {
				m.On("CreateEvent",
					mock.Anything,
					mock.AnythingOfType("storage.Events"),
					mock.AnythingOfType("[]storage.Tickets"),
					mock.AnythingOfType("[]storage.Promo"),
					mock.AnythingOfType("[]storage.Early"),
					mock.AnythingOfType("[]storage.Bundle"),
				).Return(int64(0), errors.New("db failure")).Once()
			},
			wantErr: true,
		},
		{
			name:    "repo panic simulated (recover in test)",
			actor:   domain.Performer{ID: 7, Role: Performer},
			event:   &domain.Event{VenueID: 20, Name: "Panic Sim", StartDate: now.Add(4 * time.Hour)},
			sectors: makeSectors(),
			mockSetup: func(m *MockEventRepo) {
				m.On("CreateEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) { panic("unexpected panic") }).
					Return(int64(0), nil).Once()
			},
			wantErr:   true,
			wantPanic: true,
		},
		{
			name:    "empty sectors (allowed) => repo called with empty tickets slice",
			actor:   domain.Performer{ID: 8, Role: Performer},
			event:   &domain.Event{VenueID: 20, Name: "No Sectors", StartDate: now.Add(24 * time.Hour)},
			sectors: nil,
			mockSetup: func(m *MockEventRepo) {
				m.On(
					"CreateEvent",
					mock.Anything,
					mock.AnythingOfType("storage.Events"),
					mock.MatchedBy(func(t []storage.Tickets) bool { return len(t) == 0 }),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(int64(777), nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := new(MockEventRepo)
			tc.mockSetup(repo)

			srv := NewEventService(repo, zap.NewNop())

			var (
				pan any
				err error
			)

			func() {
				defer func() {
					if r := recover(); r != nil {
						pan = r
					}
				}()

				_, err = srv.CreateEvent(context.Background(), tc.actor, tc.event, tc.sectors, nil, nil, nil)
			}()

			if tc.wantPanic {
				assert.NotNil(t, pan)
			} else {
				assert.Nil(t, pan)
				if tc.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}

			if tc.actor.Role != Performer || tc.event == nil || (tc.event != nil && tc.event.StartDate.Before(time.Now())) {
				repo.AssertNotCalled(t, "CreateEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestEventService_UpdateEvent(t *testing.T) {
	type tc struct {
		name string

		actor domain.Performer
		event *domain.Event

		mockSetup func(m *MockEventRepo)

		wantErr          bool
		wantPanic        bool
		wantGetOwnerCall bool
		wantUpdateCall   bool
	}

	tests := []tc{
		{
			name:  "success update basic",
			actor: domain.Performer{ID: 10, Role: Performer},
			event: &domain.Event{ID: 55, Name: "Charity Event"},
			mockSetup: func(m *MockEventRepo) {
				m.On("GetEventPerformerID", mock.Anything, int64(55)).Return(int64(10), nil).Once()
				m.On("UpdateEvent", mock.Anything, mock.AnythingOfType("storage.Events")).Return(nil).Once()
			},
			wantErr:          false,
			wantPanic:        false,
			wantGetOwnerCall: true,
			wantUpdateCall:   true,
		},
		{
			name:  "success update with empty name (no validation here)",
			actor: domain.Performer{ID: 10, Role: Performer},
			event: &domain.Event{ID: 56, Name: ""},
			mockSetup: func(m *MockEventRepo) {
				m.On("GetEventPerformerID", mock.Anything, int64(56)).Return(int64(10), nil).Once()
				m.On("UpdateEvent", mock.Anything, mock.AnythingOfType("storage.Events")).Return(nil).Once()
			},
			wantErr:          false,
			wantPanic:        false,
			wantGetOwnerCall: true,
			wantUpdateCall:   true,
		},
		{
			name:             "nil event -> panic (matches current implementation)",
			actor:            domain.Performer{ID: 10, Role: Performer},
			event:            nil,
			mockSetup:        func(m *MockEventRepo) {},
			wantErr:          true,
			wantPanic:        true,
			wantGetOwnerCall: false,
			wantUpdateCall:   false,
		},
		{
			name:  "actor role ignored by UpdateEvent -> still works if owner matches",
			actor: domain.Performer{ID: 10, Role: "user"},
			event: &domain.Event{ID: 57, Name: "Role Irrelevant"},
			mockSetup: func(m *MockEventRepo) {
				m.On("GetEventPerformerID", mock.Anything, int64(57)).Return(int64(10), nil).Once()
				m.On("UpdateEvent", mock.Anything, mock.AnythingOfType("storage.Events")).Return(nil).Once()
			},
			wantErr:          false,
			wantPanic:        false,
			wantGetOwnerCall: true,
			wantUpdateCall:   true,
		},
		{
			name:  "wrong owner => access denied",
			actor: domain.Performer{ID: 2, Role: Performer},
			event: &domain.Event{ID: 88, Name: "Secret Show"},
			mockSetup: func(m *MockEventRepo) {
				m.On("GetEventPerformerID", mock.Anything, int64(88)).Return(int64(5), nil).Once()
			},
			wantErr:          true,
			wantPanic:        false,
			wantGetOwnerCall: true,
			wantUpdateCall:   false,
		},
		{
			name:  "GetEventPerformerID error",
			actor: domain.Performer{ID: 1, Role: Performer},
			event: &domain.Event{ID: 77, Name: "DB Error"},
			mockSetup: func(m *MockEventRepo) {
				m.On("GetEventPerformerID", mock.Anything, int64(77)).
					Return(int64(0), errors.New("db err")).Once()
			},
			wantErr:          true,
			wantPanic:        false,
			wantGetOwnerCall: true,
			wantUpdateCall:   false,
		},
		{
			name:  "UpdateEvent error (failed update)",
			actor: domain.Performer{ID: 11, Role: Performer},
			event: &domain.Event{ID: 99, Name: "Update Fail"},
			mockSetup: func(m *MockEventRepo) {
				m.On("GetEventPerformerID", mock.Anything, int64(99)).Return(int64(11), nil).Once()
				m.On("UpdateEvent", mock.Anything, mock.AnythingOfType("storage.Events")).
					Return(errors.New("failed update")).Once()
			},
			wantErr:          true,
			wantPanic:        false,
			wantGetOwnerCall: true,
			wantUpdateCall:   true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := new(MockEventRepo)
			tc.mockSetup(repo)

			srv := NewEventService(repo, zap.NewNop())

			var (
				err error
				pan any
			)

			func() {
				defer func() {
					if r := recover(); r != nil {
						pan = r
					}
				}()
				err = srv.UpdateEvent(context.Background(), tc.actor, tc.event)
			}()

			if tc.wantPanic {
				assert.NotNil(t, pan)
			} else {
				assert.Nil(t, pan)
				if tc.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}

			if !tc.wantGetOwnerCall {
				repo.AssertNotCalled(t, "GetEventPerformerID", mock.Anything, mock.Anything)
			}
			if !tc.wantUpdateCall {
				repo.AssertNotCalled(t, "UpdateEvent", mock.Anything, mock.Anything)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestEventService_TransferTicket_Table(t *testing.T) {
	tests := []struct {
		name      string
		transfer  domain.TicketTransfer
		mockSetup func(m *MockEventRepo)
		wantErr   bool
	}{
		{
			name:      "invalid input -> skip (no repo call, no error)",
			transfer:  domain.TicketTransfer{TicketID: 0, FromUserID: 1, ToUserID: 2},
			mockSetup: func(m *MockEventRepo) {},
			wantErr:   false,
		},
		{
			name:      "same user -> skip (no repo call, no error)",
			transfer:  domain.TicketTransfer{TicketID: 10, FromUserID: 5, ToUserID: 5},
			mockSetup: func(m *MockEventRepo) {},
			wantErr:   false,
		},
		{
			name:     "repo error -> returns error",
			transfer: domain.TicketTransfer{TicketID: 11, FromUserID: 1, ToUserID: 2},
			mockSetup: func(m *MockEventRepo) {
				m.On(
					"UpdateTicketOwner",
					mock.Anything,
					int64(11),
					int64(1),
					int64(2),
				).Return(int64(0), errors.New("db down")).Once()
			},
			wantErr: true,
		},
		{
			name:     "ok -> repo called, no error",
			transfer: domain.TicketTransfer{TicketID: 12, FromUserID: 7, ToUserID: 8},
			mockSetup: func(m *MockEventRepo) {
				m.On(
					"UpdateTicketOwner",
					mock.Anything,
					int64(12),
					int64(7),
					int64(8),
				).Return(int64(1), nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := new(MockEventRepo)
			tc.mockSetup(repo)

			srv := NewEventService(repo, zap.NewNop())

			err := srv.TransferTicket(context.Background(), tc.transfer)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.transfer.TicketID <= 0 || tc.transfer.FromUserID <= 0 || tc.transfer.ToUserID <= 0 || tc.transfer.FromUserID == tc.transfer.ToUserID {
				repo.AssertNotCalled(t, "UpdateTicketOwner", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}

			repo.AssertExpectations(t)
		})
	}
}
