package domain

type User struct {
	Id    int    `json:"id"`
	Login string `json:"login" validate:"required"`
}

func (u *User) getId() int {
	return 1
}
