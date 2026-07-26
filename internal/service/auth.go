package service

import (
	"errors"
	"strings"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/auth"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo *repository.UserRepository
	authRepo *repository.AuthRepository
	jwtCfg   auth.JWTConfig
}

func NewAuthService(userRepo *repository.UserRepository, authRepo *repository.AuthRepository, jwtCfg auth.JWTConfig) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		authRepo: authRepo,
		jwtCfg:   jwtCfg,
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	MFACode  string `json:"mfa_code"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	ID                  string `json:"id"`
	Email               string `json:"email"`
	Name                string `json:"name"`
	Status              string `json:"status"`
	ForcePasswordChange bool   `json:"force_password_change"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=12"`
	Name     string `json:"name" binding:"required"`
}

func (s *AuthService) Login(req LoginRequest, ip, userAgent string) (*LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	// Check account status
	if user.Status == "deleted" {
		return nil, errors.New("account not found")
	}
	if user.Status == "locked" {
		return nil, errors.New("account is locked")
	}
	if user.Status == "deactivated" {
		return nil, errors.New("account is deactivated")
	}
	if user.Status == "pending" {
		return nil, errors.New("email not verified")
	}

	// Check password
	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		_ = s.userRepo.IncrementFailedAttempts(user.ID)
		count := s.authRepo.CountRecentFailedLogins(user.Email, 15)
		if count >= 5 {
			_ = s.userRepo.LockAccount(user.ID, "")
		}
		return nil, errors.New("invalid email or password")
	}

	// Check MFA
	if user.MFAEnabled {
		if req.MFACode == "" {
			return nil, errors.New("mfa code required")
		}
		// TODO: validate TOTP code against user.MFASecret
	}

	// Reset failed attempts
	_ = s.userRepo.ResetFailedAttempts(user.ID)
	_ = s.userRepo.UpdateLastLogin(user.ID, ip)

	// Get roles and permissions
	roles, _ := s.userRepo.GetUserRoles(user.ID)
	permissions, _ := s.userRepo.GetUserPermissions(user.ID)

	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name
	}
	permStrings := make([]string, len(permissions))
	for i, p := range permissions {
		permStrings[i] = p.Resource + "." + p.Action
	}

	companyID := ""
	cid, _ := s.userRepo.GetUserCompany(user.ID)
	if cid != nil {
		companyID = *cid
	}

	// Generate access token
	accessToken, err := auth.GenerateAccessToken(s.jwtCfg, user.ID, user.Email, companyID, roleNames, permStrings)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, tokenHash, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(s.jwtCfg.RefreshTokenTTL)
	refreshRecord := &models.RefreshToken{
		UserID:     user.ID,
		TokenHash:  tokenHash,
		DeviceInfo: userAgent,
		IPAddress:  ip,
		ExpiresAt:  expiresAt,
	}
	_ = s.authRepo.SaveRefreshToken(refreshRecord)

	// Log login
	_ = s.authRepo.SaveLoginHistory(&models.LoginHistory{
		UserID:    &user.ID,
		Email:     user.Email,
		Status:    "success",
		IPAddress: ip,
		UserAgent: userAgent,
		Browser:   extractBrowser(userAgent),
		OS:        extractOS(userAgent),
	})

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtCfg.AccessTokenTTL.Seconds()),
		User: UserResponse{
			ID:                  user.ID,
			Email:               user.Email,
			Name:                user.Name,
			Status:              user.Status,
			ForcePasswordChange: user.ForcePasswordChange,
		},
	}, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (*LoginResponse, error) {
	tokenHash := auth.HashToken(refreshToken)
	record, err := s.authRepo.FindRefreshToken(tokenHash)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	// Revoke the old token
	_ = s.authRepo.RevokeRefreshToken(tokenHash)

	// Get user
	user, err := s.userRepo.FindByID(record.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.Status != "active" {
		return nil, errors.New("account is not active")
	}

	// Get roles and permissions
	roles, _ := s.userRepo.GetUserRoles(user.ID)
	permissions, _ := s.userRepo.GetUserPermissions(user.ID)

	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name
	}
	permStrings := make([]string, len(permissions))
	for i, p := range permissions {
		permStrings[i] = p.Resource + "." + p.Action
	}

	companyID := ""
	cid, _ := s.userRepo.GetUserCompany(user.ID)
	if cid != nil {
		companyID = *cid
	}

	// Generate new access token
	accessToken, err := auth.GenerateAccessToken(s.jwtCfg, user.ID, user.Email, companyID, roleNames, permStrings)
	if err != nil {
		return nil, err
	}

	// Generate new refresh token
	newRefreshToken, newTokenHash, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(s.jwtCfg.RefreshTokenTTL)
	newRecord := &models.RefreshToken{
		UserID:     user.ID,
		TokenHash:  newTokenHash,
		DeviceInfo: record.DeviceInfo,
		IPAddress:  record.IPAddress,
		ExpiresAt:  expiresAt,
	}
	_ = s.authRepo.SaveRefreshToken(newRecord)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.jwtCfg.AccessTokenTTL.Seconds()),
		User: UserResponse{
			ID:                  user.ID,
			Email:               user.Email,
			Name:                user.Name,
			Status:              user.Status,
			ForcePasswordChange: user.ForcePasswordChange,
		},
	}, nil
}

func (s *AuthService) Logout(refreshToken string) error {
	tokenHash := auth.HashToken(refreshToken)
	return s.authRepo.RevokeRefreshToken(tokenHash)
}

func (s *AuthService) LogoutAll(userID string) error {
	return s.authRepo.RevokeAllUserTokens(userID)
}

func (s *AuthService) Register(req RegisterRequest) (*models.User, error) {
	existing, _ := s.userRepo.FindByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		Status:       "pending",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) ChangePassword(userID, currentPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if !auth.CheckPassword(currentPassword, user.PasswordHash) {
		return errors.New("current password is incorrect")
	}

	// Check password history
	history, _ := s.userRepo.GetPasswordHistory(userID, 5)
	newHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}
	for _, h := range history {
		if auth.CheckPassword(newPassword, h.PasswordHash) {
			return errors.New("password cannot match last 5 passwords")
		}
	}

	if err := s.userRepo.UpdatePassword(userID, newHash); err != nil {
		return errors.New("failed to update password")
	}
	if err := s.userRepo.AddPasswordHistory(userID, newHash); err != nil {
		return errors.New("failed to save password history")
	}
	_ = s.userRepo.SetForcePasswordChange(userID, false)
	if err := s.authRepo.RevokeAllUserTokens(userID); err != nil {
		return err
	}

	return nil
}

type ProfileResponse struct {
	ID                  string     `json:"id"`
	Email               string     `json:"email"`
	Name                string     `json:"name"`
	Status              string     `json:"status"`
	ForcePasswordChange bool       `json:"force_password_change"`
	Roles               []RoleInfo `json:"roles,omitempty"`
	CreatedAt           string     `json:"created_at"`
}

type RoleInfo struct {
	Name string `json:"name"`
}

func toProfileResponse(user *models.User, roles []models.Role) ProfileResponse {
	roleInfos := make([]RoleInfo, len(roles))
	for i, r := range roles {
		roleInfos[i] = RoleInfo{Name: r.Name}
	}
	return ProfileResponse{
		ID:                  user.ID,
		Email:               user.Email,
		Name:                user.Name,
		Status:              user.Status,
		ForcePasswordChange: user.ForcePasswordChange,
		Roles:               roleInfos,
		CreatedAt:           user.CreatedAt.Format(time.RFC3339),
	}
}

func (s *AuthService) GetUser(userID string) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *AuthService) GetProfile(userID string) (*ProfileResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	roles, _ := s.userRepo.GetUserRoles(user.ID)
	resp := toProfileResponse(user, roles)
	return &resp, nil
}

type UpdateProfileRequest struct {
	Name string `json:"name"`
}

func (s *AuthService) UpdateProfile(userID string, req UpdateProfileRequest) (*ProfileResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	roles, _ := s.userRepo.GetUserRoles(user.ID)
	resp := toProfileResponse(user, roles)
	return &resp, nil
}

func (s *AuthService) GetSessions(userID string) ([]models.Session, error) {
	return s.authRepo.GetActiveSessions(userID)
}

func extractBrowser(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg"):
		return "Chrome"
	case strings.Contains(ua, "firefox"):
		return "Firefox"
	case strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome"):
		return "Safari"
	case strings.Contains(ua, "edg"):
		return "Edge"
	default:
		return "Unknown"
	}
}

func extractOS(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "windows"):
		return "Windows"
	case strings.Contains(ua, "mac"):
		return "macOS"
	case strings.Contains(ua, "linux"):
		return "Linux"
	case strings.Contains(ua, "android"):
		return "Android"
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		return "iOS"
	default:
		return "Unknown"
	}
}
