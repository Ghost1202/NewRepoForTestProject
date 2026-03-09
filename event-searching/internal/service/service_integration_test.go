package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/olivere/elastic/v7"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/turtlepavlo/event-searching/internal/storage"
	sqlc "github.com/turtlepavlo/event-searching/internal/storage/postgres/sqlc"
	"go.uber.org/zap"
)

type noopCommentRepo struct{}

func (noopCommentRepo) CreateComment(_ context.Context, _ sqlc.CreateCommentParams) (sqlc.CreateCommentRow, error) {
	panic("CreateComment must not be called in this integration test")
}

func (noopCommentRepo) ListCommentsByEvent(_ context.Context, _ sqlc.ListCommentsByEventParams) ([]sqlc.EventComment, error) {
	panic("ListCommentsByEvent must not be called in this integration test")
}

func (noopCommentRepo) GetRatingByEvent(_ context.Context, _ int64) (sqlc.GetRatingByEventRow, error) {
	panic("GetRatingByEvent must not be called in this integration test")
}

func TestSearchService_GetPopularEvents_Warmup_Integration(t *testing.T) {
	t.Parallel()

	cfg := Config{
		JWTSecret:            "test",
		CommentsDefaultLimit: 20,
		CommentsMaxLimit:     100,
		CommentsMaxOffset:    100000,
	}

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if _, err := io.ReadAll(r.Body); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if err := r.Body.Close(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte(esSearchResponse(`{"id":101}`, `{"id":102}`))); err != nil {
			return
		}
	}))
	defer srv.Close()

	es, err := elastic.NewClient(
		elastic.SetURL(srv.URL),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
	)
	require.NoError(t, err)

	favouriteRepo := storage.NewSearchRepo(rdb, zap.NewNop(), 2*time.Minute)
	eventRepo := storage.NewFavouriteStore(es, zap.NewNop())

	svc := NewSearchService(
		cfg,
		favouriteRepo,
		eventRepo,
		noopCommentRepo{},
		zap.NewNop(),
	)

	events, err := svc.GetPopularEvents(context.Background(), 2, 0)
	require.NoError(t, err)
	require.Len(t, events, 2)

	deadline := time.Now().Add(1 * time.Second)
	for {
		got, gErr := favouriteRepo.GetPopular(context.Background(), 2, 0)
		if gErr == nil && len(got) == 2 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("warmup did not populate redis in time (last err=%v, last len=%d)", gErr, len(got))
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func esSearchResponse(sources ...string) string {
	var b strings.Builder
	b.WriteString(`{"took":1,"timed_out":false,"hits":{"total":{"value":`)
	b.WriteString(strconv.Itoa(len(sources)))
	b.WriteString(`,"relation":"eq"},"max_score":1.0,"hits":[`)

	for i, src := range sources {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(`{"_index":"event_service.public.events","_type":"_doc","_id":"`)
		b.WriteString(strconv.Itoa(i + 1))
		b.WriteString(`","_score":1.0,"_source":`)
		b.WriteString(src)
		b.WriteString("}")
	}

	b.WriteString(`]}}`)
	return b.String()
}
