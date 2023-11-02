package handler

import (
	"net/http"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/internal"
	"github.com/SawitProRecruitment/UserService/repository"
	"github.com/labstack/echo/v4"
)

type Server struct {
	Repository       repository.RepositoryInterface
	JWTClaim         internal.JWTSigner
	PasswordComparer internal.PasswordComparer
}

type NewServerOptions struct {
	Repository       repository.RepositoryInterface
	JWTClaim         internal.JWTSigner
	PasswordComparer internal.PasswordComparer
}

func NewServer(opts NewServerOptions) *Server {
	return &Server{
		Repository:       opts.Repository,
		JWTClaim:         opts.JWTClaim,
		PasswordComparer: opts.PasswordComparer,
	}
}

type httpStatusCodeProvider interface {
	HTTPStatusCode() int
}

func errorCode(err error) int {
	code := http.StatusInternalServerError
	pr, ok := err.(httpStatusCodeProvider)
	if ok {
		code = pr.HTTPStatusCode()
	}
	return code
}

func handleError(c echo.Context, err error) error {
	return c.JSON(errorCode(err), generated.ErrorResponse{
		Message: err.Error(),
	})
}
