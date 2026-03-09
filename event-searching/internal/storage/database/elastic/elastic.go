package elastic

import (
	"context"
	"fmt"
	"net/http"

	"github.com/olivere/elastic/v7"
	"github.com/turtlepavlo/event-searching/internal/lib/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func New(ctx context.Context, cfg Config, log *zap.Logger) (*elastic.Client, error) {
	const op = "storage.database.elastic.New"
	tracer := otel.Tracer("event-searching/internal/storage/database/elastic")

	url := fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port)

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "elasticsearch"),
			attribute.String("db.url", url),
		),
	)
	defer span.End()

	log = telemetry.WithTrace(ctx, log)

	httpClient := &http.Client{
		Timeout: cfg.DialTimeout,
	}

	client, err := elastic.NewClient(
		elastic.SetURL(url),
		elastic.SetBasicAuth(cfg.User, cfg.Password),
		elastic.SetHealthcheck(cfg.HealthcheckEnabled),
		elastic.SetSniff(cfg.SniffEnabled),
		elastic.SetHttpClient(httpClient),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create elastic client")

		log.Error("failed to connect to elasticsearch",
			zap.String("op", op),
			zap.Error(err),
			zap.String("url", url),
		)

		return nil, err
	}

	_, _, err = client.Ping(url).Do(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "elastic ping failed")

		log.Error("elasticsearch ping failed",
			zap.String("op", op),
			zap.Error(err),
		)

		return nil, err
	}

	log.Info("successfully connected to elasticsearch",
		zap.String("url", url),
	)

	return client, nil
}
