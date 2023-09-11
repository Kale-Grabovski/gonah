package repo

import (
	"context"

	"github.com/Kale-Grabovski/gonah/src/domain"
)

type UserRepo struct {
	db domain.DB
}

func NewUserRepository(db domain.DB) *UserRepo {
	return &UserRepo{db}
}

func (r *UserRepo) GetAll() (ret []domain.User, err error) {
	q := `SELECT id, login FROM users ORDER BY id`
	rows, err := r.db.Query(context.Background(), q)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var u domain.User
		err = rows.Scan(&u.Id, &u.Login)
		if err != nil {
			return
		}
		ret = append(ret, u)
	}
	return
}

func (r *UserRepo) Create(user *domain.User) (err error) {
	q := `INSERT INTO users (login) VALUES ($1) RETURNING id`
	err = r.db.QueryRow(context.Background(), q, user.Login).Scan(&user.Id)
	return
}

func (r *UserRepo) GetByLogin(login string) (qnt int, err error) {
	q := `SELECT count(*) FROM users WHERE login = $1`
	err = r.db.QueryRow(context.Background(), q, login).Scan(&qnt)
	return
}

func (r *UserRepo) GetById(id int) (user domain.User, err error) {
	q := `SELECT id, login FROM users WHERE id = $1`
	err = r.db.QueryRow(context.Background(), q, id).Scan(&user.Id, &user.Login)
	return
}

func (r *UserRepo) Delete(id int) error {
	q := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(context.Background(), q, id)
	return err
}
