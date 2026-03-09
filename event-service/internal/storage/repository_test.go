package storage

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func newMockedRepo(t *testing.T) (*Repository, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Silent),
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	require.NoError(t, err)

	cleanup := func() {
		_ = sqlDB.Close()
	}

	return NewRepository(gdb, zap.NewNop()), mock, cleanup
}

func TestRepository_CreateEvent(t *testing.T) {
	t.Run("success (event + tickets + promos) -> commit", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		ctx := context.Background()
		now := time.Now().UTC()

		ev := Events{
			PerformerID:   10,
			VenueID:       20,
			Name:          "Full Event",
			DateStart:     now.Add(24 * time.Hour),
			SaleStartDate: now.Add(2 * time.Hour),
			MaxPriceCof:   2.0,
			MinPriceCof:   1.0,
		}

		tickets := []Tickets{
			{VenueID: 20, SectorName: "A", RowNumber: 1, SeatNumber: 1, Price: 1000},
			{VenueID: 20, SectorName: "A", RowNumber: 1, SeatNumber: 2, Price: 1000},
		}

		promos := []Promo{
			{Code: "PROMO1", Type: "fixed", Value: 100},
			{Code: "PROMO2", Type: "percent", Value: 10},
		}
		early := []Early{
			{Code: "EARLY1", Type: "percent", Value: 15, ValidUntil: now.Add(48 * time.Hour)},
		}
		bundles := []Bundle{
			{Code: "B2G1", BuyCount: 2, GetCount: 1},
		}

		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "events"`)).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(101)))

		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "tickets"`)).
			WillReturnRows(
				sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2),
			)

		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "promo"`)).
			WillReturnRows(
				sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2),
			)
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "early"`)).
			WillReturnRows(
				sqlmock.NewRows([]string{"id"}).AddRow(3),
			)
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "bundle"`)).
			WillReturnRows(
				sqlmock.NewRows([]string{"id"}).AddRow(4),
			)

		mock.ExpectCommit()

		id, err := repo.CreateEvent(ctx, ev, tickets, promos, early, bundles)
		require.NoError(t, err)
		require.Equal(t, int64(101), id)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success (event only) -> commit", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		ctx := context.Background()
		ev := Events{Name: "Simple Event"}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "events"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(202)))
		mock.ExpectCommit()

		id, err := repo.CreateEvent(ctx, ev, nil, nil, nil, nil)
		require.NoError(t, err)
		require.Equal(t, int64(202), id)
	})

	t.Run("event insert error -> rollback", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		ctx := context.Background()
		ev := Events{Name: "Fail Event"}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "events"`)).
			WillReturnError(errors.New("insert failed"))
		mock.ExpectRollback()

		id, err := repo.CreateEvent(ctx, ev, nil, nil, nil, nil)
		require.Error(t, err)
		require.Equal(t, int64(0), id)
	})
}

func TestRepository_UpdateEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		ctx := context.Background()
		ev := Events{ID: 55, Name: "Updated Name"}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "events"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateEvent(ctx, ev)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update error => rollback", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		ctx := context.Background()
		ev := Events{ID: 77, Name: "Fail Update"}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "events"`)).
			WillReturnError(errors.New("update failed"))
		mock.ExpectRollback()

		err := repo.UpdateEvent(ctx, ev)
		require.Error(t, err)
	})
}

func TestRepository_GetEventPerformerID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		ctx := context.Background()
		eventID := int64(777)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "performer_id" FROM "events" WHERE id = $1`)).
			WithArgs(eventID).
			WillReturnRows(sqlmock.NewRows([]string{"performer_id"}).AddRow(int64(42)))

		got, err := repo.GetEventPerformerID(ctx, eventID)
		require.NoError(t, err)
		require.Equal(t, int64(42), got)
	})

	t.Run("empty result => 0, nil", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		ctx := context.Background()
		eventID := int64(888)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "performer_id" FROM "events"`)).
			WithArgs(eventID).
			WillReturnRows(sqlmock.NewRows([]string{"performer_id"}))

		got, err := repo.GetEventPerformerID(ctx, eventID)
		require.NoError(t, err)
		require.Equal(t, int64(0), got)
	})
}

func TestRepository_UpdateTicketStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		ctx := context.Background()
		ticketID := int64(1001)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tickets"`)).
			WithArgs("sold", ticketID, "CREATED").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		rows, err := repo.UpdateTicketStatus(ctx, ticketID, "CREATED", "sold")
		require.NoError(t, err)
		require.Equal(t, int64(1), rows)
	})

	t.Run("update error", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		ctx := context.Background()
		ticketID := int64(1002)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tickets"`)).
			WillReturnError(errors.New("update failed"))
		mock.ExpectCommit()

		_, err := repo.UpdateTicketStatus(ctx, ticketID, "CREATED", "sold")
		require.Error(t, err)
	})
}

func TestRepository_UpdateTicketStatusAndOwner(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		ctx := context.Background()
		ticketID := int64(2001)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tickets"`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), ticketID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		rows, err := repo.UpdateTicketStatusAndOwner(ctx, ticketID, "CREATED", 0)
		require.NoError(t, err)
		require.Equal(t, int64(1), rows)
	})
}

func TestRepository_UpdateTicketOwner(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		ctx := context.Background()
		ticketID := int64(1)
		fromUserID := int64(10)
		toUserID := int64(20)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tickets" SET "user_id"=$1 WHERE id = $2 AND user_id = $3`)).
			WithArgs(toUserID, ticketID, fromUserID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		rows, err := repo.UpdateTicketOwner(ctx, ticketID, fromUserID, toUserID)
		require.NoError(t, err)
		require.Equal(t, int64(1), rows)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no rows affected -> returns 0, nil", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		ctx := context.Background()
		ticketID := int64(1)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tickets"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		rows, err := repo.UpdateTicketOwner(ctx, ticketID, 10, 20)
		require.NoError(t, err)
		require.Equal(t, int64(0), rows)
	})
}

func TestRepository_AddPromo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		ctx := context.Background()
		promos := []Promo{
			{Code: "A", Type: "fix", Value: 50},
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "promo"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		ids, err := repo.AddPromo(ctx, promos)
		require.NoError(t, err)
		require.Len(t, ids, 1)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_AddEarly(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		ctx := context.Background()
		early := []Early{
			{Code: "E1", Type: "perc", Value: 10, ValidUntil: time.Now().Add(24 * time.Hour)},
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "early"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
		mock.ExpectCommit()

		ids, err := repo.AddEarly(ctx, early)
		require.NoError(t, err)
		require.Len(t, ids, 1)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_AddBundle(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		ctx := context.Background()
		bundles := []Bundle{
			{Code: "B2G1", BuyCount: 2, GetCount: 1},
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "bundle"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
		mock.ExpectCommit()

		ids, err := repo.AddBundle(ctx, bundles)
		require.NoError(t, err)
		require.Len(t, ids, 1)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_GetPromo(t *testing.T) {
	t.Run("success with filter", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		ctx := context.Background()
		eventID := int64(100)
		limit := int64(10)
		offset := int64(0)

		now := time.Now()

		columns := []string{
			"id", "event_id", "code", "type", "value",
			"sector_name", "created_at",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "promo" WHERE event_id = $1 ORDER BY id DESC LIMIT $2`)).
			WithArgs(eventID, limit).
			WillReturnRows(sqlmock.NewRows(columns).
				AddRow(1, 100, "TEST1", "fixed", 100.0, "VIP", now).
				AddRow(2, 100, "TEST2", "percent", 10.0, "", now))

		promos, err := repo.GetPromo(ctx, eventID, limit, offset)
		require.NoError(t, err)
		require.Len(t, promos, 2)
		require.Equal(t, "TEST1", promos[0].Code)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("select error -> error", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		eventID := int64(100)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "promo"`)).
			WillReturnError(errors.New("select failed"))

		_, err := repo.GetPromo(context.Background(), eventID, 0, 0)
		require.Error(t, err)
	})
}

func TestRepository_GetEarly(t *testing.T) {
	t.Run("success with filter", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		ctx := context.Background()
		eventID := int64(100)
		limit := int64(10)
		offset := int64(0)

		now := time.Now()

		earlyColumns := []string{
			"id", "event_id", "code", "type", "value",
			"sector_name", "valid_until", "created_at",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "early" WHERE event_id = $1 ORDER BY id DESC LIMIT $2`)).
			WithArgs(eventID, limit).
			WillReturnRows(sqlmock.NewRows(earlyColumns).
				AddRow(3, 100, "EARLY1", "percent", 15.0, "VIP", now.Add(time.Hour), now))

		early, err := repo.GetEarly(ctx, eventID, limit, offset)
		require.NoError(t, err)
		require.Len(t, early, 1)
		require.Equal(t, "EARLY1", early[0].Code)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_GetBundle(t *testing.T) {
	t.Run("success with filter", func(t *testing.T) {
		repo, mock, cleanup := newMockedRepo(t)
		defer cleanup()

		ctx := context.Background()
		eventID := int64(100)
		limit := int64(10)
		offset := int64(0)

		now := time.Now()

		bundleColumns := []string{
			"id", "event_id", "code", "sector_name",
			"bundle_buy_count", "bundle_get_count", "created_at",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "bundle" WHERE event_id = $1 ORDER BY id DESC LIMIT $2`)).
			WithArgs(eventID, limit).
			WillReturnRows(sqlmock.NewRows(bundleColumns).
				AddRow(4, 100, "B2G1", "VIP", 2, 1, now))

		bundles, err := repo.GetBundle(ctx, eventID, limit, offset)
		require.NoError(t, err)
		require.Len(t, bundles, 1)
		require.Equal(t, "B2G1", bundles[0].Code)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
