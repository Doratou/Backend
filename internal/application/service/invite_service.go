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
	"log"
	"time"

	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	domainrepo "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/pkg/hasher"
)

type inviteService struct {
	groups              domainrepo.GroupRepository
	invites             domainrepo.InviteRepository
	friends             domainrepo.FriendshipRepository
	users               domainrepo.UserRepository
	hasher              hasher.Hasher
	inviteEncryptionKey []byte
	rejectCooldown      time.Duration
}

func NewInviteService(groups domainrepo.GroupRepository, invites domainrepo.InviteRepository, friends domainrepo.FriendshipRepository, users domainrepo.UserRepository, passwordHasher hasher.Hasher, inviteEncryptionSecret string, rejectCooldown time.Duration) *inviteService {
	if passwordHasher == nil {
		passwordHasher = hasher.NewBcryptHasher()
	}
	key := sha256.Sum256([]byte("dor atou:group-invite-token:v1:" + inviteEncryptionSecret))
	var encryptionKey []byte
	if inviteEncryptionSecret != "" {
		encryptionKey = key[:]
	}
	if rejectCooldown <= 0 {
		rejectCooldown = 24 * time.Hour
		log.Printf("[CONFIG NOTICE] GROUP_DIRECT_INVITE_REJECT_COOLDOWN is not configured or non-positive, applying default: %v", rejectCooldown)
	}
	return &inviteService{groups: groups, invites: invites, friends: friends, users: users, hasher: passwordHasher, inviteEncryptionKey: encryptionKey, rejectCooldown: rejectCooldown}
}

func (s *inviteService) SendDirect(ctx context.Context, userID int64, role enum.UserRole, req dto.CreateDirectInviteRequest) ([]dto.DirectInviteSendResult, error) {
	if userID <= 0 || req.GroupID <= 0 || len(req.UserIDs) == 0 || len(req.UserIDs) > 100 {
		return nil, exception.ErrDirectInviteInvalidPayload
	}
	if role != enum.UserRoleUser {
		return nil, exception.ErrDirectInviteForbidden
	}
	group, _, err := s.groups.FindByID(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}
	if group.CreatedByUserID != userID {
		return nil, exception.ErrDirectInviteForbidden
	}
	if group.Status != enum.GroupStatusActive {
		return nil, exception.ErrGroupInactive
	}

	results := make([]dto.DirectInviteSendResult, 0, len(req.UserIDs))
	seen := make(map[int64]bool, len(req.UserIDs))
	for _, recipientID := range req.UserIDs {
		result := dto.DirectInviteSendResult{UserID: recipientID}
		if recipientID <= 0 {
			result.Error = &dto.DirectInviteError{Code: exception.ErrDirectInviteInvalidPayload.Code, Message: "user_id must be a positive integer"}
		} else if seen[recipientID] {
			result.Error = &dto.DirectInviteError{Code: "DUPLICATE_RECIPIENT", Message: "user was included more than once"}
		} else {
			seen[recipientID] = true
			result.Error = s.sendOne(ctx, userID, group.ID, recipientID, &result)
			result.Success = result.Error == nil
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *inviteService) sendOne(ctx context.Context, senderID, groupID, recipientID int64, result *dto.DirectInviteSendResult) *dto.DirectInviteError {
	if recipientID == senderID {
		return &dto.DirectInviteError{Code: "DIRECT_INVITE_SELF", Message: "you cannot invite yourself"}
	}
	friendship, err := s.friends.FindByUserPair(ctx, senderID, recipientID)
	if err != nil {
		return directInviteError(err)
	}
	if friendship.Status != enum.FriendshipStatusAccepted {
		return &dto.DirectInviteError{Code: exception.ErrDirectInviteNotFriend.Code, Message: exception.ErrDirectInviteNotFriend.Message}
	}
	user, err := s.users.FindByID(ctx, recipientID)
	if err != nil {
		return directInviteError(err)
	}
	if user.Status != enum.UserStatusActive {
		return &dto.DirectInviteError{Code: exception.ErrUserInactive.Code, Message: exception.ErrUserInactive.Message}
	}
	invite := &entity.DirectInvite{GroupID: groupID, SenderUserID: senderID, RecipientUserID: recipientID, Status: enum.DirectInviteStatusPending}
	if err := s.invites.CreateDirect(ctx, invite, s.rejectCooldown); err != nil {
		return directInviteError(err)
	}
	result.Invite = toDirectInviteResponse(invite)
	return nil
}

func (s *inviteService) RespondToDirect(ctx context.Context, userID, inviteID int64, response string) (*dto.DirectInviteResponse, error) {
	if inviteID <= 0 {
		return nil, exception.ErrInviteInvalidID
	}
	accept := response == string(enum.DirectInviteStatusAccepted)
	if !accept && response != string(enum.DirectInviteStatusRejected) {
		return nil, exception.ErrDirectInviteInvalidResponse
	}
	invite, err := s.invites.RespondDirect(ctx, inviteID, userID, accept)
	if err != nil {
		return nil, err
	}
	return toDirectInviteResponse(invite), nil
}

func (s *inviteService) ListMyDirectInvites(ctx context.Context, userID int64) ([]dto.DirectInviteResponse, error) {
	invites, err := s.invites.FindDirectByRecipient(ctx, userID)
	if err != nil {
		return nil, err
	}
	return mapDirectInvites(invites), nil
}

func (s *inviteService) ListGroupDirectInvites(ctx context.Context, userID int64, role enum.UserRole, groupID int64) ([]dto.DirectInviteResponse, error) {
	if groupID <= 0 {
		return nil, exception.ErrGroupInvalidID
	}
	group, _, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if role != enum.UserRoleUser || group.CreatedByUserID != userID {
		return nil, exception.ErrGroupForbidden
	}
	invites, err := s.invites.FindDirectByGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return mapDirectInvites(invites), nil
}

func directInviteError(err error) *dto.DirectInviteError {
	if errors.Is(err, exception.ErrDomainFriendshipNotFound) {
		return &dto.DirectInviteError{Code: exception.ErrDirectInviteNotFriend.Code, Message: exception.ErrDirectInviteNotFriend.Message}
	}
	var appErr *exception.AppError
	if !errors.As(err, &appErr) {
		return &dto.DirectInviteError{Code: exception.ErrInternal.Code, Message: exception.ErrInternal.Message}
	}
	return &dto.DirectInviteError{Code: appErr.Code, Message: appErr.Message}
}

func (s *inviteService) CreateLink(ctx context.Context, userID int64, role enum.UserRole, groupID int64, req dto.CreateLinkInviteRequest) (*dto.LinkInviteResponse, error) {
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
	if err := validateInviteParams(req.ExpiresAt, req.MaxUses); err != nil {
		return nil, err
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
	invite := &entity.LinkInvite{
		GroupID: groupID, TokenHash: hashInviteToken(token), TokenCiphertext: tokenCiphertext, PasswordHash: passwordHash,
		ExpiresAt: req.ExpiresAt, MaxUses: req.MaxUses,
	}
	if err := s.invites.CreateLink(ctx, groupID, userID, invite); err != nil {
		return nil, err
	}
	response := toLinkInviteResponse(invite)
	response.Token = token
	response.InviteURL = "/api/v1/invites/links/" + token + "/join"
	return &response, nil
}

func (s *inviteService) DeactivateLink(ctx context.Context, userID int64, role enum.UserRole, groupID, inviteID int64) error {
	if _, err := s.authorizedInvite(ctx, userID, role, groupID, inviteID); err != nil {
		return err
	}
	return s.invites.RevokeLink(ctx, groupID, inviteID)
}

func (s *inviteService) ChangeLinkPassword(ctx context.Context, userID int64, role enum.UserRole, groupID, inviteID int64, password string) error {
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
	return s.invites.UpdateLinkPassword(ctx, groupID, inviteID, passwordHash)
}

func (s *inviteService) UpdateLink(ctx context.Context, userID int64, role enum.UserRole, groupID, inviteID int64, req dto.UpdateLinkInviteRequest) (*dto.LinkInviteResponse, error) {
	invite, err := s.authorizedInvite(ctx, userID, role, groupID, inviteID)
	if err != nil {
		return nil, err
	}
	if req.ExpiresAt == nil && req.MaxUses == nil {
		return nil, exception.ErrInviteNoFieldsToUpdate
	}
	if err := validateInviteParams(req.ExpiresAt, req.MaxUses); err != nil {
		return nil, err
	}
	if req.MaxUses != nil && *req.MaxUses < invite.UsedCount {
		return nil, exception.ErrInviteInvalidMaxUses
	}
	invite.ExpiresAt = req.ExpiresAt
	invite.MaxUses = req.MaxUses
	if err := s.invites.UpdateLink(ctx, invite); err != nil {
		return nil, err
	}
	updated, err := s.invites.FindLinkByID(ctx, groupID, inviteID)
	if err != nil {
		return nil, err
	}
	response := toLinkInviteResponse(updated)
	return &response, nil
}

func (s *inviteService) JoinByLink(ctx context.Context, userID int64, token, password string) (*dto.GroupResponse, error) {
	if token == "" || password == "" {
		return nil, exception.ErrInviteJoinPayload
	}
	invite, err := s.invites.FindLinkByTokenHash(ctx, hashInviteToken(token))
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
	groupID, err := s.invites.JoinGroupWithLink(ctx, invite.ID, userID)
	if err != nil {
		return nil, err
	}
	group, count, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	res, err := buildInviteJoinGroupResponse(ctx, s.groups, group, count)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *inviteService) ListGroupLinks(ctx context.Context, userID int64, role enum.UserRole, groupID int64) ([]dto.LinkInviteResponse, error) {
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

// =================================================================
// SECURITY & ENCRYPTION HELPERS
// =================================================================
func (s *inviteService) encryptInviteToken(token string) (string, error) {
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

func (s *inviteService) decryptInviteToken(encoded string) (string, error) {
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

func (s *inviteService) inviteAEAD() (cipher.AEAD, error) {
	if len(s.inviteEncryptionKey) == 0 {
		return nil, errors.New("group invite encryption secret is not configured")
	}
	block, err := aes.NewCipher(s.inviteEncryptionKey)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func hashInviteToken(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

// =================================================================
// AUTHORIZATION & VALIDATION HELPERS
// =================================================================
func (s *inviteService) authorizedInvite(ctx context.Context, userID int64, role enum.UserRole, groupID, inviteID int64) (*entity.LinkInvite, error) {
	group, _, err := s.groups.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if err := requireInviteManager(group, userID, role); err != nil {
		return nil, err
	}
	return s.invites.FindLinkByID(ctx, groupID, inviteID)
}

func validateInviteParams(expiresAt *time.Time, maxUses *int) error {
	if maxUses != nil && *maxUses <= 0 {
		return exception.ErrInviteInvalidMaxUses
	}
	if expiresAt != nil && !expiresAt.After(time.Now()) {
		return exception.ErrInviteInvalidExpiration
	}
	return nil
}

func requireInviteManager(group *entity.Group, userID int64, role enum.UserRole) error {
	if role != enum.UserRoleAdmin && group.CreatedByUserID != userID {
		return exception.ErrGroupOwnerRequired
	}
	return nil
}

// =================================================================
// RESPONSE BUILDERS & MAPPERS
// =================================================================

func mapDirectInvites(invites []entity.DirectInvite) []dto.DirectInviteResponse {
	result := make([]dto.DirectInviteResponse, 0, len(invites))
	for i := range invites {
		result = append(result, *toDirectInviteResponse(&invites[i]))
	}
	return result
}

func toDirectInviteResponse(invite *entity.DirectInvite) *dto.DirectInviteResponse {
	response := &dto.DirectInviteResponse{ID: invite.ID, Status: string(invite.Status), CreatedAt: invite.CreatedAt, RespondedAt: invite.RespondedAt}
	if invite.Group != nil {
		response.Group = &dto.DirectInviteGroup{ID: invite.Group.ID, Title: invite.Group.Title}
	}
	if invite.Sender != nil {
		response.Sender = directInviteUser(invite.Sender)
	}
	if invite.Recipient != nil {
		response.Recipient = directInviteUser(invite.Recipient)
	}
	return response
}

func directInviteUser(user *entity.User) *dto.DirectInviteUser {
	return &dto.DirectInviteUser{ID: user.ID, Name: user.Name, AvatarAttachmentID: user.AvatarAttachmentID}
}

func buildInviteJoinGroupResponse(ctx context.Context, groups domainrepo.GroupRepository, group *entity.Group, count int64) (*dto.GroupResponse, error) {
	result := toGroupResponse(group, count)
	members, err := groups.ListActiveMembers(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	result.Members = toGroupMemberResponses(members, group.CreatedByUserID)
	return &result, nil
}

func (s *inviteService) groupInviteResponses(ctx context.Context, groupID int64) ([]dto.LinkInviteResponse, error) {
	invites, err := s.invites.FindLinksByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	responses := make([]dto.LinkInviteResponse, 0, len(invites))
	for i := range invites {
		response := toLinkInviteResponse(&invites[i])
		if invites[i].TokenCiphertext != "" {
			token, err := s.decryptInviteToken(invites[i].TokenCiphertext)
			if err != nil {
				return nil, exception.Wrap(exception.ErrInternal, err)
			}
			response.Token = token
			response.InviteURL = "/api/v1/invites/links/" + token + "/join"
		}
		responses = append(responses, response)
	}
	return responses, nil
}

func toLinkInviteResponse(invite *entity.LinkInvite) dto.LinkInviteResponse {
	return dto.LinkInviteResponse{
		ID: invite.ID, GroupID: invite.GroupID, ExpiresAt: invite.ExpiresAt, MaxUses: invite.MaxUses,
		UsedCount: invite.UsedCount, RevokedAt: invite.RevokedAt, CreatedAt: invite.CreatedAt,
	}
}
