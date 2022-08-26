package user

type Repository interface {
	Create(User) (User, error)
	Read(uint) (User, error)
	FindByLogin(string, uint) ([]*User, error)
	List(Filter) (Pagination, error)
	Update(User) (User, error)
	Delete(uint) error
}
