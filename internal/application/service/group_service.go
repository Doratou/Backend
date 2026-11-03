package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domainrepo "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/pkg/hasher"
)

type groupService struct {
	groups              domainrepo.GroupRepository
	hasher              hasher.Hasher
	inviteEncryptionKey []byte
}

func NewGroupService(groups domainrepo.GroupRepository, passwordHasher hasher.Hasher, inviteEncryptionSecret string) *groupService {
	if passwordHasher == nil {
		passwordHasher = hasher.NewBcryptHasher()
	}
	key := sha256.Sum256([]byte("dor atou:group-invite-token:v1:" + inviteEncryptionSecret))
	var encryptionKey []byte
	if inviteEncryptionSecret != "" {
		encryptionKey = key[:]
	}
	return &groupService{groups: groups, hasher: passwordHasher, inviteEncryptionKey: encryptionKey}
}

func (s *groupService) Create(ctx context.Context, userID int64, req dto.CreateGroupRequest) (*dto.GroupResponse, error) {
	group := &entity.Group{Title: req.Title, Description: req.Description, CreatedByUserID: userID,
		Status: enum.GroupStatusActive}
	// TODO: Connect the group cover upload/attachment service when it is implemented.
	if err := s.groups.CreateWithOwner(ctx, group); err != nil {
		return nil, err
	}
	res := toGroupResponse(group, 1)
	return &res, nil
}

func (s *groupService) ListMine(ctx context.Context, userID int64, status *string, limit, offset int) ([]dto.GroupResponse, int64, error) {
	groupStatus, err := parseGroupStatus(status)
	if err != nil {
		return nil, 0, err
	}
	groups, total, err := s.groups.FindByMember(ctx, userID, groupStatus, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.GroupResponse, 0, len(groups))
	for i := range groups {
		result = append(result, toGroupResponse(&groups[i], 0))
	}
	return result, total, nil
}

func (s *groupService) ListAll(ctx context.Context, status *string, limit, offset int) ([]dto.GroupResponse, int64, error) {
	groupStatus, err := parseGroupStatus(status)
	if err != nil {
		return nil, 0, err
	}
	groups, total, err := s.groups.FindAll(ctx, groupStatus, limit, offset)

	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.GroupResponse, 0, len(groups))
	for i := range groups {
		result = append(result, toGroupResponse(&groups[i], 0))
	}
	return result, total, nil
}

func parseGroupStatus(status *string) (*enum.GroupStatus, error) {
	if status == nil {
		return nil, nil
	}
	value := strings.ToUpper(*status)
	if value != string(enum.GroupStatusActive) && value != string(enum.GroupStatusInactive) {
		return nil, exception.ErrGroupInvalidStatus
	}
	parsed := enum.GroupStatus(value)
	return &parsed, nil
}

func (s *groupService) ownedGroup(ctx context.Context, userID int64, role enum.UserRole, groupID int64) (*entity.Group, error) {
	group, _, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if role != enum.UserRoleAdmin && group.CreatedByUserID != userID {
		return nil, exception.ErrGroupForbidden
	}
	return group, nil
}

func (s *groupService) Update(ctx context.Context, userID int64, role enum.UserRole, groupID int64, req dto.UpdateGroupRequest) (*dto.GroupResponse, error) {
	if req.Title == nil && req.Description == nil {
		return nil, exception.ErrGroupNoFieldsToUpdate
	}
	group, err := s.ownedGroup(ctx, userID, role, groupID)
	if err != nil {
		return nil, err
	}
	if req.Title != nil {
		group.Title = *req.Title
	}
	if req.Description != nil {
		group.Description = req.Description
	}
	if err := s.groups.Update(ctx, group); err != nil {
		return nil, err
	}
	updated, count, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	result := toGroupResponse(updated, count)
	return &result, nil
}

func (s *groupService) ChangeCover(ctx context.Context, userID int64, role enum.UserRole, groupID int64) error {
	if _, err := s.ownedGroup(ctx, userID, role, groupID); err != nil {
		return err
	}
	// TODO: Implement cover image changes when the attachment service is available.
	return nil
}

func (s *groupService) Archive(ctx context.Context, userID int64, role enum.UserRole, groupID int64) error {
	if _, err := s.ownedGroup(ctx, userID, role, groupID); err != nil {
		return err
	}
	return s.groups.SetStatus(ctx, groupID, enum.GroupStatusInactive)
}

func (s *groupService) Delete(ctx context.Context, userID int64, role enum.UserRole, groupID int64) error {
	if _, err := s.ownedGroup(ctx, userID, role, groupID); err != nil {
		return err
	}
	return s.groups.Delete(ctx, groupID)
}

func (s *groupService) Clone(ctx context.Context, userID int64, role enum.UserRole, groupID int64, req dto.CloneGroupRequest) (*dto.GroupResponse, error) {
	source, _, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if role != enum.UserRoleAdmin && source.CreatedByUserID != userID {
		return nil, exception.ErrGroupOwnerRequired
	}
	if source.Status != enum.GroupStatusInactive {
		return nil, exception.ErrGroupMustBeInactiveToClone
	}
	clone := &entity.Group{Title: req.Title, Description: req.Description, CreatedByUserID: userID, Status: enum.GroupStatusActive}
	if err := s.groups.CloneWithMembers(ctx, groupID, clone); err != nil {
		return nil, err
	}
	result := toGroupResponse(clone, 1)
	return &result, nil
}

func (s *groupService) Get(ctx context.Context, userID int64, role enum.UserRole, groupID int64) (*dto.GroupResponse, error) {
	group, count, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	// Groups are only visible to active members.
	if role != enum.UserRoleAdmin {
		isMember, err := s.groups.IsActiveMember(ctx, groupID, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, exception.ErrGroupForbidden
		}
	}
	result := toGroupResponse(group, count)
	members, err := s.groups.ListActiveMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}
	result.Members = toGroupMemberResponses(members, group.CreatedByUserID)
	result.Invites, err = s.groupInviteResponses(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *groupService) ListMembers(ctx context.Context, userID int64, role enum.UserRole, groupID int64) ([]dto.GroupMemberResponse, error) {
	group, _, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if role != enum.UserRoleAdmin {
		isMember, err := s.groups.IsActiveMember(ctx, groupID, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, exception.ErrGroupForbidden
		}
	}
	members, err := s.groups.ListActiveMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return toGroupMemberResponses(members, group.CreatedByUserID), nil
}

func (s *groupService) RemoveMember(ctx context.Context, userID int64, role enum.UserRole, groupID, memberID int64) error {
	group, _, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if role != enum.UserRoleAdmin && group.CreatedByUserID != userID {
		return exception.ErrGroupOwnerRequired
	}
	member, err := s.groups.FindActiveMemberByID(ctx, groupID, memberID)
	if err != nil {
		return err
	}
	if member.UserID == group.CreatedByUserID {
		return exception.ErrGroupOwnerCannotBeRemoved
	}
	return s.groups.SetMemberInactiveByID(ctx, groupID, memberID)
}

func (s *groupService) LeaveGroup(ctx context.Context, userID, groupID int64) error {
	group, _, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group.CreatedByUserID == userID {
		return exception.ErrGroupOwnerCannotLeave
	}
	return s.groups.SetMemberInactive(ctx, groupID, userID)
}

func (s *groupService) TransferOwnership(ctx context.Context, userID int64, role enum.UserRole, groupID, newOwnerID int64) error {
	group, _, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if role != enum.UserRoleAdmin && group.CreatedByUserID != userID {
		return exception.ErrGroupOwnerRequired
	}
	return s.groups.TransferOwnership(ctx, groupID, newOwnerID)
}

func (s *groupService) CreateInvite(ctx context.Context, userID int64, role enum.UserRole, groupID int64, req dto.CreateGroupInviteRequest) (*dto.GroupInviteResponse, error) {
	group, _, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if err := requireInviteManager(group, userID, role); err != nil {
		return nil, err
	}
	if req.Password == "" || len(req.Password) > 72 {
		return nil, exception.ErrInviteInvalidPasswordFormat
	}
	if req.MaxUses != nil && *req.MaxUses <= 0 {
		return nil, exception.ErrInviteInvalidMaxUses
	}
	if req.ExpiresAt != nil && !req.ExpiresAt.After(time.Now()) {
		return nil, exception.ErrInviteInvalidExpiration
	}
	passwordHash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	tokenCiphertext, err := s.encryptInviteToken(token)
	if err != nil {
		return nil, exception.Wrap(exception.ErrInternal, err)
	}
	invite := &entity.GroupInvite{
		GroupID: groupID, TokenHash: hashInviteToken(token), TokenCiphertext: tokenCiphertext, PasswordHash: passwordHash,
		ExpiresAt: req.ExpiresAt, MaxUses: req.MaxUses,
	}
	if err := s.groups.CreateInvite(ctx, groupID, userID, invite); err != nil {
		return nil, err
	}
	response := toGroupInviteResponse(invite)
	response.Token = token
	response.InviteURL = "/api/v1/groups/invites/" + token + "/join"
	return &response, nil
}

func (s *groupService) DeactivateInvite(ctx context.Context, userID int64, role enum.UserRole, groupID, inviteID int64) error {
	if _, err := s.authorizedInvite(ctx, userID, role, groupID, inviteID); err != nil {
		return err
	}
	return s.groups.RevokeInvite(ctx, groupID, inviteID)
}

func (s *groupService) ChangeInvitePassword(ctx context.Context, userID int64, role enum.UserRole, groupID, inviteID int64, password string) error {
	if _, err := s.authorizedInvite(ctx, userID, role, groupID, inviteID); err != nil {
		return err
	}
	if password == "" || len(password) > 72 {
		return exception.ErrInviteInvalidPasswordFormat
	}
	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return exception.Wrap(exception.ErrInternal, err)
	}
	return s.groups.UpdateInvitePassword(ctx, groupID, inviteID, passwordHash)
}

func (s *groupService) UpdateInvite(ctx context.Context, userID int64, role enum.UserRole, groupID, inviteID int64, req dto.UpdateGroupInviteRequest) (*dto.GroupInviteResponse, error) {
	invite, err := s.authorizedInvite(ctx, userID, role, groupID, inviteID)
	if err != nil {
		return nil, err
	}
	if req.ExpiresAt == nil && req.MaxUses == nil {
		return nil, exception.ErrInviteNoFieldsToUpdate
	}
	if req.ExpiresAt != nil && !req.ExpiresAt.After(time.Now()) {
		return nil, exception.ErrInviteInvalidExpiration
	}
	if req.MaxUses != nil {
		if *req.MaxUses <= 0 {
			return nil, exception.ErrInviteInvalidMaxUses
		}
		if *req.MaxUses < invite.UsedCount {
			return nil, exception.ErrInviteInvalidMaxUses
		}
	}
	invite.ExpiresAt = req.ExpiresAt
	invite.MaxUses = req.MaxUses
	if err := s.groups.UpdateInvite(ctx, invite); err != nil {
		return nil, err
	}
	updated, err := s.groups.FindInviteByID(ctx, groupID, inviteID)
	if err != nil {
		return nil, err
	}
	response := toGroupInviteResponse(updated)
	return &response, nil
}

func (s *groupService) JoinGroupByInvite(ctx context.Context, userID int64, token, password string) (*dto.GroupResponse, error) {
	if token == "" || password == "" {
		return nil, exception.ErrInviteJoinPayload
	}
	invite, err := s.groups.FindInviteByTokenHash(ctx, hashInviteToken(token))
	if err != nil {
		return nil, err
	}
	if invite.RevokedAt != nil {
		return nil, exception.ErrInviteRevoked
	}
	if invite.ExpiresAt != nil && !invite.ExpiresAt.After(time.Now()) {
		return nil, exception.ErrInviteExpired
	}
	if invite.MaxUses != nil && invite.UsedCount >= *invite.MaxUses {
		return nil, exception.ErrInviteUsesExhausted
	}
	if err := s.hasher.Compare(password, invite.PasswordHash); err != nil {
		return nil, exception.ErrInvitePasswordInvalid
	}
	groupID, err := s.groups.JoinGroupWithInvite(ctx, invite.ID, userID)
	if err != nil {
		return nil, err
	}
	group, count, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	response := toGroupResponse(group, count)
	members, err := s.groups.ListActiveMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}
	response.Members = toGroupMemberResponses(members, group.CreatedByUserID)
	response.Invites, err = s.groupInviteResponses(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *groupService) ListInvites(ctx context.Context, userID int64, role enum.UserRole, groupID int64) ([]dto.GroupInviteResponse, error) {
	if _, _, err := s.groups.FindByID(ctx, groupID); err != nil {
		return nil, err
	}
	if role != enum.UserRoleAdmin {
		isMember, err := s.groups.IsActiveMember(ctx, groupID, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, exception.ErrGroupForbidden
		}
	}
	return s.groupInviteResponses(ctx, groupID)
}

func (s *groupService) groupInviteResponses(ctx context.Context, groupID int64) ([]dto.GroupInviteResponse, error) {
	invites, err := s.groups.FindInvitesByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	responses := make([]dto.GroupInviteResponse, 0, len(invites))
	for i := range invites {
		response := toGroupInviteResponse(&invites[i])
		if invites[i].TokenCiphertext != "" {
			token, err := s.decryptInviteToken(invites[i].TokenCiphertext)
			if err != nil {
				return nil, exception.Wrap(exception.ErrInternal, err)
			}
			response.Token = token
			response.InviteURL = "/api/v1/groups/invites/" + token + "/join"
		}
		responses = append(responses, response)
	}
	return responses, nil
}

func (s *groupService) encryptInviteToken(token string) (string, error) {
	aead, err := s.inviteAEAD()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := aead.Seal(nonce, nonce, []byte(token), nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func (s *groupService) decryptInviteToken(encoded string) (string, error) {
	aead, err := s.inviteAEAD()
	if err != nil {
		return "", err
	}
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	if len(data) < aead.NonceSize() {
		return "", errors.New("invalid invite token ciphertext")
	}
	nonce, ciphertext := data[:aead.NonceSize()], data[aead.NonceSize():]
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func (s *groupService) inviteAEAD() (cipher.AEAD, error) {
	if len(s.inviteEncryptionKey) == 0 {
		return nil, errors.New("group invite encryption secret is not configured")
	}
	block, err := aes.NewCipher(s.inviteEncryptionKey)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func (s *groupService) authorizedInvite(ctx context.Context, userID int64, role enum.UserRole, groupID, inviteID int64) (*entity.GroupInvite, error) {
	group, _, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if err := requireInviteManager(group, userID, role); err != nil {
		return nil, err
	}
	return s.groups.FindInviteByID(ctx, groupID, inviteID)
}

func requireInviteManager(group *entity.Group, userID int64, role enum.UserRole) error {
	if role != enum.UserRoleAdmin && group.CreatedByUserID != userID {
		return exception.ErrGroupOwnerRequired
	}
	return nil
}

func hashInviteToken(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

func toGroupInviteResponse(invite *entity.GroupInvite) dto.GroupInviteResponse {
	return dto.GroupInviteResponse{
		ID: invite.ID, GroupID: invite.GroupID, ExpiresAt: invite.ExpiresAt, MaxUses: invite.MaxUses,
		UsedCount: invite.UsedCount, RevokedAt: invite.RevokedAt, CreatedAt: invite.CreatedAt,
	}
}

func toGroupMemberResponses(members []entity.GroupMember, ownerID int64) []dto.GroupMemberResponse {
	responses := make([]dto.GroupMemberResponse, 0, len(members))
	for _, member := range members {
		name := ""
		if member.User != nil {
			name = member.User.Name
		}
		role := enum.GroupMemberRoleMember
		if member.UserID == ownerID {
			role = enum.GroupMemberRoleOwner
		}
		responses = append(responses, dto.GroupMemberResponse{
			MemberID: member.ID, UserID: member.UserID, Name: name, Role: string(role), JoinedAt: member.JoinedAt,
		})
	}
	return responses
}

func toGroupResponse(group *entity.Group, count int64) dto.GroupResponse {
	return dto.GroupResponse{ID: group.ID, Title: group.Title, Description: group.Description,
		CreatedByUserID: group.CreatedByUserID, CoverAttachmentID: group.CoverAttachmentID,
		Status: string(group.Status), IsPremium: group.IsPremium, ArchivedAt: group.ArchivedAt,
		CreatedAt: group.CreatedAt, UpdatedAt: group.UpdatedAt, MemberCount: count}
}
