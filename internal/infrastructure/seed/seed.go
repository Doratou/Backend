package seed

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Doratou/Backend/bootstrap"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domain_db "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"github.com/Doratou/Backend/pkg/hasher"
)

// Run seeds default users, admin, and manual testing groups into PostgreSQL if they do not exist.
// This operation is completely idempotent and safe to run on every application startup.
func Run(
	ctx context.Context,
	userRepo domain_db.UserRepository,
	db database.Database,
	groupRepo domain_db.GroupRepository,
	inviteRepo domain_db.InviteRepository,
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

	// 3. Seed Manual Testing Users and Groups (Idempotent)
	if db != nil {
		if err := SeedManualTestData(ctx, db, h, cfg); err != nil {
			log.Printf("[SEED NOTICE] SeedManualTestData note: %v", err)
		}
		if err := SeedPollTestData(ctx, db); err != nil {
			log.Printf("[SEED NOTICE] SeedPollTestData note: %v", err)
		}
		if err := SeedLuckyDrawTestData(ctx, db); err != nil {
			log.Printf("[SEED NOTICE] SeedLuckyDrawTestData note: %v", err)
		}
	}
	if err := seedGroup(ctx, userRepo, groupRepo, inviteRepo, h, cfg); err != nil {
		return fmt.Errorf("seed group: %w", err)
	}

	return nil
}

func seedGroup(ctx context.Context, users domain_db.UserRepository, groups domain_db.GroupRepository, invites domain_db.InviteRepository, h hasher.Hasher, cfg bootstrap.SeedConfig) error {
	owner, err := findSeedUser(ctx, users, cfg.UserEmail, cfg.UserPhone)
	if err != nil {
		return err
	}
	admin, err := findSeedUser(ctx, users, cfg.AdminEmail, cfg.AdminPhone)
	if err != nil {
		return err
	}
	if owner == nil || admin == nil {
		log.Printf("[SEED NOTICE] Skipping group seeding: regular user or admin is not configured/available")
		return nil
	}
	owned, _, err := groups.FindByMember(ctx, owner.ID, nil, 100, 0)
	if err != nil {
		return err
	}
	var group *entity.Group
	for i := range owned {
		if owned[i].Title == cfg.GroupTitle && owned[i].CreatedByUserID == owner.ID {
			group = &owned[i]
			break
		}
	}
	if group == nil {
		group = &entity.Group{Title: cfg.GroupTitle, Description: stringPointer(cfg.GroupDescription), CreatedByUserID: owner.ID, Status: enum.GroupStatusActive}
		if err := groups.CreateWithOwner(ctx, group); err != nil {
			return err
		}
		log.Printf("[SEED SUCCESS] Created test group: id=%d owner=%d", group.ID, owner.ID)
	}
	links, err := invites.FindLinksByGroupID(ctx, group.ID)
	if err != nil {
		return err
	}
	var active *entity.LinkInvite
	for i := range links {
		link := &links[i]
		if link.RevokedAt == nil && (link.ExpiresAt == nil || link.ExpiresAt.After(time.Now())) && (link.MaxUses == nil || link.UsedCount < *link.MaxUses) {
			active = link
			break
		}
	}
	if active == nil {
		if cfg.GroupInvitePassword == "" {
			return fmt.Errorf("SEED_GROUP_INVITE_PASSWORD must be configured to create the seeded group invite")
		}
		passHash, err := h.Hash(cfg.GroupInvitePassword)
		if err != nil {
			return err
		}
		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			return err
		}
		token := base64.RawURLEncoding.EncodeToString(raw)
		digest := sha256.Sum256([]byte(token))
		active = &entity.LinkInvite{GroupID: group.ID, TokenHash: hex.EncodeToString(digest[:]), PasswordHash: passHash}
		if err := invites.CreateLink(ctx, group.ID, owner.ID, active); err != nil {
			return err
		}
		log.Printf("[SEED SUCCESS] Created active invite link for test group: group_id=%d", group.ID)
	}
	if admin.ID != owner.ID {
		member, err := groups.IsActiveMember(ctx, group.ID, admin.ID)
		if err != nil {
			return err
		}
		if !member {
			if _, err := invites.JoinGroupWithLink(ctx, active.ID, admin.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func findSeedUser(ctx context.Context, repo domain_db.UserRepository, email, phone string) (*entity.User, error) {
	if email != "" {
		user, err := repo.FindByEmail(ctx, email)
		if err == nil {
			return user, nil
		}
		if !errors.Is(err, exception.ErrUserNotFound) {
			return nil, err
		}
	}
	if phone != "" {
		user, err := repo.FindByPhone(ctx, phone)
		if err == nil {
			return user, nil
		}
		if !errors.Is(err, exception.ErrUserNotFound) {
			return nil, err
		}
	}
	return nil, nil
}

func stringPointer(value string) *string { return &value }

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
