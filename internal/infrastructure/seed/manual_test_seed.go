package seed

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Doratou/Backend/bootstrap"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/infrastructure/database"
	"github.com/Doratou/Backend/pkg/hasher"
	"gorm.io/gorm"
)

type seedUserData struct {
	Name     string
	Email    string
	Phone    string
	Role     enum.UserRole
	Password string
}

// SeedManualTestData populates a rich set of users, groups, and membership relations
// specifically tailored for manual and end-to-end testing of permissions and corner cases.
// It integrates User 1 (Regular User) and User 2 (Admin) across varied group scenarios.
func SeedManualTestData(ctx context.Context, db database.Database, h hasher.Hasher, cfg bootstrap.SeedConfig) error {
	if db == nil {
		return nil
	}
	gdb := db.GetGormDB().WithContext(ctx)

	const defaultTestPassword = "123123123"
	pwdHash, err := h.Hash(defaultTestPassword)
	if err != nil {
		return fmt.Errorf("hash test password: %w", err)
	}

	// 1. Resolve User 1 (Regular) and User 2 (Admin) created by the primary seed
	var user1, user2 entity.User
	if cfg.UserEmail != "" {
		_ = gdb.Where("email = ?", cfg.UserEmail).First(&user1).Error
	}
	if user1.ID == 0 && cfg.UserPhone != "" {
		_ = gdb.Where("phone = ?", cfg.UserPhone).First(&user1).Error
	}
	if user1.ID == 0 {
		_ = gdb.Where("system_role = ?", enum.UserRoleUser).Order("id ASC").First(&user1).Error
	}

	if cfg.AdminEmail != "" {
		_ = gdb.Where("email = ?", cfg.AdminEmail).First(&user2).Error
	}
	if user2.ID == 0 && cfg.AdminPhone != "" {
		_ = gdb.Where("phone = ?", cfg.AdminPhone).First(&user2).Error
	}
	if user2.ID == 0 {
		_ = gdb.Where("system_role = ?", enum.UserRoleAdmin).Order("id ASC").First(&user2).Error
	}

	// 2. Seed Rich Auxiliary Users with password 123123123
	testUsers := []seedUserData{
		// Primary persona aliases (for backward compatibility with previous scripts)
		{Name: "Matin Admin", Email: "m9652973@gmail.com", Phone: "+989970000001", Role: enum.UserRoleAdmin, Password: defaultTestPassword},
		{Name: "Alice Owner", Email: "owner@test.com", Phone: "+989120000001", Role: enum.UserRoleUser, Password: defaultTestPassword},
		{Name: "Bob Member", Email: "member@test.com", Phone: "+989120000002", Role: enum.UserRoleUser, Password: defaultTestPassword},
		{Name: "Charlie Inactive", Email: "inactive@test.com", Phone: "+989120000003", Role: enum.UserRoleUser, Password: defaultTestPassword},
		{Name: "David Outsider", Email: "outsider@test.com", Phone: "+989120000004", Role: enum.UserRoleUser, Password: defaultTestPassword},
		{Name: "Sarah Admin", Email: "admin@test.com", Phone: "+989120000005", Role: enum.UserRoleAdmin, Password: defaultTestPassword},

		// Additional realistic community members
		{Name: "Sara Karimi", Email: "sara@test.com", Phone: "+989120000006", Role: enum.UserRoleUser, Password: defaultTestPassword},
		{Name: "Reza Mohammadi", Email: "reza@test.com", Phone: "+989120000007", Role: enum.UserRoleUser, Password: defaultTestPassword},
		{Name: "Nima Kazemi", Email: "nima@test.com", Phone: "+989120000008", Role: enum.UserRoleUser, Password: defaultTestPassword},
		{Name: "Mahsa Roshan", Email: "mahsa@test.com", Phone: "+989120000009", Role: enum.UserRoleUser, Password: defaultTestPassword},
		{Name: "Sina Rahmani", Email: "sina@test.com", Phone: "+989120000010", Role: enum.UserRoleUser, Password: defaultTestPassword},
	}

	userMap := make(map[string]*entity.User)
	for _, u := range testUsers {
		var existing entity.User
		err := gdb.Where("email = ?", u.Email).First(&existing).Error
		if err == nil {
			// Update password hash to 123123123 to guarantee login works effortlessly
			_ = gdb.Model(&existing).Update("password_hash", pwdHash).Error
			userMap[u.Email] = &existing
			continue
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return fmt.Errorf("check user %s: %w", u.Email, err)
		}

		email := u.Email
		phone := u.Phone
		newUser := entity.User{
			Name:            u.Name,
			Email:           &email,
			Phone:           &phone,
			PasswordHash:    &pwdHash,
			SystemRole:      u.Role,
			Status:          enum.UserStatusActive,
			IsEmailVerified: true,
			IsPhoneVerified: true,
			Wallet:          500000,
		}

		if err := gdb.Create(&newUser).Error; err != nil {
			return fmt.Errorf("create user %s: %w", u.Email, err)
		}
		userMap[u.Email] = &newUser
		log.Printf("[SEED TEST] Created auxiliary test user: %s (id=%d, email=%s)", u.Name, newUser.ID, u.Email)
	}

	alice := userMap["owner@test.com"]
	bob := userMap["member@test.com"]
	charlie := userMap["inactive@test.com"]
	david := userMap["outsider@test.com"]
	sara := userMap["sara@test.com"]
	reza := userMap["reza@test.com"]
	nima := userMap["nima@test.com"]
	mahsa := userMap["mahsa@test.com"]
	sina := userMap["sina@test.com"]

	// Fallback anchor if User 1 / User 2 are missing in a stripped environment
	var primaryUserID int64 = 1
	if user1.ID != 0 {
		primaryUserID = user1.ID
	} else if alice != nil {
		primaryUserID = alice.ID
	}

	var adminUserID int64 = 2
	if user2.ID != 0 {
		adminUserID = user2.ID
	} else if bob != nil {
		adminUserID = bob.ID
	}

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	threeDaysAgo := now.Add(-72 * time.Hour)
	sevenDaysAgo := now.Add(-7 * 24 * time.Hour)

	// =========================================================================
	// 3. Seed Comprehensive Group Matrix (Heavily involving User 1 and User 2)
	// =========================================================================

	// --- Group 1: User 1 is Owner, Active, large group with User 2 and other members ---
	g1Desc := "سفر تابستانی شمال ۱۴۰۳ - مدیریت مشترک هزینه‌ها و گالری عکس‌ها (مالک: کاربر ۱)"
	g1, err := getOrCreateGroup(gdb, "سفر شمال تابستان ۱۴۰۳", &g1Desc, primaryUserID, enum.GroupStatusActive, nil)
	if err != nil {
		return fmt.Errorf("seed group 1: %w", err)
	}
	_ = upsertGroupMember(gdb, g1.ID, primaryUserID, enum.GroupMemberStatusActive, nil)
	_ = upsertGroupMember(gdb, g1.ID, adminUserID, enum.GroupMemberStatusActive, nil)
	if alice != nil {
		_ = upsertGroupMember(gdb, g1.ID, alice.ID, enum.GroupMemberStatusActive, nil)
	}
	if bob != nil {
		_ = upsertGroupMember(gdb, g1.ID, bob.ID, enum.GroupMemberStatusActive, nil)
	}
	if sara != nil {
		_ = upsertGroupMember(gdb, g1.ID, sara.ID, enum.GroupMemberStatusActive, nil)
	}
	if reza != nil {
		_ = upsertGroupMember(gdb, g1.ID, reza.ID, enum.GroupMemberStatusActive, nil)
	}
	if charlie != nil {
		// Charlie is a departed/inactive member
		_ = upsertGroupMember(gdb, g1.ID, charlie.ID, enum.GroupMemberStatusInactive, &yesterday)
	}

	// --- Group 2: User 1 is Owner, Active, Roommates flat squad ---
	g2Desc := "خرج و مخارج همخونه‌ها و خریدهای سوپرمارکتی آپارتمان ۵ (مالک: کاربر ۱)"
	g2, err := getOrCreateGroup(gdb, "خرج و مخارج همخونه‌ها", &g2Desc, primaryUserID, enum.GroupStatusActive, nil)
	if err != nil {
		return fmt.Errorf("seed group 2: %w", err)
	}
	_ = upsertGroupMember(gdb, g2.ID, primaryUserID, enum.GroupMemberStatusActive, nil)
	_ = upsertGroupMember(gdb, g2.ID, adminUserID, enum.GroupMemberStatusActive, nil)
	if david != nil {
		_ = upsertGroupMember(gdb, g2.ID, david.ID, enum.GroupMemberStatusActive, nil)
	}
	if nima != nil {
		_ = upsertGroupMember(gdb, g2.ID, nima.ID, enum.GroupMemberStatusActive, nil)
	}

	// --- Group 3: User 1 is Single Solo Owner (Private Vault) ---
	g3Desc := "صندوقچه و فضای کاملاً شخصی و انفرادی من (مالک: فقط کاربر ۱)"
	g3, err := getOrCreateGroup(gdb, "صندوقچه شخصی من", &g3Desc, primaryUserID, enum.GroupStatusActive, nil)
	if err != nil {
		return fmt.Errorf("seed group 3: %w", err)
	}
	_ = upsertGroupMember(gdb, g3.ID, primaryUserID, enum.GroupMemberStatusActive, nil)

	// --- Group 4: User 2 (Admin) is Owner; User 1 is a regular Member (NOT owner) ---
	g4Desc := "تیم کوهنوردی توچال و دماوند - هماهنگی برنامه‌های هفتگی (مالک: ادمین، کاربر ۱ عضو عادی)"
	g4, err := getOrCreateGroup(gdb, "تیم کوهنوردی توچال", &g4Desc, adminUserID, enum.GroupStatusActive, nil)
	if err != nil {
		return fmt.Errorf("seed group 4: %w", err)
	}
	_ = upsertGroupMember(gdb, g4.ID, adminUserID, enum.GroupMemberStatusActive, nil)
	_ = upsertGroupMember(gdb, g4.ID, primaryUserID, enum.GroupMemberStatusActive, nil)
	if bob != nil {
		_ = upsertGroupMember(gdb, g4.ID, bob.ID, enum.GroupMemberStatusActive, nil)
	}
	if sina != nil {
		_ = upsertGroupMember(gdb, g4.ID, sina.ID, enum.GroupMemberStatusActive, nil)
	}
	if mahsa != nil {
		_ = upsertGroupMember(gdb, g4.ID, mahsa.ID, enum.GroupMemberStatusActive, nil)
	}

	// --- Group 5: Third-party Owner (Alice); both User 1 and User 2 are regular Members ---
	var g5OwnerID int64 = primaryUserID
	if alice != nil {
		g5OwnerID = alice.ID
	}
	g5Desc := "مستندات و هزینه‌های خرید قطعات پروژه فارغ‌التحصیلی (مالک: آلیس)"
	g5, err := getOrCreateGroup(gdb, "پروژه مشترک دانشگاهی", &g5Desc, g5OwnerID, enum.GroupStatusActive, nil)
	if err != nil {
		return fmt.Errorf("seed group 5: %w", err)
	}
	_ = upsertGroupMember(gdb, g5.ID, g5OwnerID, enum.GroupMemberStatusActive, nil)
	_ = upsertGroupMember(gdb, g5.ID, primaryUserID, enum.GroupMemberStatusActive, nil)
	_ = upsertGroupMember(gdb, g5.ID, adminUserID, enum.GroupMemberStatusActive, nil)
	if bob != nil {
		_ = upsertGroupMember(gdb, g5.ID, bob.ID, enum.GroupMemberStatusActive, nil)
	}

	// --- Group 6: User 1 is an INACTIVE / LEFT Member (for testing left-member forbidden guard) ---
	var g6OwnerID int64 = adminUserID
	if bob != nil {
		g6OwnerID = bob.ID
	}
	g6Desc := "دورهمی قدیمی دانشگاه (کاربر ۱ قبلاً عضو بوده و ترک کرده است)"
	g6, err := getOrCreateGroup(gdb, "کافه گردی بچه‌های ورودی ۹۸", &g6Desc, g6OwnerID, enum.GroupStatusActive, nil)
	if err != nil {
		return fmt.Errorf("seed group 6: %w", err)
	}
	_ = upsertGroupMember(gdb, g6.ID, g6OwnerID, enum.GroupMemberStatusActive, nil)
	if alice != nil {
		_ = upsertGroupMember(gdb, g6.ID, alice.ID, enum.GroupMemberStatusActive, nil)
	}
	if sara != nil {
		_ = upsertGroupMember(gdb, g6.ID, sara.ID, enum.GroupMemberStatusActive, nil)
	}
	// User 1 has LEFT this group
	_ = upsertGroupMember(gdb, g6.ID, primaryUserID, enum.GroupMemberStatusInactive, &threeDaysAgo)

	// --- Group 7: Secret DevOps Core Team — User 1 is an OUTSIDER (not a member at all) ---
	g7Desc := "تیم محرمانه زیرساخت کلود و دواپس (کاربر ۱ به این گروه دسترسی ندارد - Outsider)"
	g7, err := getOrCreateGroup(gdb, "تیم زیرساخت کلود و سرورها", &g7Desc, adminUserID, enum.GroupStatusActive, nil)
	if err != nil {
		return fmt.Errorf("seed group 7: %w", err)
	}
	_ = upsertGroupMember(gdb, g7.ID, adminUserID, enum.GroupMemberStatusActive, nil)
	if david != nil {
		_ = upsertGroupMember(gdb, g7.ID, david.ID, enum.GroupMemberStatusActive, nil)
	}
	if sina != nil {
		_ = upsertGroupMember(gdb, g7.ID, sina.ID, enum.GroupMemberStatusActive, nil)
	}

	// --- Group 8: Archived / Inactive Group (for testing group freeze and read-only guards) ---
	g8Desc := "گروه آرشیو و فریز شده پس از پایان مسابقات هکاتون (وضعیت: غیرفعال)"
	g8, err := getOrCreateGroup(gdb, "هکاتون استارتاپی پاییز ۱۴۰۲", &g8Desc, primaryUserID, enum.GroupStatusInactive, &sevenDaysAgo)
	if err != nil {
		return fmt.Errorf("seed group 8: %w", err)
	}
	_ = upsertGroupMember(gdb, g8.ID, primaryUserID, enum.GroupMemberStatusActive, nil)
	_ = upsertGroupMember(gdb, g8.ID, adminUserID, enum.GroupMemberStatusActive, nil)
	if alice != nil {
		_ = upsertGroupMember(gdb, g8.ID, alice.ID, enum.GroupMemberStatusActive, nil)
	}
	if reza != nil {
		_ = upsertGroupMember(gdb, g8.ID, reza.ID, enum.GroupMemberStatusActive, nil)
	}

	log.Printf("[SEED TEST] Comprehensive test matrix populated successfully: 8 Groups covering Owner, Regular Member, Solo, Inactive Member, Outsider, and Archived states.")
	return nil
}

func getOrCreateGroup(gdb *gorm.DB, title string, desc *string, creatorID int64, status enum.GroupStatus, archivedAt *time.Time) (*entity.Group, error) {
	var g entity.Group
	err := gdb.Where("title = ? AND created_by_user_id = ?", title, creatorID).First(&g).Error
	if err == nil {
		_ = gdb.Model(&g).Updates(map[string]interface{}{
			"description": desc,
			"status":      status,
			"archived_at": archivedAt,
		}).Error
		return &g, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	g = entity.Group{
		Title:           title,
		Description:     desc,
		CreatedByUserID: creatorID,
		Status:          status,
		ArchivedAt:      archivedAt,
	}
	if err := gdb.Create(&g).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func upsertGroupMember(gdb *gorm.DB, groupID, userID int64, status enum.GroupMemberStatus, leftAt *time.Time) error {
	var m entity.GroupMember
	err := gdb.Where("group_id = ? AND user_id = ?", groupID, userID).First(&m).Error
	if err == nil {
		return gdb.Model(&m).Updates(map[string]interface{}{
			"status":  status,
			"left_at": leftAt,
		}).Error
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	newMember := entity.GroupMember{
		GroupID:  groupID,
		UserID:   userID,
		Status:   status,
		JoinedAt: time.Now().Add(-14 * 24 * time.Hour),
		LeftAt:   leftAt,
	}
	return gdb.Create(&newMember).Error
}
