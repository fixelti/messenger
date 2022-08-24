package user

type Repository interface {
	Create(User) (User, error)
	Read(uint) (User, error)
	ReadByLogin(string) (User, error)
	List(Filter) (Pagination, error)
	Update(User) (User, error)
	Delete(uint) error
}
