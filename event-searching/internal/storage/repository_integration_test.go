package storage

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestFavouriteRepo_PromoteToPopular_Integration_Redis(t *testing.T) {
	t.Parallel()

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	repo := NewSearchRepo(
		rdb,
		zap.NewNop(),
		2*time.Minute,
	)

	events := []EventModel{
		{ID: 101},
		{ID: 102},
	}

	require.NoError(t, repo.PromoteToPopular(context.Background(), events))

	z, err := mr.ZMembers(prefixPopular)
	require.NoError(t, err)
	require.NotEmpty(t, z)

	require.True(t, mr.Exists(prefixData+"101"))
	require.True(t, mr.Exists(prefixData+"102"))

	ttl101 := mr.TTL(prefixData + "101")
	require.True(t, ttl101 > 0)
}

func TestFavouriteRepo_GetPopular_Integration_Redis(t *testing.T) {
	t.Parallel()

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	_, err := mr.ZAdd(prefixPopular, 10, "1")
	require.NoError(t, err)

	_, err = mr.ZAdd(prefixPopular, 50, "2")
	require.NoError(t, err)

	_, err = mr.ZAdd(prefixPopular, 20, "3")
	require.NoError(t, err)

	require.NoError(t, mr.Set(prefixData+"2", `{"id":2}`))
	require.NoError(t, mr.Set(prefixData+"3", `{"id":3}`))
	require.NoError(t, mr.Set(prefixData+"1", `{"id":1}`))

	repo := NewSearchRepo(
		rdb,
		zap.NewNop(),
		24*time.Hour,
	)

	got, err := repo.GetPopular(context.Background(), 2, 0)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, int64(2), got[0].ID)
	require.Equal(t, int64(3), got[1].ID)
}
