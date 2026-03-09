package storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	redismock "github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	qSelectTickets = `SELECT .* FROM "tickets" WHERE event_id = \$1 AND status = \$2`
	qSelectByID    = `SELECT .* FROM "tickets" WHERE "tickets"\."id" = \$1.*`
)

func newMockGormDB(t *testing.T) (db *gorm.DB, mock sqlmock.Sqlmock, sqlDB *sql.DB) {
	t.Helper()

	var err error
	sqlDB, mock, err = sqlmock.New()
	require.NoError(t, err)

	db, err = gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	return db, mock, sqlDB
}

func getAvailable(ctx context.Context, tr *TicketRepo, cr *CacheRepo, eventID int64) ([]TicketModel, error) {
	models, err := tr.GetTicketsByEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if len(models) == 0 {
		return []TicketModel{}, nil
	}

	keys := make([]string, len(models))
	for i, m := range models {
		keys[i] = TicketLockKey(m.ID)
	}

	locks, err := cr.MGet(ctx, keys)
	if err != nil {
		return nil, err
	}

	result := make([]TicketModel, 0, len(models))
	for i, lock := range locks {
		if lock == nil {
			result = append(result, models[i])
		}
	}
	return result, nil
}

func TestTicketRepo_GetTicketsByEvent_Table(t *testing.T) {
	tests := []struct {
		name      string
		eventID   int64
		prepare   func(m sqlmock.Sqlmock)
		wantCount int
		wantErr   bool
	}{
		{
			name:    "success_multiple_tickets",
			eventID: 1,
			prepare: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "event_id", "status"}).
					AddRow(1, 1, StatusCreated).
					AddRow(2, 1, StatusCreated)
				m.ExpectQuery(qSelectTickets).
					WithArgs(int64(1), StatusCreated).
					WillReturnRows(rows)
			},
			wantCount: 2,
		},
		{
			name:    "db_error",
			eventID: 3,
			prepare: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(qSelectTickets).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, raw := newMockGormDB(t)
			defer raw.Close()

			repo := NewTicketRepo(db, zap.NewNop())
			tt.prepare(mock)

			res, err := repo.GetTicketsByEvent(context.Background(), tt.eventID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, res, tt.wantCount)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTicketRepo_GetTicketByID_Scenarios(t *testing.T) {
	tests := []struct {
		name     string
		ticketID int64
		prepare  func(m sqlmock.Sqlmock)
		wantErr  bool
	}{
		{
			name:     "found_successfully",
			ticketID: 10,
			prepare: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(qSelectByID).
					WithArgs(int64(10), 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "price"}).AddRow(10, 100))
			},
			wantErr: false,
		},
		{
			name:     "record_not_found",
			ticketID: 99,
			prepare: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(qSelectByID).
					WithArgs(int64(99), 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, raw := newMockGormDB(t)
			defer raw.Close()

			repo := NewTicketRepo(db, zap.NewNop())
			tt.prepare(mock)

			res, err := repo.GetTicketByID(context.Background(), tt.ticketID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.ticketID, res.ID)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCacheRepo_Lock_Scenarios(t *testing.T) {
	tests := []struct {
		name     string
		ticketID int64
		userID   int64
		prepare  func(m redismock.ClientMock, key string)
		want     bool
		wantErr  bool
	}{
		{
			name:     "lock_acquired",
			ticketID: 1, userID: 100,
			prepare: func(m redismock.ClientMock, key string) {
				m.ExpectSetNX(key, int64(100), time.Minute).SetVal(true)
			},
			want: true,
		},
		{
			name:     "already_locked",
			ticketID: 2, userID: 200,
			prepare: func(m redismock.ClientMock, key string) {
				m.ExpectSetNX(key, int64(200), time.Minute).SetVal(false)
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rdb, mock := redismock.NewClientMock()
			repo := NewCacheRepo(rdb, zap.NewNop())
			key := TicketLockKey(tt.ticketID)
			tt.prepare(mock, key)

			ok, err := repo.Lock(context.Background(), tt.ticketID, tt.userID, time.Minute)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, ok)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCacheRepo_MGet_Extended(t *testing.T) {
	tests := []struct {
		name    string
		keys    []string
		prepare func(m redismock.ClientMock)
		want    []interface{}
		wantErr bool
	}{
		{
			name: "mixed_results",
			keys: []string{"k1", "k2", "k3"},
			prepare: func(m redismock.ClientMock) {
				m.ExpectMGet("k1", "k2", "k3").SetVal([]interface{}{"u1", nil, "u3"})
			},
			want: []interface{}{"u1", nil, "u3"},
		},
		{
			name: "redis_error",
			keys: []string{"k1"},
			prepare: func(m redismock.ClientMock) {
				m.ExpectMGet("k1").SetErr(errors.New("redis fail"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rdb, mock := redismock.NewClientMock()
			repo := NewCacheRepo(rdb, zap.NewNop())
			tt.prepare(mock)

			res, err := repo.MGet(context.Background(), tt.keys)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, res)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetAvailable_Orchestration(t *testing.T) {
	t.Run("full_success_flow", func(t *testing.T) {
		db, dbMock, raw := newMockGormDB(t)
		defer raw.Close()
		rdb, rMock := redismock.NewClientMock()

		tr := NewTicketRepo(db, zap.NewNop())
		cr := NewCacheRepo(rdb, zap.NewNop())

		eventID := int64(500)

		dbMock.ExpectQuery(qSelectTickets).
			WithArgs(eventID, StatusCreated).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))

		rMock.ExpectMGet(TicketLockKey(1), TicketLockKey(2)).
			SetVal([]interface{}{"u1", nil})

		got, err := getAvailable(context.Background(), tr, cr, eventID)

		require.NoError(t, err)
		require.Len(t, got, 1)
		require.Equal(t, int64(2), got[0].ID)

		require.NoError(t, dbMock.ExpectationsWereMet())
		require.NoError(t, rMock.ExpectationsWereMet())
	})

	t.Run("empty_db_skips_redis", func(t *testing.T) {
		db, dbMock, raw := newMockGormDB(t)
		defer raw.Close()
		rdb, rMock := redismock.NewClientMock()

		tr := NewTicketRepo(db, zap.NewNop())
		cr := NewCacheRepo(rdb, zap.NewNop())

		dbMock.ExpectQuery(qSelectTickets).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		got, err := getAvailable(context.Background(), tr, cr, 999)
		require.NoError(t, err)
		require.Empty(t, got)

		require.NoError(t, dbMock.ExpectationsWereMet())
		require.NoError(t, rMock.ExpectationsWereMet())
	})
}
