package usecase

import (
	"errors"

	"github.com/oteldemo/service2-data/repository"
	"github.com/oteldemo/service2-data/types"
	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

// UserUsecase contains the business rules around fetching users.
type UserUsecase interface {
	GetUser(params types.GetUserParams) (*types.UserResponse, error)
}

type userUsecase struct {
	users repository.UserRepository
}

func NewUserUsecase(users repository.UserRepository) UserUsecase {
	return &userUsecase{users: users}
}

func (u *userUsecase) GetUser(params types.GetUserParams) (*types.UserResponse, error) {
	user, err := u.users.GetByID(params.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	resp := types.NewUserResponse(*user)
	return &resp, nil
}
