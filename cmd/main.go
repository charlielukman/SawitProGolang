package main

import (
	"os"
	"strings"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/handler"
	"github.com/SawitProRecruitment/UserService/internal"
	"github.com/SawitProRecruitment/UserService/middleware"
	"github.com/SawitProRecruitment/UserService/repository"

	echoMiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	server := newServer()
	var serverInterface generated.ServerInterface = server

	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.HasPrefix(c.Request().URL.Path, "/api/users") {
				return middleware.BearerAuthMiddleware(server.JWTClaim, "public.pem", next)(c)
			}
			return next(c)
		}
	})

	api := e.Group("/api")
	generated.RegisterHandlers(api, serverInterface)

	e.Logger.Fatal(e.Start(":1323"))
}

func newServer() *handler.Server {
	dbDsn := os.Getenv("DATABASE_URL")
	var repo repository.RepositoryInterface = repository.NewRepository(repository.NewRepositoryOptions{
		Dsn: dbDsn,
	})
	jwt, err := internal.NewJWT("private.pem")
	if err != nil {
		panic(err)
	}

	opts := handler.NewServerOptions{
		Repository:       repo,
		JWTClaim:         jwt,
		PasswordComparer: internal.PasswordComparerImpl{},
	}
	return handler.NewServer(opts)
}
