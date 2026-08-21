package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/shakil5281/peoplehub-api/internal/auth"
	"github.com/shakil5281/peoplehub-api/internal/config"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/models"
	"gorm.io/gorm"
)

func main() {
	email := os.Getenv("SUPERADMIN_EMAIL")
	password := os.Getenv("SUPERADMIN_PASSWORD")
	name := os.Getenv("SUPERADMIN_NAME")

	if email == "" {
		email = "superadmin@peoplehub.com"
	}
	if password == "" {
		password = "superadmin1234"
	}
	if name == "" {
		name = "Super Admin"
	}

	cfg := config.Load()
	database.Connect(cfg)
	db := database.DB

	role := seedRole(db)
	seedPermissions(db, role)
	user := seedUser(db, email, password, name)
	assignRole(db, user, role)

	fmt.Println("\nSuperadmin created successfully!")
	fmt.Printf("  Email:    %s\n", email)
	fmt.Printf("  Password: %s\n", password)
}

func seedRole(db *gorm.DB) *models.Role {
	var role models.Role
	err := db.Where("name = ? AND is_system = ?", "super_admin", true).First(&role).Error
	if err == nil {
		fmt.Println("Superadmin role already exists, skipping creation")
		return &role
	}

	role = models.Role{
		Name:        "super_admin",
		Description: "Super administrator with full system access",
		IsSystem:    true,
	}
	if err := db.Create(&role).Error; err != nil {
		log.Fatal("Failed to create superadmin role:", err)
	}
	fmt.Println("Created superadmin role")
	return &role
}

func seedPermissions(db *gorm.DB, role *models.Role) {
	resources := []string{
		"users", "roles", "permissions",
		"companies", "departments", "sections", "designations", "lines", "groups", "floors",
		"employees",
		"shifts",
		"attendance",
		"data_logs",
		"divisions", "districts", "upazilas", "unions",
		"dashboard", "audit_logs", "settings",
	}
	actions := []string{"create", "read", "update", "delete", "manage"}

	for _, res := range resources {
		for _, act := range actions {
			permName := res + "." + act
			var perm models.Permission
			err := db.Where("resource = ? AND action = ?", res, act).First(&perm).Error
			if err != nil {
				perm = models.Permission{
					Resource:    res,
					Action:      act,
					Description: fmt.Sprintf("Can %s %s", act, res),
				}
				if err := db.Create(&perm).Error; err != nil {
					log.Printf("Failed to create permission %s: %v", permName, err)
					continue
				}
			}

			var rp models.RolePermission
			err = db.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).First(&rp).Error
			if err != nil {
				rp = models.RolePermission{
					RoleID:       role.ID,
					PermissionID: perm.ID,
				}
				if err := db.Create(&rp).Error; err != nil {
					log.Printf("Failed to assign permission %s to role: %v", permName, err)
				}
			}
		}
	}
	fmt.Println("Seeded all permissions and assigned to superadmin role")
}

func seedUser(db *gorm.DB, email, password, name string) *models.User {
	var existing models.User
	err := db.Where("email = ?", email).First(&existing).Error
	if err == nil {
		fmt.Printf("User %s already exists, skipping creation\n", email)
		return &existing
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	now := time.Now()
	user := models.User{
		Email:           email,
		PasswordHash:    hash,
		Name:            name,
		Status:          "active",
		EmailVerifiedAt: &now,
		ForcePasswordChange: false,
	}
	if err := db.Create(&user).Error; err != nil {
		log.Fatal("Failed to create superadmin user:", err)
	}
	fmt.Printf("Created superadmin user: %s\n", email)
	return &user
}

func assignRole(db *gorm.DB, user *models.User, role *models.Role) {
	var ur models.UserRole
	err := db.Where("user_id = ? AND role_id = ?", user.ID, role.ID).First(&ur).Error
	if err == nil {
		fmt.Println("Role already assigned to user")
		return
	}

	ur = models.UserRole{
		UserID: user.ID,
		RoleID: role.ID,
	}
	if err := db.Create(&ur).Error; err != nil {
		log.Fatal("Failed to assign superadmin role to user:", err)
	}
	fmt.Printf("Assigned superadmin role to user %s\n", user.Email)
}
