package services

import (
	"TaskManager/project/internal/repositories"
	"TaskManager/project/pkg/models"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessData struct {
	AccessToken  string
	RefreshToken string
}

type LoginData struct {
	User   models.StorageUser
	Access AccessData
}

type AuthService interface {
	Register(ctx context.Context, creds models.Credentionals) error
	Login(ctx context.Context, creds models.Credentionals) (LoginData, error)
	Validate(ctx context.Context, tokenString string) (*Claims, error)
	Refresh(ctx context.Context, refreshToken string) (AccessData, error)
}

type authServiceImp struct {
	repo repositories.AuthRepository
}

func NewAuthService(repo repositories.AuthRepository) AuthService {
	return &authServiceImp{
		repo: repo,
	}
}

func (s *authServiceImp) Register(ctx context.Context, creds models.Credentionals) error {
	if error := s.repo.Register(creds); error != nil {
		return error
	}

	return nil
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var secretKey = []byte("app_secret_key")

func (s *authServiceImp) Login(ctx context.Context, creds models.Credentionals) (LoginData, error) {
	user, error := s.repo.Login(creds)
	if error != nil {
		return LoginData{}, error
	}

	if user.Password != creds.Password {
		return LoginData{}, errors.New("unauthorized")
	}

	access, error := s.generateJWT(creds.Login, 5*time.Second)
	if error != nil {
		return LoginData{}, errors.New("access token error")
	}

	refresh, error := s.generateJWT(creds.Login, 720*time.Hour)
	if error != nil {
		return LoginData{}, errors.New("access token error")
	}

	return LoginData{
		User: user,
		Access: AccessData{
			AccessToken:  access,
			RefreshToken: refresh,
		},
	}, nil
}

func (s *authServiceImp) generateJWT(login string, duration time.Duration) (string, error) {
	accessTime := jwt.NewNumericDate(time.Now().Add(duration))

	claims := Claims{
		Username: login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: accessTime,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, error := token.SignedString(secretKey)
	if error != nil {
		return "", error
	}

	return signed, nil
}

func (s *authServiceImp) Validate(ctx context.Context, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("token invalid")
	}
}

func (s *authServiceImp) Refresh(ctx context.Context, refreshToken string) (AccessData, error) {
	claims, error := s.Validate(ctx, refreshToken)
	if error != nil {
		return AccessData{}, errors.New("token in invalid")
	}

	issuer := claims.Username
	if !s.repo.IsUserExists(issuer) {
		return AccessData{}, errors.New("token in invalid")
	}

	access, error := s.generateJWT(issuer, 3*time.Minute)
	if error != nil {
		return AccessData{}, errors.New("access token error")
	}

	refresh, error := s.generateJWT(issuer, 720*time.Hour)
	if error != nil {
		return AccessData{}, errors.New("access token error")
	}

	return AccessData{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}
