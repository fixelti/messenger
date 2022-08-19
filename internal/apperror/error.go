package apperror

import "encoding/json"

var (
	ErrNotFound = NewAppError(nil, "not found", "", "US-000001")
	ErrNotAuth  = NewAppError(nil, "not authorization", "", "US-000002")
)

type AppError struct {
	Err              error  `json:"-"`
	Message          string `json:"message"`
	DeveloperMessage string `json:"developer_message"`
	Code             string `json:"code"`
}

func NewAppError(err error, message, developerMessage, code string) *AppError {
	return &AppError{
		Err:              err,
		Message:          message,
		DeveloperMessage: developerMessage,
		Code:             code,
	}
}

func (a *AppError) Error() string {
	return a.Message
}

func (a *AppError) Unwrap() error {
	return a.Err
}

func (a *AppError) Marshal() []byte {
	marshal, err := json.Marshal(a)
	if err != nil {
		return nil
	}
	return marshal
}

func (a *AppError) systemError(err error) *AppError {
	return NewAppError(err, "internal system error", err.Error(), "SE-000001")
}
