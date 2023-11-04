package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/SawitProRecruitment/UserService/entities"
	"github.com/SawitProRecruitment/UserService/internal"
)

func (r *Repository) CreateUser(ctx context.Context, user entities.User) (userID int, err error) {
	err = r.Db.QueryRowContext(ctx,
		`INSERT INTO users (full_name, phone_number, password, created_at) 
		VALUES ($1, $2, $3, $4, NOW()) RETURNING id`,
		user.FullName, user.PhoneNumber, user.Password).
		Scan(&userID)
	if err != nil {
		return
	}
	return
}

func (r *Repository) IsExistUser(ctx context.Context, user entities.User) (bool, error) {
	var id int
	err := r.Db.QueryRowContext(ctx,
		`SELECT id FROM users WHERE phone_number = $1`,
		user.PhoneNumber).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}
	return true, nil
}

func (r *Repository) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (entities.User, error) {
	var user entities.User
	err := r.Db.QueryRowContext(ctx,
		`SELECT 
				id,
				full_name,
				phone_number,
				password,
			FROM users 
			WHERE phone_number = $1`,
		phoneNumber).Scan(&user.ID, &user.FullName, &user.PhoneNumber, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return entities.User{}, internal.BadRequestError{
				Message: "user not registered",
			}
		}
		return user, internal.InternalServerError{
			Message: fmt.Errorf("failed to get user by phone number: %w", err).Error(),
		}
	}
	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int) (entities.User, error) {
	var user entities.User
	err := r.Db.QueryRowContext(ctx,
		`SELECT 
				id,
				full_name,
				phone_number,
				password
			FROM users 
			WHERE id = $1`,
		id).Scan(&user.ID, &user.FullName, &user.PhoneNumber, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return entities.User{}, internal.ForbiddenError{
				Message: "user not registered",
			}
		}
		return user, internal.InternalServerError{
			Message: fmt.Errorf("failed to get user by id: %w", err).Error(),
		}
	}
	return user, nil
}

func (r *Repository) UpdateUserLoginSuccess(ctx context.Context, user entities.User) error {
	_, err := r.Db.ExecContext(ctx,
		`UPDATE users 
			SET last_login_at = NOW(), 
				successful_logins = successful_logins + 1 
			WHERE id = $1`,
		user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user login success: %w", err)
	}
	return nil
}

func (r *Repository) UpdateUserProfile(ctx context.Context, user entities.User) error {
	_, err := r.Db.ExecContext(ctx,
		`UPDATE users 
			SET full_name = CASE WHEN $1 != '' THEN $1 ELSE full_name END,
				phone_number = CASE WHEN $2 != '' THEN $2 ELSE phone_number END
			WHERE id = $3`,
		user.FullName,
		user.PhoneNumber,
		user.ID)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == "unique_violation" && strings.Contains(err.Detail, "phone_number") {
				return internal.ConflictError{
					Message: "phone number already registered",
				}
			}
		}
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	return nil
}
