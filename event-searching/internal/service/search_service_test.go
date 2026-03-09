package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/turtlepavlo/event-searching/internal/domain"
	"github.com/turtlepavlo/event-searching/internal/storage"
	sqlc "github.com/turtlepavlo/event-searching/internal/storage/postgres/sqlc"
	"go.uber.org/zap"
)

type mockSearchStore struct {
	mu sync.Mutex

	getPopularFn func(ctx context.Context, limit, offset int64) ([]storage.EventModel, error)
	searchFn     func(ctx context.Context, filter storage.SearchFilterDTO) ([]storage.EventModel, error)
	promoteFn    func(ctx context.Context, events []storage.EventModel) error

	createCommentFn func(ctx context.Context, arg sqlc.CreateCommentParams) (sqlc.CreateCommentRow, error)
	listCommentsFn  func(ctx context.Context, arg sqlc.ListCommentsByEventParams) ([]sqlc.EventComment, error)
	getRatingFn     func(ctx context.Context, eventID int64) (sqlc.GetRatingByEventRow, error)

	getPopularCalls int
	searchCalls     int
	promoteCalls    int

	createCommentCalls int
	listCommentsCalls  int
	getRatingCalls     int

	lastLimit  int64
	lastOffset int64

	lastCreateComment sqlc.CreateCommentParams
	lastListComments  sqlc.ListCommentsByEventParams
	lastRatingEventID int64

	promoteCh       chan promoteCall
	searchCh        chan storage.SearchFilterDTO
	createCommentCh chan sqlc.CreateCommentParams
	listCommentsCh  chan sqlc.ListCommentsByEventParams
	getRatingCh     chan int64
}

type promoteCall struct {
	ctx    context.Context
	events []storage.EventModel
}

func newMockStore() *mockSearchStore {
	return &mockSearchStore{
		promoteCh:       make(chan promoteCall, 10),
		searchCh:        make(chan storage.SearchFilterDTO, 10),
		createCommentCh: make(chan sqlc.CreateCommentParams, 10),
		listCommentsCh:  make(chan sqlc.ListCommentsByEventParams, 10),
		getRatingCh:     make(chan int64, 10),
	}
}

func (m *mockSearchStore) GetPopular(ctx context.Context, limit, offset int64) ([]storage.EventModel, error) {
	m.mu.Lock()
	m.getPopularCalls++
	m.lastLimit = limit
	m.lastOffset = offset
	fn := m.getPopularFn
	m.mu.Unlock()

	if fn == nil {
		return nil, nil
	}
	return fn(ctx, limit, offset)
}

func (m *mockSearchStore) Search(ctx context.Context, filter storage.SearchFilterDTO) ([]storage.EventModel, error) {
	m.mu.Lock()
	m.searchCalls++
	fn := m.searchFn
	m.mu.Unlock()

	select {
	case m.searchCh <- filter:
	default:
	}

	if fn == nil {
		return nil, nil
	}
	return fn(ctx, filter)
}

func (m *mockSearchStore) PromoteToPopular(ctx context.Context, events []storage.EventModel) error {
	m.mu.Lock()
	m.promoteCalls++
	fn := m.promoteFn
	m.mu.Unlock()

	select {
	case m.promoteCh <- promoteCall{ctx: ctx, events: events}:
	default:
	}

	if fn == nil {
		return nil
	}
	return fn(ctx, events)
}

func (m *mockSearchStore) CreateComment(ctx context.Context, arg sqlc.CreateCommentParams) (sqlc.CreateCommentRow, error) {
	m.mu.Lock()
	m.createCommentCalls++
	m.lastCreateComment = arg
	fn := m.createCommentFn
	m.mu.Unlock()

	select {
	case m.createCommentCh <- arg:
	default:
	}

	if fn == nil {
		return sqlc.CreateCommentRow{}, nil
	}
	return fn(ctx, arg)
}

func (m *mockSearchStore) ListCommentsByEvent(ctx context.Context, arg sqlc.ListCommentsByEventParams) ([]sqlc.EventComment, error) {
	m.mu.Lock()
	m.listCommentsCalls++
	m.lastListComments = arg
	fn := m.listCommentsFn
	m.mu.Unlock()

	select {
	case m.listCommentsCh <- arg:
	default:
	}

	if fn == nil {
		return nil, nil
	}
	return fn(ctx, arg)
}

func (m *mockSearchStore) GetRatingByEvent(ctx context.Context, eventID int64) (sqlc.GetRatingByEventRow, error) {
	m.mu.Lock()
	m.getRatingCalls++
	m.lastRatingEventID = eventID
	fn := m.getRatingFn
	m.mu.Unlock()

	select {
	case m.getRatingCh <- eventID:
	default:
	}

	if fn == nil {
		return sqlc.GetRatingByEventRow{}, nil
	}
	return fn(ctx, eventID)
}

func TestSearchService_SearchEventsFilter_Table(t *testing.T) {
	t.Parallel()

	cfg := Config{
		JWTSecret:            "test",
		CommentsDefaultLimit: 20,
		CommentsMaxLimit:     100,
		CommentsMaxOffset:    100000,
	}

	tests := []struct {
		name        string
		setup       func(m *mockSearchStore)
		filter      domain.SearchFilter
		wantErr     bool
		wantLen     int
		wantPromote bool
	}{
		{
			name: "eventRepo.Search returns error -> return error, no promote",
			setup: func(m *mockSearchStore) {
				m.searchFn = func(_ context.Context, _ storage.SearchFilterDTO) ([]storage.EventModel, error) {
					return nil, errors.New("elastic down")
				}
			},
			filter:      domain.SearchFilter{},
			wantErr:     true,
			wantLen:     0,
			wantPromote: false,
		},
		{
			name: "eventRepo.Search returns empty -> ok, no promote",
			setup: func(m *mockSearchStore) {
				m.searchFn = func(_ context.Context, _ storage.SearchFilterDTO) ([]storage.EventModel, error) {
					return []storage.EventModel{}, nil
				}
			},
			filter:      domain.SearchFilter{},
			wantErr:     false,
			wantLen:     0,
			wantPromote: false,
		},
		{
			name: "eventRepo.Search returns docs -> ok, promote called async",
			setup: func(m *mockSearchStore) {
				m.searchFn = func(_ context.Context, _ storage.SearchFilterDTO) ([]storage.EventModel, error) {
					return []storage.EventModel{
						{ID: 1},
						{ID: 2},
					}, nil
				}
				m.promoteFn = func(_ context.Context, _ []storage.EventModel) error { return nil }
			},
			filter:      domain.SearchFilter{},
			wantErr:     false,
			wantLen:     2,
			wantPromote: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := newMockStore()
			tc.setup(m)

			srv := NewSearchService(cfg, m, m, m, zap.NewNop())

			got, err := srv.SearchEventsFilter(context.Background(), tc.filter)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, got, tc.wantLen)
			}

			if tc.wantPromote {
				select {
				case call := <-m.promoteCh:
					require.Len(t, call.events, tc.wantLen)
					_, hasDeadline := call.ctx.Deadline()
					require.False(t, hasDeadline)
				case <-time.After(500 * time.Millisecond):
					t.Fatalf("expected PromoteToPopular async call, but did not receive it")
				}
			} else {
				select {
				case <-m.promoteCh:
					t.Fatalf("PromoteToPopular should NOT be called")
				case <-time.After(150 * time.Millisecond):
				}
			}
		})
	}
}

func TestSearchService_GetPopularEvents_Table(t *testing.T) {
	t.Parallel()

	cfg := Config{
		JWTSecret:            "test",
		CommentsDefaultLimit: 20,
		CommentsMaxLimit:     100,
		CommentsMaxOffset:    100000,
	}

	tests := []struct {
		name            string
		limit           int64
		offset          int64
		setup           func(m *mockSearchStore)
		wantErr         bool
		wantLen         int
		wantSearchCalls int
		wantPromote     bool
	}{
		{
			name:   "redis hit -> return docs, no search, no promote",
			limit:  2,
			offset: 0,
			setup: func(m *mockSearchStore) {
				m.getPopularFn = func(_ context.Context, _ int64, _ int64) ([]storage.EventModel, error) {
					return []storage.EventModel{{ID: 10}, {ID: 11}}, nil
				}
				m.searchFn = func(_ context.Context, _ storage.SearchFilterDTO) ([]storage.EventModel, error) {
					t.Fatalf("Search must not be called on redis hit")
					return nil, nil
				}
			},
			wantErr:         false,
			wantLen:         2,
			wantSearchCalls: 0,
			wantPromote:     false,
		},
		{
			name:   "redis error + empty -> fallback to search, promote warmup async",
			limit:  2,
			offset: 0,
			setup: func(m *mockSearchStore) {
				m.getPopularFn = func(_ context.Context, _ int64, _ int64) ([]storage.EventModel, error) {
					return nil, errors.New("redis down")
				}
				m.searchFn = func(_ context.Context, _ storage.SearchFilterDTO) ([]storage.EventModel, error) {
					return []storage.EventModel{{ID: 1}, {ID: 2}}, nil
				}
				m.promoteFn = func(_ context.Context, _ []storage.EventModel) error { return nil }
			},
			wantErr:         false,
			wantLen:         2,
			wantSearchCalls: 1,
			wantPromote:     true,
		},
		{
			name:   "redis miss (empty, no error) -> fallback to search, search error -> return error",
			limit:  5,
			offset: 10,
			setup: func(m *mockSearchStore) {
				m.getPopularFn = func(_ context.Context, _ int64, _ int64) ([]storage.EventModel, error) {
					return []storage.EventModel{}, nil
				}
				m.searchFn = func(_ context.Context, _ storage.SearchFilterDTO) ([]storage.EventModel, error) {
					return nil, errors.New("elastic down")
				}
			},
			wantErr:         true,
			wantLen:         0,
			wantSearchCalls: 1,
			wantPromote:     false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := newMockStore()
			tc.setup(m)

			srv := NewSearchService(cfg, m, m, m, zap.NewNop())

			got, err := srv.GetPopularEvents(context.Background(), tc.limit, tc.offset)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, got, tc.wantLen)
			}

			m.mu.Lock()
			require.Equal(t, tc.limit, m.lastLimit)
			require.Equal(t, tc.offset, m.lastOffset)
			require.Equal(t, tc.wantSearchCalls, m.searchCalls)
			m.mu.Unlock()

			if tc.wantSearchCalls > 0 {
				select {
				case f := <-m.searchCh:
					require.Equal(t, tc.limit, f.Limit)
					require.Equal(t, tc.offset, f.Offset)
				case <-time.After(200 * time.Millisecond):
					t.Fatalf("expected Search to be called with filter, but did not capture it")
				}
			}

			if tc.wantPromote {
				select {
				case call := <-m.promoteCh:
					require.Len(t, call.events, tc.wantLen)
					_, hasDeadline := call.ctx.Deadline()
					require.True(t, hasDeadline)
				case <-time.After(700 * time.Millisecond):
					t.Fatalf("expected PromoteToPopular warmup async call, but did not receive it")
				}
			} else {
				select {
				case <-m.promoteCh:
					t.Fatalf("PromoteToPopular should NOT be called")
				case <-time.After(150 * time.Millisecond):
				}
			}
		})
	}
}

func TestSearchService_SetComment_Table(t *testing.T) {
	t.Parallel()

	cfg := Config{
		JWTSecret:            "test",
		CommentsDefaultLimit: 20,
		CommentsMaxLimit:     100,
		CommentsMaxOffset:    100000,
	}

	tests := []struct {
		name    string
		setup   func(m *mockSearchStore)
		in      domain.UpsertComment
		wantErr bool
	}{
		{
			name: "commentRepo.CreateComment error -> return error",
			setup: func(m *mockSearchStore) {
				m.createCommentFn = func(_ context.Context, _ sqlc.CreateCommentParams) (sqlc.CreateCommentRow, error) {
					return sqlc.CreateCommentRow{}, errors.New("db down")
				}
			},
			in: domain.UpsertComment{
				EventID: 11,
				UserID:  22,
				Nick:    "pavlo",
				Text:    "ok",
				Rating:  48,
			},
			wantErr: true,
		},
		{
			name: "ok -> CreateComment called with mapped params",
			setup: func(m *mockSearchStore) {
				m.createCommentFn = func(_ context.Context, arg sqlc.CreateCommentParams) (sqlc.CreateCommentRow, error) {
					return sqlc.CreateCommentRow{
						ID:           1,
						EventID:      arg.EventID,
						UserID:       arg.UserID,
						Nick:         arg.Nick,
						Text:         arg.Text,
						RatingTenths: arg.RatingTenths,
						CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
					}, nil
				}
			},
			in: domain.UpsertComment{
				EventID: 1,
				UserID:  2,
				Nick:    "pavlo",
				Text:    "nice",
				Rating:  50,
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := newMockStore()
			tc.setup(m)

			srv := NewSearchService(cfg, m, m, m, zap.NewNop())

			err := srv.SetComment(context.Background(), tc.in)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			select {
			case got := <-m.createCommentCh:
				require.Equal(t, tc.in.EventID, got.EventID)
				require.Equal(t, tc.in.UserID, got.UserID)
				require.Equal(t, tc.in.Nick, got.Nick)
				require.Equal(t, tc.in.Text, got.Text)
				require.Equal(t, tc.in.Rating, got.RatingTenths)
			case <-time.After(200 * time.Millisecond):
				t.Fatalf("expected CreateComment to be called")
			}
		})
	}
}

func TestSearchService_GetComments_Table(t *testing.T) {
	t.Parallel()

	cfg := Config{
		JWTSecret:            "test",
		CommentsDefaultLimit: 20,
		CommentsMaxLimit:     50,
		CommentsMaxOffset:    1000,
	}

	now := time.Now().UTC()

	tests := []struct {
		name        string
		setup       func(m *mockSearchStore)
		in          domain.ListComments
		wantErr     bool
		wantErrIs   error
		wantLimit   int64
		wantOffset  int64
		wantEventID int64
		wantLen     int
	}{
		{
			name: "limit=0 -> use default limit, ok empty",
			setup: func(m *mockSearchStore) {
				m.listCommentsFn = func(_ context.Context, _ sqlc.ListCommentsByEventParams) ([]sqlc.EventComment, error) {
					return []sqlc.EventComment{}, nil
				}
			},
			in:          domain.ListComments{EventID: 10, Limit: 0, Offset: 0},
			wantErr:     false,
			wantLimit:   20,
			wantOffset:  0,
			wantEventID: 10,
			wantLen:     0,
		},
		{
			name: "limit > max -> ErrInvalidPagination",
			setup: func(m *mockSearchStore) {
				m.listCommentsFn = func(_ context.Context, _ sqlc.ListCommentsByEventParams) ([]sqlc.EventComment, error) {
					t.Fatalf("ListCommentsByEvent must not be called on invalid pagination")
					return nil, nil
				}
			},
			in:        domain.ListComments{EventID: 1, Limit: 999, Offset: 0},
			wantErr:   true,
			wantErrIs: ErrInvalidPagination,
		},
		{
			name: "offset > max -> ErrInvalidPagination",
			setup: func(m *mockSearchStore) {
				m.listCommentsFn = func(_ context.Context, _ sqlc.ListCommentsByEventParams) ([]sqlc.EventComment, error) {
					t.Fatalf("ListCommentsByEvent must not be called on invalid pagination")
					return nil, nil
				}
			},
			in:        domain.ListComments{EventID: 1, Limit: 10, Offset: 50000},
			wantErr:   true,
			wantErrIs: ErrInvalidPagination,
		},
		{
			name: "ok -> maps sqlc.EventComment to domain.Comment",
			setup: func(m *mockSearchStore) {
				m.listCommentsFn = func(_ context.Context, arg sqlc.ListCommentsByEventParams) ([]sqlc.EventComment, error) {
					return []sqlc.EventComment{
						{
							ID:           7,
							EventID:      arg.EventID,
							UserID:       101,
							Nick:         "pavlo",
							Text:         "nice",
							RatingTenths: 48,
							CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
						},
					}, nil
				}
			},
			in:          domain.ListComments{EventID: 33, Limit: 5, Offset: 10},
			wantErr:     false,
			wantLimit:   5,
			wantOffset:  10,
			wantEventID: 33,
			wantLen:     1,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := newMockStore()
			tc.setup(m)

			srv := NewSearchService(cfg, m, m, m, zap.NewNop())

			items, err := srv.GetComments(context.Background(), tc.in)
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrIs != nil {
					require.ErrorIs(t, err, tc.wantErrIs)
				}
				return
			}

			require.NoError(t, err)
			require.Len(t, items, tc.wantLen)

			select {
			case got := <-m.listCommentsCh:
				require.Equal(t, tc.wantEventID, got.EventID)
				require.Equal(t, tc.wantLimit, got.PageLimit)
				require.Equal(t, tc.wantOffset, got.PageOffset)
			case <-time.After(200 * time.Millisecond):
				t.Fatalf("expected ListCommentsByEvent to be called")
			}

			if tc.wantLen == 1 {
				require.Equal(t, int64(7), items[0].ID)
				require.Equal(t, tc.wantEventID, items[0].EventID)
				require.Equal(t, int64(101), items[0].UserID)
				require.Equal(t, "pavlo", items[0].Nick)
				require.Equal(t, "nice", items[0].Text)
				require.Equal(t, int64(48), items[0].Rating)
				require.True(t, items[0].CreatedAt.Equal(now))
			}
		})
	}
}

func TestSearchService_GetRating_Table(t *testing.T) {
	t.Parallel()

	cfg := Config{
		JWTSecret:            "test",
		CommentsDefaultLimit: 20,
		CommentsMaxLimit:     100,
		CommentsMaxOffset:    100000,
	}

	tests := []struct {
		name    string
		setup   func(m *mockSearchStore)
		eventID int64
		wantErr bool
		want    domain.Rating
	}{
		{
			name: "repo error -> return error",
			setup: func(m *mockSearchStore) {
				m.getRatingFn = func(_ context.Context, _ int64) (sqlc.GetRatingByEventRow, error) {
					return sqlc.GetRatingByEventRow{}, errors.New("db down")
				}
			},
			eventID: 10,
			wantErr: true,
		},
		{
			name: "ok -> map to domain.Rating",
			setup: func(m *mockSearchStore) {
				m.getRatingFn = func(_ context.Context, eventID int64) (sqlc.GetRatingByEventRow, error) {
					return sqlc.GetRatingByEventRow{
						EventID:   eventID,
						AvgTenths: 49,
						Count:     123,
					}, nil
				}
			},
			eventID: 77,
			wantErr: false,
			want: domain.Rating{
				EventID: 77,
				Avg:     49,
				Count:   123,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := newMockStore()
			tc.setup(m)

			srv := NewSearchService(cfg, m, m, m, zap.NewNop())

			got, err := srv.GetRating(context.Background(), tc.eventID)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)

			select {
			case id := <-m.getRatingCh:
				require.Equal(t, tc.eventID, id)
			case <-time.After(200 * time.Millisecond):
				t.Fatalf("expected GetRatingByEvent to be called")
			}
		})
	}
}
