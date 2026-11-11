package dto

type UserPreviewResponse struct {
	ID                 int64   `json:"id"`
	Name               string  `json:"name"`
	AvatarAttachmentID *int64  `json:"avatar_attachment_id"`
	AvatarURL          *string `json:"avatar_url"`
}
