package user

const (
	userURL = "/users"
)

type handler struct {
	logger     *logging.Logger
	repository Repository
}
