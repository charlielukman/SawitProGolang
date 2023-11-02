package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SawitProRecruitment/UserService/entities"
	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/internal"
	"github.com/SawitProRecruitment/UserService/repository"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestServer_Profile(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name                 string
		mockRepo             func(*gomock.Controller) repository.RepositoryInterface
		mockJWT              func(*gomock.Controller) internal.JWTSigner
		mockPasswordComparer func(*gomock.Controller) internal.PasswordComparer
		contextUserID        int
		expectedCode         int
		expectedResponse     interface{}
	}{
		{
			name: "no user_id in context",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				return nil
			},
			mockJWT: func(ctrl *gomock.Controller) internal.JWTSigner {
				return nil
			},
			mockPasswordComparer: func(ctrl *gomock.Controller) internal.PasswordComparer {
				return nil
			},
			expectedCode: http.StatusForbidden,
			expectedResponse: generated.ErrorResponse{
				Message: "user not logged in",
			},
		},
		{
			name: "user not found",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				mockRepo := repository.NewMockRepositoryInterface(ctrl)
				mockRepo.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(entities.User{}, internal.ForbiddenError{
					Message: "user not registered",
				})
				return mockRepo
			},
			mockJWT: func(ctrl *gomock.Controller) internal.JWTSigner {
				return nil
			},
			mockPasswordComparer: func(ctrl *gomock.Controller) internal.PasswordComparer {
				return nil
			},
			contextUserID: 1,
			expectedCode:  http.StatusForbidden,
			expectedResponse: generated.ErrorResponse{
				Message: "user not registered",
			},
		},
		{
			name: "when get user got error then internal server error",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				mockRepo := repository.NewMockRepositoryInterface(ctrl)
				mockRepo.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(entities.User{}, errors.New("some error"))
				return mockRepo
			},
			mockJWT: func(ctrl *gomock.Controller) internal.JWTSigner {
				return nil
			},
			mockPasswordComparer: func(ctrl *gomock.Controller) internal.PasswordComparer {
				return nil
			},
			contextUserID: 1,
			expectedCode:  http.StatusInternalServerError,
			expectedResponse: generated.ErrorResponse{
				Message: "some error",
			},
		},
		{
			name: "get profile success",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				mockRepo := repository.NewMockRepositoryInterface(ctrl)
				mockRepo.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(entities.User{
					FullName:    "John Doe",
					PhoneNumber: "+628123456789",
				}, nil)
				return mockRepo
			},
			mockJWT: func(ctrl *gomock.Controller) internal.JWTSigner {
				return nil
			},
			mockPasswordComparer: func(ctrl *gomock.Controller) internal.PasswordComparer {
				return nil
			},
			contextUserID: 1,
			expectedCode:  http.StatusOK,
			expectedResponse: generated.UserResponse{
				Data: struct {
					FullName    string `json:"fullName"`
					PhoneNumber string `json:"phoneNumber"`
				}{
					FullName:    "John Doe",
					PhoneNumber: "+628123456789",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpReq := httptest.NewRequest(http.MethodGet, "/api/users", nil)
			httpResp := httptest.NewRecorder()
			ctx := e.NewContext(httpReq, httpResp)
			if tt.contextUserID != 0 {
				ctx.Set("user_id", tt.contextUserID)
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRepo := tt.mockRepo(ctrl)
			mockJWT := tt.mockJWT(ctrl)
			mockPasswordComparer := tt.mockPasswordComparer(ctrl)

			s := NewServer(NewServerOptions{
				Repository:       mockRepo,
				JWTClaim:         mockJWT,
				PasswordComparer: mockPasswordComparer,
			})
			s.Profile(ctx)

			assert.Equal(t, tt.expectedCode, ctx.Response().Status)

			respBody, _ := io.ReadAll(httpResp.Body)
			switch expected := tt.expectedResponse.(type) {
			case generated.UserResponse:
				var resp generated.UserResponse
				json.Unmarshal(respBody, &resp)
				assert.Equal(t, expected, resp)
			case generated.ErrorResponse:
				var resp generated.ErrorResponse
				json.Unmarshal(respBody, &resp)
				assert.Equal(t, expected, resp)
			}
		})
	}
}

func TestServer_UpdateProfile(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name                 string
		fullName             string
		phoneNumber          string
		mockRepo             func(*gomock.Controller) repository.RepositoryInterface
		mockJWT              func(*gomock.Controller) internal.JWTSigner
		mockPasswordComparer func(*gomock.Controller) internal.PasswordComparer
		contextUserID        int
		expectedCode         int
		expectedResponse     interface{}
	}{
		{
			name:        "no user_id in context",
			fullName:    "John Doe",
			phoneNumber: "+628123456789",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				return nil
			},
			mockJWT: func(ctrl *gomock.Controller) internal.JWTSigner {
				return nil
			},
			mockPasswordComparer: func(ctrl *gomock.Controller) internal.PasswordComparer {
				return nil
			},
			expectedCode: http.StatusForbidden,
		},
		{
			name:        "no fullName and phoneNumber updated then bad request",
			fullName:    "",
			phoneNumber: "",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				return nil
			},
			mockJWT: func(ctrl *gomock.Controller) internal.JWTSigner {
				return nil
			},
			mockPasswordComparer: func(ctrl *gomock.Controller) internal.PasswordComparer {
				return nil
			},
			contextUserID: 1,
			expectedCode:  http.StatusBadRequest,
			expectedResponse: generated.ErrorResponse{
				Message: "nothing to update",
			},
		},
		{
			name:     "fullName less than 3 characters then bad request",
			fullName: "Jo",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				return nil
			},
			mockJWT: func(ctrl *gomock.Controller) internal.JWTSigner {
				return nil
			},
			mockPasswordComparer: func(ctrl *gomock.Controller) internal.PasswordComparer {
				return nil
			},
			contextUserID: 1,
			expectedCode:  http.StatusBadRequest,
			expectedResponse: generated.ErrorResponse{
				Message: "full name must be between 3 and 60 characters",
			},
		},
		{
			name:        "phoneNumber not valid then bad request",
			phoneNumber: "123456789",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				return nil
			},
			mockJWT: func(ctrl *gomock.Controller) internal.JWTSigner {
				return nil
			},
			mockPasswordComparer: func(ctrl *gomock.Controller) internal.PasswordComparer {
				return nil
			},
			contextUserID: 1,
			expectedCode:  http.StatusBadRequest,
			expectedResponse: generated.ErrorResponse{
				Message: "phone number must be between 10 and 13 characters, phone number must start with +62",
			},
		},
		{
			name:        "phoneNumber already registered then conflict",
			phoneNumber: "+628123456789",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				mockRepo := repository.NewMockRepositoryInterface(ctrl)
				mockRepo.EXPECT().UpdateUserProfile(gomock.Any(), gomock.Any()).Return(
					internal.ConflictError{
						Message: "phone number already registered",
					},
				)
				return mockRepo
			},
			mockJWT: func(ctrl *gomock.Controller) internal.JWTSigner {
				return nil
			},
			mockPasswordComparer: func(ctrl *gomock.Controller) internal.PasswordComparer {
				return nil
			},
			contextUserID: 1,
			expectedCode:  http.StatusConflict,
			expectedResponse: generated.ErrorResponse{
				Message: "phone number already registered",
			},
		},
		{
			name:        "update profile database fail then internal server error",
			fullName:    "John Doe",
			phoneNumber: "+628123456789",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				mockRepo := repository.NewMockRepositoryInterface(ctrl)
				mockRepo.EXPECT().UpdateUserProfile(gomock.Any(), gomock.Any()).Return(errors.New("some error"))
				return mockRepo
			},
			mockJWT: func(ctrl *gomock.Controller) internal.JWTSigner {
				return nil
			},
			mockPasswordComparer: func(ctrl *gomock.Controller) internal.PasswordComparer {
				return nil
			},
			contextUserID: 1,
			expectedCode:  http.StatusInternalServerError,
			expectedResponse: generated.ErrorResponse{
				Message: "some error",
			},
		},
		{
			name:        "update profile success",
			fullName:    "John Doe",
			phoneNumber: "+628123456789",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				mockRepo := repository.NewMockRepositoryInterface(ctrl)
				mockRepo.EXPECT().UpdateUserProfile(gomock.Any(), gomock.Any()).Return(nil)
				return mockRepo
			},
			mockJWT: func(ctrl *gomock.Controller) internal.JWTSigner {
				return nil
			},
			mockPasswordComparer: func(ctrl *gomock.Controller) internal.PasswordComparer {
				return nil
			},
			contextUserID: 1,
			expectedCode:  http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param := generated.UpdateProfileJSONBody{
				FullName:    tt.fullName,
				PhoneNumber: tt.phoneNumber,
			}
			body, _ := json.Marshal(param)
			httpReq := httptest.NewRequest(http.MethodPut, "/api/users", bytes.NewReader(body))
			httpReq.Header.Set("Content-Type", "application/json")
			httpResp := httptest.NewRecorder()
			ctx := e.NewContext(httpReq, httpResp)
			if tt.contextUserID != 0 {
				ctx.Set("user_id", tt.contextUserID)
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRepo := tt.mockRepo(ctrl)
			mockJWT := tt.mockJWT(ctrl)
			mockPasswordComparer := tt.mockPasswordComparer(ctrl)

			s := NewServer(NewServerOptions{
				Repository:       mockRepo,
				JWTClaim:         mockJWT,
				PasswordComparer: mockPasswordComparer,
			})
			s.UpdateProfile(ctx)

			assert.Equal(t, tt.expectedCode, ctx.Response().Status)

			respBody, _ := io.ReadAll(httpResp.Body)
			switch expected := tt.expectedResponse.(type) {
			case generated.ErrorResponse:
				var resp generated.ErrorResponse
				json.Unmarshal(respBody, &resp)
				assert.Equal(t, expected, resp)
			}
		})
	}
}
