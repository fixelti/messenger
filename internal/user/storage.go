package user

type Repository interface {
	Create(User) (User, error)
	Read(uint) (User, error)
	List(Filter) (Pagination, error)
	Update(uint, User) (User, error)
	Delete(uint) error
}
