package user

import "time"

type User struct {
	ID         uint `json:"id"`
	CreatedAt  time.Time
	DeletedAt  *time.Time `sql:"index"`
	Email      string     `json:"email"`
	Login      string     `json:"login"`
	Password   string     `json:"password"`
	SecretWord string     `json:"secret_word"`
	FindVision bool       `json:"find_vision"`
	AddFriend  bool       `json:"add_friend"`
	Friends    []uint     `json:"friends"`
	UserRole   uint       `json:"user_role"`
}

// TODO: возможно надо будет отредактировать эту структуру

type CreateUserDTO struct {
	Login      string `json:"login" binding:"required"`
	Email      string `json:"email" binding:"required"`
	Password   string `json:"password" binding:"required"`
	SecretWord string `json:"secret_word" binding:"required"`
}

type Filter struct {
	PageID   int64 `json:"page_id"`
	PageSize int64 `json:"page_size"`
}

type Pagination struct {
	PageID       int64         `json:"page_id"`
	PageSize     int64         `json:"page_size"`
	TotalRecords int64         `json:"total_records"`
	TotalCount   int64         `json:"total_count"`
	Records      []interface{} `json:"records"`
}
