package user

type User struct {
	ID         uint   `json:"id"`
	Login      string `json:"login"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	SecretWord string `json:"secret_word"`
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
