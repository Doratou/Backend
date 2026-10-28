package seed

import (
	"context"
	"fmt"
	"log"

	"github.com/Doratou/Backend/bootstrap"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/pkg/hasher"
)

// Run seeds default regular user and admin user into PostgreSQL if they do not exist.
// This operation is completely idempotent and safe to run on every application startup.
func Run(
	ctx context.Context,
	userRepo domain_db.UserRepository,
	h hasher.Hasher,
	cfg bootstrap.SeedConfig,
) error {
	// 1. Seed Regular User
	if err := seedUser(ctx, userRepo, h, cfg.UserName, cfg.UserEmail, cfg.UserPhone, cfg.UserPassword, enum.UserRoleUser); err != nil {
		return fmt.Errorf("seed regular user: %w", err)
	}

	// 2. Seed Admin User
	if err := seedUser(ctx, userRepo, h, cfg.AdminName, cfg.AdminEmail, cfg.AdminPhone, cfg.AdminPassword, enum.UserRoleAdmin); err != nil {
		return fmt.Errorf("seed admin user: %w", err)
	}

	return nil
}

func seedUser(
	ctx context.Context,
	userRepo domain_db.UserRepository,
	h hasher.Hasher,
	name, email, phone, rawPassword string,
	role enum.UserRole,
) error {
	if email == "" && phone == "" {
		log.Printf("[SEED NOTICE] Skipping %s seeding: neither email nor phone is configured", role)
		return nil
	}

	// Check if user already exists by email
	if email != "" {
		exists, err := userRepo.ExistsByEmail(ctx, email)
		if err != nil {
			return err
		}
		if exists {
			log.Printf("[SEED NOTICE] %s already exists with email %s, skipping creation", role, email)
			return nil
		}
	}

	// Check if user already exists by phone
	if phone != "" {
		exists, err := userRepo.ExistsByPhone(ctx, phone)
		if err != nil {
			return err
		}
		if exists {
			log.Printf("[SEED NOTICE] %s already exists with phone %s, skipping creation", role, phone)
			return nil
		}
	}

	if rawPassword == "" {
		rawPassword = "Password@123"
	}

	hashedPassword, err := h.Hash(rawPassword)
	if err != nil {
		return fmt.Errorf("hash password for %s (%s): %w", role, name, err)
	}

	var emailPtr, phonePtr *string
	if email != "" {
		emailPtr = &email
	}
	if phone != "" {
		phonePtr = &phone
	}

	user := &entity.User{
		Name:               name,
		PasswordHash:       hashedPassword,
		Email:              emailPtr,
		Phone:              phonePtr,
		IsEmailVerified:    emailPtr != nil,
		IsPhoneVerified:    phonePtr != nil,
		AvatarAttachmentID: nil,
		SystemRole:         role,
		Status:             enum.UserStatusActive,
		Wallet:             0,
		TelegramChatID:     nil,
		TelegramUserID:     nil,
	}

	if err := userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("create %s (%s): %w", role, name, err)
	}

	log.Printf("[SEED SUCCESS] Created %s: name=%s, email=%s, phone=%s, id=%d", role, name, email, phone, user.ID)
	return nil
}
