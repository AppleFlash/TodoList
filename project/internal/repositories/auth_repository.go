package repositories

import (
	"TaskManager/project/pkg/models"
	"errors"
	"strconv"
	"sync"
)

type AuthRepository interface {
	Register(creds models.Credentionals) error
	Login(creds models.Credentionals) (models.StorageUser, error)
	IsUserExists(id string) bool
}

type authRepositoryImp struct {
	users        map[string]*models.StorageUser
	currentIndex int
	mu           sync.Mutex
}

func NewAuthRepository() AuthRepository {
	return &authRepositoryImp{
		users:        make(map[string]*models.StorageUser),
		currentIndex: 0,
	}
}

func (r *authRepositoryImp) Register(creds models.Credentionals) error {
	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
		r.currentIndex++
	}()

	if _, exists := r.users[creds.Login]; exists {
		return errors.New("user already exist")
	}

	user := models.StorageUser{
		User:     models.User{Id: strconv.Itoa(r.currentIndex), Login: creds.Login},
		Password: creds.Password,
	}
	r.users[creds.Login] = &user
	return nil
}

func (r *authRepositoryImp) Login(creds models.Credentionals) (models.StorageUser, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[creds.Login]
	if !exists {
		return models.StorageUser{}, errors.New("user doesn't exist")
	}

	return *user, nil
}

func (r *authRepositoryImp) IsUserExists(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.users[id]
	return exists
}
