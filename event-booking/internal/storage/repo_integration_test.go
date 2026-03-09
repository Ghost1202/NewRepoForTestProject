package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	gsqlite "github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func requireIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("INTEGRATION_TESTS") != "1" {
		t.Skip("Set INTEGRATION_TESTS=1 to run integration tests")
	}
}

func setupTestRepos(t *testing.T) (*TicketRepo, *CacheRepo, *miniredis.Miniredis, func()) {
	t.Helper()

	mr, err := miniredis.Run()
	require.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	db, err := gorm.Open(gsqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&TicketModel{})
	require.NoError(t, err)

	tr := NewTicketRepo(db, zap.NewNop())
	cr := NewCacheRepo(rdb, zap.NewNop())

	cleanup := func() {
		_ = rdb.Close()
		mr.Close()
	}

	return tr, cr, mr, cleanup
}

func TestIntegration_TicketRepo_GetByID_Scenarios(t *testing.T) {
	requireIntegration(t)
	tr, _, _, cleanup := setupTestRepos(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("success_find_ticket", func(t *testing.T) {
		expected := TicketModel{ID: 100, EventID: 1, Status: StatusCreated, Price: 500}
		require.NoError(t, tr.db.Create(&expected).Error)

		got, err := tr.GetTicketByID(ctx, 100)
		require.NoError(t, err)
		require.Equal(t, expected.ID, got.ID)
		require.Equal(t, expected.Price, got.Price)
	})

	t.Run("error_not_found", func(t *testing.T) {
		_, err := tr.GetTicketByID(ctx, 9999)
		require.Error(t, err)
		require.Contains(t, err.Error(), "record not found")
	})
}

func TestIntegration_TicketRepo_GetByEvent_Filtering(t *testing.T) {
	requireIntegration(t)
	tr, _, _, cleanup := setupTestRepos(t)
	defer cleanup()

	ctx := context.Background()
	eventID := int64(77)

	tickets := []TicketModel{
		{ID: 1, EventID: eventID, Status: StatusCreated},
		{ID: 2, EventID: eventID, Status: StatusCreated},
		{ID: 3, EventID: eventID, Status: "SOLD"},
		{ID: 4, EventID: 88, Status: StatusCreated},
	}
	for _, tk := range tickets {
		require.NoError(t, tr.db.Create(&tk).Error)
	}

	got, err := tr.GetTicketsByEvent(ctx, eventID)
	require.NoError(t, err)
	require.Len(t, got, 2, "Should only return CREATED tickets for specified event")

	for _, tk := range got {
		require.Equal(t, StatusCreated, tk.Status)
		require.Equal(t, eventID, tk.EventID)
	}
}

func TestIntegration_CacheRepo_Lock_Concurrency(t *testing.T) {
	requireIntegration(t)
	_, cr, _, cleanup := setupTestRepos(t)
	defer cleanup()

	ctx := context.Background()
	ticketID := int64(123)

	t.Run("race_for_lock", func(t *testing.T) {
		ok, err := cr.Lock(ctx, ticketID, 1, 5*time.Second)
		require.NoError(t, err)
		require.True(t, ok)

		ok2, err := cr.Lock(ctx, ticketID, 2, 5*time.Second)
		require.NoError(t, err)
		require.False(t, ok2, "Second user should not acquire existing lock")
	})
}

func TestIntegration_MGet_Values_Detailed(t *testing.T) {
	requireIntegration(t)
	_, cr, _, cleanup := setupTestRepos(t)
	defer cleanup()

	ctx := context.Background()

	require.NoError(t, cr.redis.Set(ctx, TicketLockPrefix+"10", "user_A", time.Minute).Err())
	require.NoError(t, cr.redis.Set(ctx, TicketLockPrefix+"30", "user_C", time.Minute).Err())

	keys := []string{TicketLockPrefix + "10", TicketLockPrefix + "20", TicketLockPrefix + "30"}
	res, err := cr.MGet(ctx, keys)

	require.NoError(t, err)
	require.Len(t, res, 3)
	require.Equal(t, "user_A", res[0])
	require.Nil(t, res[1], "Key 20 should be nil as it wasn't set")
	require.Equal(t, "user_C", res[2])
}

func TestIntegration_LockExpiration_Detailed(t *testing.T) {
	requireIntegration(t)
	_, cr, mr, cleanup := setupTestRepos(t)
	defer cleanup()

	ctx := context.Background()
	ticketID := int64(777)

	ok, err := cr.Lock(ctx, ticketID, 1, 200*time.Millisecond)
	require.NoError(t, err)
	require.True(t, ok)

	res, err := cr.MGet(ctx, []string{TicketLockPrefix + "777"})
	require.NoError(t, err)
	require.NotNil(t, res[0])

	mr.FastForward(300 * time.Millisecond)

	ok2, err := cr.Lock(ctx, ticketID, 2, time.Second)
	require.NoError(t, err)
	require.True(t, ok2, "Lock should be acquirable after expiration")
}

func TestIntegration_Complex_AvailableTickets_Lifecycle(t *testing.T) {
	requireIntegration(t)
	tr, cr, _, cleanup := setupTestRepos(t)
	defer cleanup()

	ctx := context.Background()
	eventID := int64(1001)

	tickets := []TicketModel{
		{ID: 10, EventID: eventID, Status: StatusCreated},
		{ID: 11, EventID: eventID, Status: StatusCreated},
		{ID: 12, EventID: eventID, Status: StatusCreated},
	}
	for _, tk := range tickets {
		require.NoError(t, tr.db.Create(&tk).Error)
	}

	_, err := cr.Lock(ctx, 10, 888, time.Minute)
	require.NoError(t, err)
	_, err = cr.Lock(ctx, 11, 999, time.Minute)
	require.NoError(t, err)

	available, err := getAvailable(ctx, tr, cr, eventID)
	require.NoError(t, err)
	require.Len(t, available, 1)
	require.Equal(t, int64(12), available[0].ID)

	require.NoError(t, cr.redis.Del(ctx, TicketLockPrefix+"10").Err())

	availableAfterDel, err := getAvailable(ctx, tr, cr, eventID)
	require.NoError(t, err)
	require.Len(t, availableAfterDel, 2)
}
