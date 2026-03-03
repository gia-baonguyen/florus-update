package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/florus/backend/internal/config"
	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
	"github.com/florus/backend/pkg/utils"
)

var (
	ErrInvalidGoogleToken = errors.New("invalid Google token")
	ErrGoogleAuthFailed   = errors.New("Google authentication failed")
)

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

type GoogleAuthService interface {
	LoginWithGoogle(idToken string) (*dto.AuthResponse, error)
}

type googleAuthService struct {
	userRepo     repository.UserRepository
	jwtConfig    utils.JWTConfig
	googleConfig config.GoogleConfig
}

func NewGoogleAuthService(userRepo repository.UserRepository, jwtConfig utils.JWTConfig, googleConfig config.GoogleConfig) GoogleAuthService {
	return &googleAuthService{
		userRepo:     userRepo,
		jwtConfig:    jwtConfig,
		googleConfig: googleConfig,
	}
}

func (s *googleAuthService) LoginWithGoogle(idToken string) (*dto.AuthResponse, error) {
	// Verify Google token and get user info
	googleUser, err := s.verifyGoogleToken(idToken)
	if err != nil {
		return nil, err
	}

	// Check if user exists
	user, err := s.userRepo.FindByEmail(googleUser.Email)
	if err != nil {
		// User doesn't exist, create new user
		user = &models.User{
			Email:        googleUser.Email,
			Name:         googleUser.Name,
			GoogleID:     googleUser.ID,
			AvatarURL:    googleUser.Picture,
			AuthProvider: models.AuthProviderGoogle,
			Role:         models.RoleUser,
			UserStatus:   models.StatusCold,
		}

		if err := s.userRepo.Create(user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// User exists, check if it's a Google user or link accounts
		if user.AuthProvider == models.AuthProviderLocal {
			// Link existing local account with Google
			user.GoogleID = googleUser.ID
			user.AuthProvider = models.AuthProviderGoogle
			if user.AvatarURL == "" {
				user.AvatarURL = googleUser.Picture
			}
			if err := s.userRepo.Update(user); err != nil {
				return nil, fmt.Errorf("failed to link Google account: %w", err)
			}
		} else if user.GoogleID != googleUser.ID {
			// Different Google account trying to use same email
			return nil, ErrInvalidGoogleToken
		}
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role), s.jwtConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AuthResponse{
		Token: token,
		User:  dto.ToUserResponse(user),
	}, nil
}

func (s *googleAuthService) verifyGoogleToken(idToken string) (*GoogleUserInfo, error) {
	// Use Google's tokeninfo endpoint to verify the token
	// For production, you should use Google's OAuth2 library
	ctx := context.Background()

	// Method 1: Verify using tokeninfo endpoint (simpler)
	tokenInfoURL := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken)
	req, err := http.NewRequestWithContext(ctx, "GET", tokenInfoURL, nil)
	if err != nil {
		return nil, ErrGoogleAuthFailed
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, ErrGoogleAuthFailed
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Try access token method instead
		return s.getUserInfoFromAccessToken(idToken)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrGoogleAuthFailed
	}

	var tokenInfo struct {
		Aud           string `json:"aud"`
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
	}

	if err := json.Unmarshal(body, &tokenInfo); err != nil {
		return nil, ErrInvalidGoogleToken
	}

	// Verify the token was issued for our app (optional - enable in production)
	// if tokenInfo.Aud != s.googleConfig.ClientID {
	// 	return nil, ErrInvalidGoogleToken
	// }

	return &GoogleUserInfo{
		ID:            tokenInfo.Sub,
		Email:         tokenInfo.Email,
		VerifiedEmail: tokenInfo.EmailVerified == "true",
		Name:          tokenInfo.Name,
		Picture:       tokenInfo.Picture,
		GivenName:     tokenInfo.GivenName,
		FamilyName:    tokenInfo.FamilyName,
	}, nil
}

func (s *googleAuthService) getUserInfoFromAccessToken(accessToken string) (*GoogleUserInfo, error) {
	// Use access token to get user info
	userInfoURL := "https://www.googleapis.com/oauth2/v2/userinfo"
	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return nil, ErrGoogleAuthFailed
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, ErrGoogleAuthFailed
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrInvalidGoogleToken
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrGoogleAuthFailed
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, ErrInvalidGoogleToken
	}

	return &userInfo, nil
}
