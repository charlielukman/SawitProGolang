package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/SawitProRecruitment/UserService/internal"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func BearerAuthMiddleware(jwtSigner internal.JWTSigner, publicKeyPath string, next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusForbidden, "missing Authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return echo.NewHTTPError(http.StatusForbidden, "invalid Authorization header format")
		}

		token := parts[1]

		publicKeyBytes, err := os.ReadFile(publicKeyPath)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "could not read public key")
		}

		publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "could not parse public key")
		}

		claims, err := jwtSigner.VerifyJWT(token, publicKey)
		if err != nil {
			return echo.NewHTTPError(http.StatusForbidden, "invalid token")
		}

		c.Set("user_id", claims.UserID)

		return next(c)
	}
}
