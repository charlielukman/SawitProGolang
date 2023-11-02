package internal

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/SawitProRecruitment/UserService/entities"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type JWTSigner interface {
	SignJWT(user entities.User) (string, error)
	VerifyJWT(tokenString string, publicKey *rsa.PublicKey) (JWTClaim, error)
}

type JWTClaim struct {
	UserID     int
	PrivateKey *rsa.PrivateKey
	jwt.StandardClaims
}

type PasswordComparer interface {
	ComparePassword(password string, hashedPassword string, salt string) error
}

type PasswordComparerImpl struct{}

func NewJWT(privateKeyPath string) (*JWTClaim, error) {
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	return &JWTClaim{
		PrivateKey: privateKey,
	}, nil
}

func (j *JWTClaim) SignJWT(user entities.User) (string, error) {
	claims := &JWTClaim{
		UserID: user.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(j.PrivateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (j *JWTClaim) VerifyJWT(tokenString string, publicKey *rsa.PublicKey) (JWTClaim, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaim{}, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})
	if err != nil {
		return JWTClaim{}, err
	}

	if claims, ok := token.Claims.(*JWTClaim); ok && token.Valid {
		return *claims, nil
	} else {
		return JWTClaim{}, err
	}
}

func GetBearerToken(ctx echo.Context) ([]byte, error) {
	authorizationHeader := ctx.Request().Header.Get("Authorization")
	splitAuthorizationHeader := strings.Split(authorizationHeader, "Bearer")

	if len(splitAuthorizationHeader) != 2 {
		return nil, errors.New("invalid authorization bearer header")
	}

	token := strings.TrimSpace(splitAuthorizationHeader[1])

	return []byte(token), nil
}

func GenerateSalt() (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return hex.EncodeToString(salt), nil
}

func HashPassword(password string, salt string) (string, error) {
	combined := []byte(password + salt)
	hashedPassword, err := bcrypt.GenerateFromPassword(combined, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func (p PasswordComparerImpl) ComparePassword(password string, hashedPassword string, salt string) error {
	combined := []byte(password + salt)
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), combined)
}
