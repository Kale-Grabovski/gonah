package domain

type User struct {
	Id    int
	Login string
}

func (u *User) getId() int {
	return 1
}
