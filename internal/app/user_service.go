package app

import (
	"errors"
	"fmt"

	"github.com/varmiguemunoz/command_pm_app/internal/domain"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

// constructor
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (userService *UserService) FindOrCreateByOAuth(
	provider string,
	providerID string,
	email string,
	name string,
	avatarURL *string,
) (*domain.User, bool, error) {

	var user domain.User

	result := userService.db.Where("provider = ? AND provider_id = ?", provider, providerID).First(&user)

	if result.Error == nil {
		return &user, false, nil
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, false, fmt.Errorf("database error: %w", result.Error)
	}

	user = domain.User{
		Name:       name,
		Email:      email,
		Provider:   provider,
		ProviderID: providerID,
		AvatarURL:  avatarURL,
	}

	if err := userService.db.Create(&user).Error; err != nil {
		return nil, false, fmt.Errorf("could not create user: %w", err)
	}

	return &user, true, nil
}

func (userService *UserService) GetByID(id uint) (*domain.User, error) {
	var user domain.User

	result := userService.db.First(&user, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user with id %d not found", id)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("error fetching user: %w", result.Error)
	}

	return &user, nil
}

func (userService *UserService) GetByEmail(email string) (*domain.User, error) {
	var user domain.User

	result := userService.db.Where("email = ?", email).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user with email %s not found", email)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("error fetching user: %w", result.Error)
	}

	return &user, nil
}

func (userService *UserService) GetAll() ([]domain.User, error) {
	var users []domain.User

	result := userService.db.Find(&users)

	if result.Error != nil {
		return nil, fmt.Errorf("error fetching users: %w", result.Error)
	}

	return users, nil
}
