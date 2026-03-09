package storage

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	redismock "github.com/go-redis/redismock/v9"
	"github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestFavouriteRepo_GetPopular_Table(t *testing.T) {
	t.Parallel()

	type want struct {
		count int
		ids   []int64
	}

	tests := []struct {
		name      string
		limit     int64
		offset    int64
		mockRedis func(m redismock.ClientMock)
		wantErr   bool
		want      want
	}{
		{
			name:   "redis ZRevRange error",
			limit:  10,
			offset: 0,
			mockRedis: func(m redismock.ClientMock) {
				m.ExpectZRevRange(prefixPopular, 0, 9).SetErr(errors.New("redis down"))
			},
			wantErr: true,
		},
		{
			name:   "no ids -> empty slice",
			limit:  10,
			offset: 0,
			mockRedis: func(m redismock.ClientMock) {
				m.ExpectZRevRange(prefixPopular, 0, 9).SetVal([]string{})
			},
			wantErr: false,
			want:    want{count: 0},
		},
		{
			name:   "MGet error",
			limit:  3,
			offset: 5,
			mockRedis: func(m redismock.ClientMock) {
				m.ExpectZRevRange(prefixPopular, 5, 7).SetVal([]string{"10", "11", "12"})
				m.ExpectMGet(prefixData+"10", prefixData+"11", prefixData+"12").
					SetErr(errors.New("mget failed"))
			},
			wantErr: true,
		},
		{
			name:   "payload nil and non-string are skipped",
			limit:  3,
			offset: 0,
			mockRedis: func(m redismock.ClientMock) {
				m.ExpectZRevRange(prefixPopular, 0, 2).SetVal([]string{"1", "2", "3"})
				m.ExpectMGet(prefixData+"1", prefixData+"2", prefixData+"3").
					SetVal([]any{nil, 123, `{"id":3}`})
			},
			wantErr: false,
			want:    want{count: 1, ids: []int64{3}},
		},
		{
			name:   "bad json is skipped, good json returned",
			limit:  2,
			offset: 0,
			mockRedis: func(m redismock.ClientMock) {
				m.ExpectZRevRange(prefixPopular, 0, 1).SetVal([]string{"7", "8"})
				m.ExpectMGet(prefixData+"7", prefixData+"8").
					SetVal([]any{`{BAD_JSON`, `{"id":8}`})
			},
			wantErr: false,
			want:    want{count: 1, ids: []int64{8}},
		},
		{
			name:   "two valid docs returned in ranking order",
			limit:  2,
			offset: 1,
			mockRedis: func(m redismock.ClientMock) {
				m.ExpectZRevRange(prefixPopular, 1, 2).SetVal([]string{"21", "20"})
				m.ExpectMGet(prefixData+"21", prefixData+"20").
					SetVal([]any{`{"id":21}`, `{"id":20}`})
			},
			wantErr: false,
			want:    want{count: 2, ids: []int64{21, 20}},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rdb, mock := redismock.NewClientMock()

			if tc.mockRedis != nil {
				tc.mockRedis(mock)
			}

			repo := NewSearchRepo(rdb, zap.NewNop(), 24*time.Hour)

			got, err := repo.GetPopular(context.Background(), tc.limit, tc.offset)
			if tc.wantErr {
				require.Error(t, err)
				require.NoError(t, mock.ExpectationsWereMet())
				return
			}

			require.NoError(t, err)
			require.Len(t, got, tc.want.count)

			if tc.want.count > 0 {
				gotIDs := make([]int64, 0, len(got))
				for _, d := range got {
					gotIDs = append(gotIDs, d.ID)
				}
				require.Equal(t, tc.want.ids, gotIDs)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestEventRepo_Search_Table_HTTPServer(t *testing.T) {
	t.Parallel()

	type tc struct {
		name         string
		filter       SearchFilterDTO
		status       int
		responseBody string

		wantCount    int
		wantContains []string
	}

	tests := []tc{
		{
			name: "query only -> multi_match present",
			filter: SearchFilterDTO{
				Query:  "rock",
				Limit:  10,
				Offset: 0,
			},
			status:       200,
			responseBody: esSearchResponse(`{"id":1}`, `{"id":2}`),
			wantCount:    2,
			wantContains: []string{`"multi_match"`, `"name"`, `"info_header"`, `"info_body"`, `"fuzziness":"AUTO"`},
		},
		{
			name: "filters performer+venue+date range present",
			filter: SearchFilterDTO{
				PerformerID: 77,
				VenueID:     12,
				DateFrom:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				DateTo:      time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
				Limit:       5,
				Offset:      10,
			},
			status:       200,
			responseBody: esSearchResponse(`{"id":3}`),
			wantCount:    1,
			wantContains: []string{
				`"term":{"performer_id":77}`,
				`"term":{"venue_id":12}`,
				`"range":{"date_start"`,
			},
		},
		{
			name: "not found -> empty slice, no error",
			filter: SearchFilterDTO{
				Query:  "whatever",
				Limit:  10,
				Offset: 0,
			},
			status:       404,
			responseBody: `{"error":{"type":"index_not_found_exception"},"status":404}`,
			wantCount:    0,
		},
		{
			name: "bad _source json in one hit is skipped",
			filter: SearchFilterDTO{
				Query:  "x",
				Limit:  10,
				Offset: 0,
			},
			status:       200,
			responseBody: esSearchResponse(`{"id":"bad"}`, `{"id":9}`),
			wantCount:    1,
			wantContains: []string{`"multi_match"`},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var captured string

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				_ = r.Body.Close()
				if len(body) > 0 {
					captured = string(body)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer srv.Close()

			es, err := elastic.NewClient(
				elastic.SetURL(srv.URL),
				elastic.SetSniff(false),
				elastic.SetHealthcheck(false),
			)
			require.NoError(t, err)

			repo := NewFavouriteStore(es, zap.NewNop())

			got, err := repo.Search(context.Background(), tt.filter)
			require.NoError(t, err)
			require.Len(t, got, tt.wantCount)

			for _, frag := range tt.wantContains {
				require.Contains(t, captured, frag)
			}
		})
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
