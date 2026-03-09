package storage

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func setupRedis(t *testing.T, ttl time.Duration) (*CacheRepo, redismock.ClientMock) {
	db, mock := redismock.NewClientMock()
	logger := zaptest.NewLogger(t)
	repo := NewCacheRepo(db, logger, ttl)
	return repo, mock
}

func TestNewCacheRepo_Initialization(t *testing.T) {
	db, _ := redismock.NewClientMock()

	t.Run("With Valid Logger", func(t *testing.T) {
		logger := zap.NewExample()
		repo := NewCacheRepo(db, logger, time.Minute)
		assert.NotNil(t, repo)
		assert.Equal(t, logger, repo.log)
		assert.Equal(t, time.Minute, repo.defaultTTL)
	})

	t.Run("With Nil Logger", func(t *testing.T) {
		repo := NewCacheRepo(db, nil, time.Second*30)
		assert.NotNil(t, repo)
		assert.NotNil(t, repo.log)
		assert.Equal(t, time.Second*30, repo.defaultTTL)
	})
}

func TestCacheRepo_SetCode(t *testing.T) {
	defaultTTL := time.Minute * 5
	ctx := context.Background()

	tests := []struct {
		name      string
		key       string
		value     int64
		mockSetup func(mock redismock.ClientMock, key string, val int64)
		wantErr   bool
	}{
		{
			name:  "Success Set",
			key:   "auth:user:1",
			value: 123456,
			mockSetup: func(mock redismock.ClientMock, key string, val int64) {
				mock.ExpectSet(key, val, defaultTTL).SetVal("OK")
			},
			wantErr: false,
		},
		{
			name:  "Redis Connection Error",
			key:   "auth:user:2",
			value: 999999,
			mockSetup: func(mock redismock.ClientMock, key string, val int64) {
				mock.ExpectSet(key, val, defaultTTL).SetErr(errors.New("connection refused"))
			},
			wantErr: true,
		},
		{
			name:  "Zero Value",
			key:   "auth:zero",
			value: 0,
			mockSetup: func(mock redismock.ClientMock, key string, val int64) {
				mock.ExpectSet(key, val, defaultTTL).SetVal("OK")
			},
			wantErr: false,
		},
		{
			name:  "Negative Value",
			key:   "auth:neg",
			value: -555,
			mockSetup: func(mock redismock.ClientMock, key string, val int64) {
				mock.ExpectSet(key, val, defaultTTL).SetVal("OK")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupRedis(t, defaultTTL)
			tt.mockSetup(mock, tt.key, tt.value)
			err := repo.SetCode(ctx, tt.key, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCacheRepo_GetCode(t *testing.T) {
	ctx := context.Background()
	defaultTTL := time.Minute

	tests := []struct {
		name      string
		key       string
		mockSetup func(mock redismock.ClientMock, key string)
		wantVal   int64
		wantErr   error
	}{
		{
			name: "Success Get",
			key:  "auth:valid",
			mockSetup: func(mock redismock.ClientMock, key string) {
				mock.ExpectGet(key).SetVal("55555")
			},
			wantVal: 55555,
			wantErr: nil,
		},
		{
			name: "Key Not Found",
			key:  "auth:missing",
			mockSetup: func(mock redismock.ClientMock, key string) {
				mock.ExpectGet(key).RedisNil()
			},
			wantVal: 0,
			wantErr: ErrKeyNotFound,
		},
		{
			name: "Redis Error",
			key:  "auth:broken",
			mockSetup: func(mock redismock.ClientMock, key string) {
				mock.ExpectGet(key).SetErr(errors.New("timeout"))
			},
			wantVal: 0,
			wantErr: errors.New("timeout"),
		},
		{
			name: "Data Corruption",
			key:  "auth:corrupt",
			mockSetup: func(mock redismock.ClientMock, key string) {
				mock.ExpectGet(key).SetVal("not-a-number")
			},
			wantVal: 0,
			wantErr: errors.New("strconv.ParseInt: parsing"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupRedis(t, defaultTTL)
			tt.mockSetup(mock, tt.key)
			val, err := repo.GetCode(ctx, tt.key)
			if tt.wantErr != nil {
				assert.Error(t, err)
				if tt.wantErr == ErrKeyNotFound {
					assert.Equal(t, ErrKeyNotFound, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantVal, val)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCacheRepo_DeleteCode(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name      string
		key       string
		mockSetup func(mock redismock.ClientMock, key string)
		wantErr   bool
	}{
		{
			name: "Success Delete",
			key:  "del:existing",
			mockSetup: func(mock redismock.ClientMock, key string) {
				mock.ExpectDel(key).SetVal(1)
			},
			wantErr: false,
		},
		{
			name: "Success Delete Missing",
			key:  "del:missing",
			mockSetup: func(mock redismock.ClientMock, key string) {
				mock.ExpectDel(key).SetVal(0)
			},
			wantErr: false,
		},
		{
			name: "Delete Error",
			key:  "del:error",
			mockSetup: func(mock redismock.ClientMock, key string) {
				mock.ExpectDel(key).SetErr(errors.New("io error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := setupRedis(t, time.Minute)
			tt.mockSetup(mock, tt.key)
			err := repo.DeleteCode(ctx, tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCacheRepo_ContextCancellation(t *testing.T) {
	repo, mock := setupRedis(t, time.Minute)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	t.Run("SetCode Context Cancelled", func(t *testing.T) {
		err := repo.SetCode(ctx, "key", 123)
		assert.Error(t, err)
	})

	t.Run("GetCode Context Cancelled", func(t *testing.T) {
		_, err := repo.GetCode(ctx, "key")
		assert.Error(t, err)
	})

	t.Run("DeleteCode Context Cancelled", func(t *testing.T) {
		err := repo.DeleteCode(ctx, "key")
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCacheRepo_Concurrency(t *testing.T) {
	repo, mock := setupRedis(t, time.Minute)
	ctx := context.Background()
	workers := 20
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		mock.ExpectSet("concurrency-key", int64(100), time.Minute).SetVal("OK")
	}

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			_ = repo.SetCode(ctx, "concurrency-key", 100)
		}()
	}

	wg.Wait()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCacheRepo_FullFlow(t *testing.T) {
	repo, mock := setupRedis(t, time.Minute*10)
	ctx := context.Background()
	key := "flow:user:1"
	val := int64(8888)

	mock.ExpectSet(key, val, time.Minute*10).SetVal("OK")
	err := repo.SetCode(ctx, key, val)
	assert.NoError(t, err)

	mock.ExpectGet(key).SetVal("8888")
	retrievedVal, err := repo.GetCode(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, val, retrievedVal)

	mock.ExpectDel(key).SetVal(1)
	err = repo.DeleteCode(ctx, key)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func BenchmarkCacheRepo_SetCode(b *testing.B) {
	db, mock := redismock.NewClientMock()
	repo := NewCacheRepo(db, zap.NewNop(), time.Minute)
	ctx := context.Background()
	mock.ExpectSet("bench", int64(1), time.Minute).SetVal("OK")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectSet("bench", int64(1), time.Minute).SetVal("OK")
		_ = repo.SetCode(ctx, "bench", 1)
	}
}

func BenchmarkCacheRepo_GetCode(b *testing.B) {
	db, mock := redismock.NewClientMock()
	repo := NewCacheRepo(db, zap.NewNop(), time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectGet("bench").SetVal("100")
		_, _ = repo.GetCode(ctx, "bench")
	}
}
