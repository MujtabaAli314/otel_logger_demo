package repository

import (
	"github.com/oteldemo/service2-data/types"
	"gorm.io/gorm"
)

// UserRepository abstracts persistence access for users.
type UserRepository interface {
	GetByID(id uint) (*types.User, error)
}

type gormUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) GetByID(id uint) (*types.User, error) {
	var u types.User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
