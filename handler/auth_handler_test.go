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

func TestServer_Register(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name             string
		phoneNumber      string
		password         string
		fullName         string
		mockRepo         func(*gomock.Controller) repository.RepositoryInterface
		expectedCode     int
		expectedResponse interface{}
	}{
		{
			name:        "When Register full name, phone number, password not provided then return bad request",
			phoneNumber: "",
			password:    "",
			fullName:    "",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				return repository.NewMockRepositoryInterface(ctrl)
			},
			expectedCode: http.StatusBadRequest,
			expectedResponse: generated.ErrorResponse{
				Message: "phone number must be between 10 and 13 characters, phone number must start with +62, full name must be between 3 and 60 characters, password must be between 6 and 64 characters, password must contain at least one uppercase letter, password must contain at least one number, password must contain at least one special character",
			},
		},
		{
			name:        "When Register full name, phone number, password VALID then return success",
			phoneNumber: "+628123456789",
			password:    "Password123!",
			fullName:    "John Doe",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				mockRepo := repository.NewMockRepositoryInterface(ctrl)
				mockRepo.EXPECT().IsExistUser(gomock.Any(), gomock.Any()).Return(false, nil)
				mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(1, nil)
				return mockRepo
			},
			expectedCode: http.StatusCreated,
			expectedResponse: generated.UserRegistrationResponse{
				Data: struct {
					Id int `json:"id"`
				}{
					Id: 1,
				},
			},
		},
		{
			name:        "When Register user already exist then return conflict",
			phoneNumber: "+628123456789",
			password:    "Password123!",
			fullName:    "John Doe",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				mockRepo := repository.NewMockRepositoryInterface(ctrl)
				mockRepo.EXPECT().IsExistUser(gomock.Any(), gomock.Any()).Return(true, nil)
				return mockRepo
			},
			expectedCode: http.StatusConflict,
			expectedResponse: generated.ErrorResponse{
				Message: "user already exists",
			},
		},
		{
			name:        "When Register user got error database call is exist user, return internal server error",
			phoneNumber: "+628123456789",
			password:    "Password123!",
			fullName:    "John Doe",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				mockRepo := repository.NewMockRepositoryInterface(ctrl)
				mockRepo.EXPECT().IsExistUser(gomock.Any(), gomock.Any()).Return(false, errors.New("error db call exist user"))
				return mockRepo
			},
			expectedCode: http.StatusInternalServerError,
			expectedResponse: generated.ErrorResponse{
				Message: "error db call exist user",
			},
		},
		{
			name:        "When Register user got error database call create user, return internal server error",
			phoneNumber: "+628123456789",
			password:    "Password123!",
			fullName:    "John Doe",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				mockRepo := repository.NewMockRepositoryInterface(ctrl)
				mockRepo.EXPECT().IsExistUser(gomock.Any(), gomock.Any()).Return(false, nil)
				mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(1, errors.New("error db call create user"))
				return mockRepo
			},
			expectedCode: http.StatusInternalServerError,
			expectedResponse: generated.ErrorResponse{
				Message: "error db call create user",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param := generated.RegisterJSONRequestBody{
				PhoneNumber: tt.phoneNumber,
				Password:    tt.password,
				FullName:    tt.fullName,
			}
			body, _ := json.Marshal(param)
			httpReq := httptest.NewRequest(http.MethodPost, "/api/auth/registration", bytes.NewBuffer(body))
			httpReq.Header.Set("Content-Type", "application/json")
			httpResp := httptest.NewRecorder()
			ctx := e.NewContext(httpReq, httpResp)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRepo := tt.mockRepo(ctrl)

			s := NewServer(NewServerOptions{
				Repository: mockRepo,
			})
			s.Register(ctx)

			assert.Equal(t, tt.expectedCode, ctx.Response().Status)

			respBody, _ := io.ReadAll(httpResp.Body)
			switch expected := tt.expectedResponse.(type) {
			case generated.UserRegistrationResponse:
				var resp generated.UserRegistrationResponse
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

func Test_validateLoginRequest(t *testing.T) {
	type args struct {
		request generated.LoginJSONRequestBody
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "When Login Request phone number not provided then return error",
			args:        args{request: generated.LoginJSONRequestBody{}},
			wantErr:     true,
			expectedErr: "phone number must not be empty",
		},
		{
			name:        "When Login Request password not provided then return error",
			args:        args{request: generated.LoginJSONRequestBody{PhoneNumber: "+628123456789", Password: ""}},
			wantErr:     true,
			expectedErr: "password must not be empty",
		},
		{
			name:    "When Login Request phone number and password provided then return no error",
			args:    args{request: generated.LoginJSONRequestBody{PhoneNumber: "+628123456789", Password: "Password123!"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLoginRequest(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLoginRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err.Error() != tt.expectedErr {
				t.Errorf("validateLoginRequest() error = %v, expectedErr %v", err, tt.expectedErr)
			}
		})
	}
}

func TestServer_Login(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name                 string
		phoneNumber          string
		password             string
		mockRepo             func(*gomock.Controller) repository.RepositoryInterface
		mockJWT              func(*gomock.Controller) internal.JWTSigner
		mockPasswordComparer func(*gomock.Controller) internal.PasswordComparer
		expectedCode         int
		expectedResponse     interface{}
	}{
		{
			name:        "When Login phone number, password not provided then return bad request",
			phoneNumber: "",
			password:    "",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				return repository.NewMockRepositoryInterface(ctrl)
			},
			mockJWT: func(ctrl *gomock.Controller) internal.JWTSigner {
				return internal.NewMockJWTSigner(ctrl)
			},
			mockPasswordComparer: func(ctrl *gomock.Controller) internal.PasswordComparer {
				return internal.NewMockPasswordComparer(ctrl)
			},
			expectedCode: http.StatusBadRequest,
			expectedResponse: generated.ErrorResponse{
				Message: "phone number must not be empty",
			},
		},
		{
			name:        "When Login phone number, password VALID then return success",
			phoneNumber: "+628123456789",
			password:    "Password123!",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				mockRepo := repository.NewMockRepositoryInterface(ctrl)
				mockRepo.EXPECT().GetUserByPhoneNumber(gomock.Any(), gomock.Any()).Return(entities.User{
					ID:          1,
					FullName:    "John Doe",
					PhoneNumber: "+628123456789",
					Password:    "Password123!",
				}, nil)
				mockRepo.EXPECT().UpdateUserLoginSuccess(gomock.Any(), gomock.Any()).Return(nil)
				return mockRepo
			},
			mockJWT: func(ctrl *gomock.Controller) internal.JWTSigner {
				mockJWT := internal.NewMockJWTSigner(ctrl)
				mockJWT.EXPECT().SignJWT(gomock.Any()).Return("token", nil)
				return mockJWT
			},
			mockPasswordComparer: func(ctrl *gomock.Controller) internal.PasswordComparer {
				mockPasswordComparer := internal.NewMockPasswordComparer(ctrl)
				mockPasswordComparer.EXPECT().ComparePassword(gomock.Any(), gomock.Any()).Return(nil)
				return mockPasswordComparer
			},
			expectedCode: http.StatusOK,
			expectedResponse: generated.UserLoginResponse{
				Data: struct {
					Token  string `json:"token"`
					UserId int    `json:"user_id"`
				}{
					Token:  "token",
					UserId: 1,
				},
			},
		},
		{
			name:        "When Login user not registered then return bad request",
			phoneNumber: "+628123456789",
			password:    "Password123!",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				mockRepo := repository.NewMockRepositoryInterface(ctrl)
				mockRepo.EXPECT().GetUserByPhoneNumber(gomock.Any(), gomock.Any()).Return(entities.User{}, internal.BadRequestError{
					Message: "user not registered",
				})
				return mockRepo
			},
			mockJWT: func(ctrl *gomock.Controller) internal.JWTSigner {
				return internal.NewMockJWTSigner(ctrl)
			},
			mockPasswordComparer: func(ctrl *gomock.Controller) internal.PasswordComparer {
				return internal.NewMockPasswordComparer(ctrl)
			},
			expectedCode: http.StatusBadRequest,
			expectedResponse: generated.ErrorResponse{
				Message: "user not registered",
			},
		},
		{
			name:        "When Login user got error database call get user by phone number, return internal server error",
			phoneNumber: "+628123456789",
			password:    "Password123!",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				mockRepo := repository.NewMockRepositoryInterface(ctrl)
				mockRepo.EXPECT().GetUserByPhoneNumber(gomock.Any(), gomock.Any()).Return(entities.User{}, errors.New("error db call get user by phone number"))
				return mockRepo
			},
			mockJWT: func(ctrl *gomock.Controller) internal.JWTSigner {
				return internal.NewMockJWTSigner(ctrl)
			},
			mockPasswordComparer: func(ctrl *gomock.Controller) internal.PasswordComparer {
				return internal.NewMockPasswordComparer(ctrl)
			},
			expectedCode: http.StatusInternalServerError,
			expectedResponse: generated.ErrorResponse{
				Message: "error db call get user by phone number",
			},
		},
		{
			name:        "When Login user password wrong then return unauthorized",
			phoneNumber: "+628123456789",
			password:    "Password123!",
			mockRepo: func(ctrl *gomock.Controller) repository.RepositoryInterface {
				mockRepo := repository.NewMockRepositoryInterface(ctrl)
				mockRepo.EXPECT().GetUserByPhoneNumber(gomock.Any(), gomock.Any()).Return(entities.User{
					ID:          1,
					FullName:    "John Doe",
					PhoneNumber: "+628123456789",
					Password:    "Password123!",
				}, nil)
				return mockRepo
			},
			mockJWT: func(ctrl *gomock.Controller) internal.JWTSigner {
				return internal.NewMockJWTSigner(ctrl)
			},
			mockPasswordComparer: func(ctrl *gomock.Controller) internal.PasswordComparer {
				mockPasswordComparer := internal.NewMockPasswordComparer(ctrl)
				mockPasswordComparer.EXPECT().ComparePassword(gomock.Any(), gomock.Any()).Return(errors.New("wrong password"))
				return mockPasswordComparer
			},
			expectedCode: http.StatusUnauthorized,
			expectedResponse: generated.ErrorResponse{
				Message: "wrong password",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param := generated.LoginJSONRequestBody{
				PhoneNumber: tt.phoneNumber,
				Password:    tt.password,
			}
			body, _ := json.Marshal(param)
			httpReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
			httpReq.Header.Set("Content-Type", "application/json")
			httpResp := httptest.NewRecorder()
			ctx := e.NewContext(httpReq, httpResp)

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
			s.Login(ctx)

			assert.Equal(t, tt.expectedCode, ctx.Response().Status)

			respBody, _ := io.ReadAll(httpResp.Body)
			switch expected := tt.expectedResponse.(type) {
			case generated.UserLoginResponse:
				var resp generated.UserLoginResponse
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
