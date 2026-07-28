package repository

import (
	"context"
	"log/slog"

	"github.com/oteldemo/service2-data/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

// TransactionRepository abstracts persistence access for transactions.
type TransactionRepository interface {
	ListByUserID(ctx context.Context, userID uint, limit, offset int) ([]types.Transaction, error)
}

type gormTransactionRepository struct {
	db     *gorm.DB
	logger slog.Logger
}

func NewTransactionRepository(db *gorm.DB, logger slog.Logger) TransactionRepository {
	return &gormTransactionRepository{db: db, logger: logger}
}

func (r *gormTransactionRepository) ListByUserID(ctx context.Context, userID uint, limit, offset int) ([]types.Transaction, error) {
	// Record to the span created by the controller and threaded down via
	// the context. The repository does not start its own span.
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int("repository.user_id", int(userID)),
		attribute.Int("repository.limit", limit),
		attribute.Int("repository.offset", offset),
	)
	r.logger.InfoContext(ctx, "querying transactions from db", "user_id", userID, "limit", limit, "offset", offset)

	var txs []types.Transaction
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&txs).Error
	if err != nil {
		span.RecordError(err)
		r.logger.ErrorContext(ctx, "db query failed", "err", err.Error())
		return nil, err
	}
	r.logger.InfoContext(ctx, "db query completed", "rows", len(txs))
	return txs, nil
}
