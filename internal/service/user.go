package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/auth"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo *repository.UserRepository
	authRepo *repository.AuthRepository
	roleRepo *repository.RoleRepository
}

func NewUserService(userRepo *repository.UserRepository, authRepo *repository.AuthRepository, roleRepo *repository.RoleRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		authRepo: authRepo,
		roleRepo: roleRepo,
	}
}

type CreateUserRequest struct {
	Email   string   `json:"email" binding:"required,email"`
	Name    string   `json:"name" binding:"required"`
	RoleIDs []string `json:"role_ids"`
}

type UpdateUserRequest struct {
	Name   string `json:"name"`
	Status string `json:"status" binding:"omitempty,oneof=active deactivated locked pending"`
}

type AssignRolesRequest struct {
	RoleIDs []string `json:"role_ids" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=12"`
}

type UserListItem struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

const securePasswordChars = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789!@#$%"

func generateSecurePassword(length int) (string, error) {
	result := make([]byte, length)
	max := big.NewInt(int64(len(securePasswordChars)))
	for i := range result {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		result[i] = securePasswordChars[n.Int64()]
	}
	return string(result), nil
}

func (s *UserService) ListUsers(page, limit int, status, search string) ([]UserListItem, int64, error) {
	users, total, err := s.userRepo.ListFiltered(page, limit, status, search)
	if err != nil {
		return nil, 0, err
	}

	items := make([]UserListItem, len(users))
	for i, u := range users {
		items[i] = UserListItem{
			ID:        u.ID,
			Email:     u.Email,
			Name:      u.Name,
			Status:    u.Status,
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
		}
	}
	return items, total, nil
}

func (s *UserService) GetUserByID(id string) (*models.User, []models.Role, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("user not found")
		}
		return nil, nil, err
	}
	roles, _ := s.userRepo.GetUserRoles(id)
	return user, roles, nil
}

func (s *UserService) CreateUser(req CreateUserRequest, createdBy string) (*models.User, string, error) {
	existing, _ := s.userRepo.FindByEmail(req.Email)
	if existing != nil {
		return nil, "", errors.New("email already registered")
	}

	password, err := generateSecurePassword(12)
	if err != nil {
		return nil, "", err
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	user := &models.User{
		Email:               req.Email,
		PasswordHash:        hash,
		Name:                req.Name,
		Status:              "active",
		ForcePasswordChange: true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", err
	}

	if len(req.RoleIDs) > 0 {
		if err := s.userRepo.ReplaceUserRoles(user.ID, req.RoleIDs, &createdBy); err != nil {
			return nil, "", err
		}
	}

	return user, password, nil
}

func (s *UserService) UpdateUser(id string, req UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) DeleteUser(id string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	roles, _ := s.userRepo.GetUserRoles(id)
	hasSuperAdmin := false
	for _, r := range roles {
		if r.Name == "super_admin" {
			hasSuperAdmin = true
			break
		}
	}
	if hasSuperAdmin {
		count, _ := s.userRepo.CountByRole("super_admin")
		if count <= 1 {
			return errors.New("cannot delete the last super admin")
		}
	}

	user.Status = "deleted"
	return s.userRepo.Update(user)
}

func (s *UserService) AssignRoles(id string, req AssignRolesRequest, currentUserRoles []string) error {
	if _, err := s.userRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	for _, roleID := range req.RoleIDs {
		role, err := s.roleRepo.FindByID(roleID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("role not found: " + roleID)
			}
			return err
		}
		if role.Name == "super_admin" && !contains(currentUserRoles, "super_admin") {
			return errors.New("only super admin can assign the super_admin role")
		}
	}

	currentRoles, _ := s.userRepo.GetUserRoles(id)
	currentHasSuper := false
	for _, r := range currentRoles {
		if r.Name == "super_admin" {
			currentHasSuper = true
			break
		}
	}
	if currentHasSuper {
		stillHasSuper := false
		for _, roleID := range req.RoleIDs {
			role, err := s.roleRepo.FindByID(roleID)
			if err == nil && role.Name == "super_admin" {
				stillHasSuper = true
				break
			}
		}
		if !stillHasSuper {
			count, _ := s.userRepo.CountByRole("super_admin")
			if count <= 1 {
				return errors.New("cannot remove super_admin from the last super admin")
			}
		}
	}

	return s.userRepo.ReplaceUserRoles(id, req.RoleIDs, nil)
}

func (s *UserService) GetUserRoles(id string) ([]models.Role, error) {
	if _, err := s.userRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return s.userRepo.GetUserRoles(id)
}

func (s *UserService) AdminResetPassword(id string) (string, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("user not found")
		}
		return "", err
	}

	password, err := generateSecurePassword(12)
	if err != nil {
		return "", err
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return "", err
	}

	if err := s.userRepo.UpdatePassword(user.ID, hash); err != nil {
		return "", err
	}
	if err := s.userRepo.AddPasswordHistory(user.ID, hash); err != nil {
		return "", err
	}

	if err := s.authRepo.RevokeAllUserTokens(user.ID); err != nil {
		return "", err
	}
	_ = s.userRepo.UnlockAccount(user.ID)

	if err := s.userRepo.SetForcePasswordChange(user.ID, true); err != nil {
		return "", err
	}

	return password, nil
}

func (s *UserService) ForgotPassword(req ForgotPasswordRequest) (string, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		// Do not leak whether email exists
		return "", nil
	}

	token, tokenHash, err := generateResetToken()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().Add(1 * time.Hour)
	if err := s.authRepo.SavePasswordReset(&models.PasswordReset{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}); err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) ResetPassword(req ResetPasswordRequest) error {
	tokenHash := hashResetToken(req.Token)
	record, err := s.authRepo.FindPasswordReset(tokenHash)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}
	if record.UsedAt != nil {
		return errors.New("reset token already used")
	}
	if record.ExpiresAt.Before(time.Now()) {
		return errors.New("reset token expired")
	}

	user, err := s.userRepo.FindByID(record.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(user.ID, hash); err != nil {
		return err
	}
	if err := s.userRepo.AddPasswordHistory(user.ID, hash); err != nil {
		return err
	}
	if err := s.userRepo.SetForcePasswordChange(user.ID, false); err != nil {
		return err
	}
	if err := s.authRepo.RevokeAllUserTokens(user.ID); err != nil {
		return err
	}
	_ = s.authRepo.MarkPasswordResetUsed(record.ID)

	return nil
}

func contains(list []string, target string) bool {
	for _, s := range list {
		if s == target {
			return true
		}
	}
	return false
}

func generateResetToken() (string, string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}
	token := hex.EncodeToString(tokenBytes)
	return token, auth.HashToken(token), nil
}

func hashResetToken(token string) string {
	return auth.HashToken(token)
}
