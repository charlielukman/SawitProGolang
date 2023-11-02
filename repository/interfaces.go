// This file contains the interfaces for the repository layer.
// The repository layer is responsible for interacting with the database.
// For testing purpose we will generate mock implementations of these
// interfaces using mockgen. See the Makefile for more information.
package repository

import (
	"context"

	"github.com/SawitProRecruitment/UserService/entities"
)

type RepositoryInterface interface {
	CreateUser(ctx context.Context, user entities.User) (userID int, err error)
	IsExistUser(ctx context.Context, user entities.User) (bool, error)
	GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (entities.User, error)
	GetUserByID(ctx context.Context, id int) (entities.User, error)
	UpdateUserLoginSuccess(ctx context.Context, user entities.User) error
	UpdateUserProfile(ctx context.Context, user entities.User) error
}
