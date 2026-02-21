package service

import (
	"errors"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
	"github.com/florus/backend/pkg/utils"
	"gorm.io/gorm"
)

var (
	ErrUserExists         = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrWrongPassword      = errors.New("current password is incorrect")
)

type AuthService interface {
	Register(req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(req dto.LoginRequest) (*dto.AuthResponse, error)
	GetCurrentUser(userID uint) (*dto.UserResponse, error)
	UpdateProfile(userID uint, req dto.UpdateProfileRequest) (*dto.UserResponse, error)
	ChangePassword(userID uint, req dto.ChangePasswordRequest) error
}

type authService struct {
	userRepo  repository.UserRepository
	jwtConfig utils.JWTConfig
}

func NewAuthService(userRepo repository.UserRepository, jwtConfig utils.JWTConfig) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
	}
}

func (s *authService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Check if user exists
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, ErrUserExists
	}

	// Create new user
	user := &models.User{
		Email:   req.Email,
		Name:    req.Name,
		Phone:   req.Phone,
		Address: req.Address,
		Role:    models.RoleUser,
	}

	if err := user.SetPassword(req.Password); err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role), s.jwtConfig)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User:  dto.ToUserResponse(user),
	}, nil
}

func (s *authService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.CheckPassword(req.Password) {
		return nil, ErrInvalidCredentials
	}

	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role), s.jwtConfig)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User:  dto.ToUserResponse(user),
	}, nil
}

func (s *authService) GetCurrentUser(userID uint) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	resp := dto.ToUserResponse(user)
	return &resp, nil
}

func (s *authService) UpdateProfile(userID uint, req dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Update fields if provided
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Address != "" {
		user.Address = req.Address
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	resp := dto.ToUserResponse(user)
	return &resp, nil
}

func (s *authService) ChangePassword(userID uint, req dto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Verify current password
	if !user.CheckPassword(req.CurrentPassword) {
		return ErrWrongPassword
	}

	// Set new password
	if err := user.SetPassword(req.NewPassword); err != nil {
		return err
	}

	return s.userRepo.Update(user)
}
