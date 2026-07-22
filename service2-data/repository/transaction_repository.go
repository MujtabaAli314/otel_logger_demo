package repository

import (
	"github.com/oteldemo/service2-data/types"
	"gorm.io/gorm"
)

// TransactionRepository abstracts persistence access for transactions.
type TransactionRepository interface {
	ListByUserID(userID uint, limit, offset int) ([]types.Transaction, error)
}

type gormTransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &gormTransactionRepository{db: db}
}

func (r *gormTransactionRepository) ListByUserID(userID uint, limit, offset int) ([]types.Transaction, error) {
	var txs []types.Transaction
	err := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&txs).Error
	if err != nil {
		return nil, err
	}
	return txs, nil
}
