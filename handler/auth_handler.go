package handler

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/SawitProRecruitment/UserService/entities"
	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/internal"
	"github.com/labstack/echo/v4"
)

func (s *Server) Register(ctx echo.Context) error {
	var request generated.RegisterJSONRequestBody
	if err := ctx.Bind(&request); err != nil {
		return handleError(ctx, internal.BadRequestError{
			Message: err.Error(),
		})
	}

	if err := validateRegistrationRequest(request); err != nil {
		return handleError(ctx, err)
	}

	exists, err := s.Repository.IsExistUser(ctx.Request().Context(), entities.User{PhoneNumber: request.PhoneNumber})
	if err != nil {
		return handleError(ctx, err)
	}
	if exists {
		return handleError(ctx, internal.ConflictError{
			Message: "user already exists",
		})
	}

	password := request.Password
	hashedPassword, err := internal.HashPassword(password)
	if err != nil {
		return handleError(ctx, err)
	}

	user := entities.User{
		FullName:    request.FullName,
		PhoneNumber: request.PhoneNumber,
		Password:    hashedPassword,
	}
	id, err := s.Repository.CreateUser(ctx.Request().Context(), user)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, generated.UserRegistrationResponse{
		Data: struct {
			Id int `json:"id"`
		}{
			Id: id,
		},
	})
}

func validateRegistrationRequest(req generated.RegisterJSONRequestBody) error {
	var errs []string

	err := validatePhoneNumber(req.PhoneNumber)
	if err != nil {
		errs = append(errs, err.Error())
	}
	err = validateFullName(req.FullName)
	if err != nil {
		errs = append(errs, err.Error())
	}

	if len(req.Password) < entities.PasswordMinLength || len(req.Password) > entities.PasswordMaxLength {
		errs = append(errs, fmt.Sprintf("password must be between %d and %d characters", entities.PasswordMinLength, entities.PasswordMaxLength))
	}
	match, _ := regexp.MatchString(`[A-Z]`, req.Password)
	if !match {
		errs = append(errs, "password must contain at least one uppercase letter")
	}
	matchNumber, _ := regexp.MatchString(`[0-9]`, req.Password)
	if !matchNumber {
		errs = append(errs, "password must contain at least one number")
	}
	matchSpecial, _ := regexp.MatchString(`[^a-zA-Z0-9\s]`, req.Password)
	if !matchSpecial {
		errs = append(errs, "password must contain at least one special character")
	}

	if len(errs) > 0 {
		return internal.BadRequestError{
			Message: strings.Join(errs, ", "),
		}
	}

	return nil
}

func validatePhoneNumber(phoneNumber string) error {
	var errs []string

	if len(phoneNumber) < entities.PhoneNumberMinLength || len(phoneNumber) > entities.PhoneNumberMaxLength {
		errs = append(errs, fmt.Sprintf("phone number must be between %d and %d characters", entities.PhoneNumberMinLength, entities.PhoneNumberMaxLength))
	}
	if !strings.HasPrefix(phoneNumber, entities.PhoneNumberPrefix) {
		errs = append(errs, fmt.Sprintf("phone number must start with %s", entities.PhoneNumberPrefix))
	}

	if len(errs) > 0 {
		return internal.BadRequestError{
			Message: strings.Join(errs, ", "),
		}
	}

	return nil
}

func validateFullName(fullName string) error {
	if len(fullName) < entities.FullNameMinLength || len(fullName) > entities.FullNameMaxLength {
		return internal.BadRequestError{
			Message: fmt.Sprintf("full name must be between %d and %d characters", entities.FullNameMinLength, entities.FullNameMaxLength),
		}
	}

	return nil
}

func (s *Server) Login(ctx echo.Context) error {
	var request generated.LoginJSONRequestBody
	if err := ctx.Bind(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, internal.BadRequestError{
			Message: err.Error(),
		})
	}

	err := validateLoginRequest(request)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, internal.BadRequestError{
			Message: err.Error(),
		})
	}

	user, err := s.Repository.GetUserByPhoneNumber(ctx.Request().Context(), request.PhoneNumber)
	if err != nil {
		return handleError(ctx, err)
	}

	if err := s.PasswordComparer.ComparePassword(request.Password, user.Password); err != nil {
		return handleError(ctx, internal.UnauthorizedError{
			Message: "wrong password",
		})
	}

	err = s.Repository.UpdateUserLoginSuccess(ctx.Request().Context(), user)
	if err != nil {
		return handleError(ctx, err)
	}

	token, err := s.JWTClaim.SignJWT(user)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, generated.UserLoginResponse{
		Data: struct {
			Token  string `json:"token"`
			UserId int    `json:"user_id"`
		}{
			Token:  token,
			UserId: user.ID,
		},
	})
}

func validateLoginRequest(request generated.LoginJSONRequestBody) error {
	if request.PhoneNumber == "" {
		return fmt.Errorf("phone number must not be empty")
	}
	if request.Password == "" {
		return fmt.Errorf("password must not be empty")
	}

	return nil
}
