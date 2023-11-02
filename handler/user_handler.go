package handler

import (
	"fmt"
	"net/http"

	"github.com/SawitProRecruitment/UserService/entities"
	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/internal"
	"github.com/labstack/echo/v4"
)

func (s *Server) Profile(ctx echo.Context) error {
	userID, ok := ctx.Get("user_id").(int)
	if !ok {
		return handleError(ctx, internal.ForbiddenError{
			Message: "user not logged in",
		})
	}

	user, err := s.Repository.GetUserByID(ctx.Request().Context(), userID)
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, generated.UserResponse{
		Data: struct {
			FullName    string `json:"fullName"`
			PhoneNumber string `json:"phoneNumber"`
		}{
			FullName:    user.FullName,
			PhoneNumber: user.PhoneNumber,
		},
	})
}

func (s *Server) UpdateProfile(ctx echo.Context) error {
	var request generated.UpdateProfileJSONRequestBody

	if err := ctx.Bind(&request); err != nil {
		return handleError(ctx, internal.BadRequestError{
			Message: err.Error(),
		})
	}

	userID, ok := ctx.Get("user_id").(int)
	if !ok {
		return handleError(ctx, internal.ForbiddenError{
			Message: "user not logged in",
		})
	}

	err := validateUpdateProfileRequest(request)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{
			Message: err.Error(),
		})
	}

	err = s.Repository.UpdateUserProfile(ctx.Request().Context(), entities.User{
		FullName:    request.FullName,
		PhoneNumber: request.PhoneNumber,
		ID:          userID,
	})
	if err != nil {
		return handleError(ctx, err)
	}

	return ctx.NoContent(http.StatusOK)
}

func validateUpdateProfileRequest(request generated.UpdateProfileJSONRequestBody) error {
	if request.PhoneNumber == "" && request.FullName == "" {
		return fmt.Errorf("nothing to update")
	}
	if request.PhoneNumber != "" {
		return validatePhoneNumber(request.PhoneNumber)
	}
	if request.FullName != "" {
		return validateFullName(request.FullName)
	}

	return nil
}
