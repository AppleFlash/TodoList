package services

import (
	"TaskManager/project/internal/repositories"
	"TaskManager/project/pkg/models"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// MARK: Models

type AccessData struct {
	AccessToken  string
	RefreshToken string
}

type LoginData struct {
	User   models.StorageUser
	Access AccessData
}

// MARK: Service

type Claims struct {
	Username  string `json:"username"`
	SessionID string `json:"session_id"`

	jwt.RegisteredClaims
}

type AuthService interface {
	Register(ctx context.Context, creds models.Credentionals) error
	Login(ctx context.Context, creds models.Credentionals) (LoginData, error)
	Validate(ctx context.Context, accessData AccessData) (AccessData, error)
}

// MARK: Package vers

var (
	accessDuration  = 15 * time.Second
	refreshDuration = 24 * 7 * time.Hour
	secretKey       = []byte("secret_key")
	sessions        = make(map[string]string)
)

// MARK: Implementation

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

func (s *authServiceImp) Login(ctx context.Context, creds models.Credentionals) (LoginData, error) {
	user, error := s.repo.Login(creds)
	if error != nil {
		return LoginData{}, error
	}

	if user.Password != creds.Password {
		return LoginData{}, errors.New("unauthorized")
	}

	// Создаём уникальный иденитфикатор сессии, который используется и access и refresh токенах.
	// Новый session_id создаётся при логине и при пересоздании refresh токена
	id := uuid.NewString()
	accessData, error := s.generateNewAccess(creds.Login, id)
	if error != nil {
		return LoginData{}, error
	}

	sessions[id] = accessData.RefreshToken

	return LoginData{
		User:   user,
		Access: accessData,
	}, nil
}

func (s *authServiceImp) Validate(ctx context.Context, accessData AccessData) (AccessData, error) {
	accessToken, err := parseToken(accessData.AccessToken)
	accessClaims := accessToken.Claims.(*Claims)
	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		return AccessData{}, err
	}

	storedRefresh := sessions[accessClaims.SessionID]
	if storedRefresh != accessData.RefreshToken {
		return AccessData{}, errors.New("unauthorized")
	}

	// Access token живой
	if accessToken.Valid {
		return accessData, nil
	} else { // Access token протух
		// Проверяем актуальность refresh token
		refreshToken, err := parseToken(accessData.RefreshToken)
		if err != nil {
			// Refresh token протух - пользователь не акторизован
			return AccessData{}, err
		}
		if refreshToken.Valid {
			newAccessData, error := s.generateNewAccess(accessClaims.Username, accessClaims.SessionID)
			if error != nil {
				return AccessData{}, error
			}
			sessions[accessClaims.SessionID] = newAccessData.RefreshToken
			return newAccessData, nil
		} else {
			return AccessData{}, errors.New("unauthorized")
		}
	}
}

// MARK: Private

func (s *authServiceImp) generateJWT(login string, id string, duration time.Duration) (string, error) {
	accessTime := jwt.NewNumericDate(time.Now().Add(duration))

	claims := Claims{
		Username:  login,
		SessionID: id,
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

func parseToken(tokenString string) (*jwt.Token, error) {
	// Функция для проверки подписи токена
	return jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что метод подписи совпадает
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
}

func (s *authServiceImp) generateNewAccess(login string, sessionID string) (AccessData, error) {
	access, error := s.generateJWT(login, sessionID, accessDuration)
	if error != nil {
		return AccessData{}, errors.New("access token error")
	}

	refresh, error := s.generateJWT(login, sessionID, refreshDuration)
	if error != nil {
		return AccessData{}, errors.New("access token error")
	}

	return AccessData{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}
