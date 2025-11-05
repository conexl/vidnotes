package models

type CreateUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	Name      string `json:"name" binding:"required"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type UpdateUserRequest struct {
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ChangeSubscriptionRequest struct {
	Subscription string `json:"subscription" binding:"required,oneof=free premium business"`
}
