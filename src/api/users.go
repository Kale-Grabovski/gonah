package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/Kale-Grabovski/gonah/src/domain"
	"github.com/Kale-Grabovski/gonah/src/repo"
)

type UsersAction struct {
	userRepo *repo.UserRepo
	logger   domain.Logger
}

func NewUsersAction(
	userRepo *repo.UserRepo,
	logger domain.Logger,
) *UsersAction {
	return &UsersAction{userRepo, logger}
}

func (s *UsersAction) GetAll(c echo.Context) (err error) {
	users, err := s.userRepo.GetAll()
	if err != nil {
		s.logger.Error("cannot get users", zap.Error(err))
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}
	return c.JSON(http.StatusOK, users)
}

func (s *UsersAction) GetById(c echo.Context) (err error) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		s.logger.Error("wrong user ID", zap.Error(err))
		return c.String(http.StatusBadRequest, "wrong user ID")
	}

	user, err := s.userRepo.GetById(id)
	if errors.Is(err, domain.ErrNoRows) {
		return c.String(http.StatusNotFound, "user not found")
	} else if err != nil {
		s.logger.Error("cannot get user by ID", zap.Error(err))
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}
	return c.JSON(http.StatusOK, user)
}

func (s *UsersAction) Create(c echo.Context) (err error) {
	user := &domain.User{}

	if err = c.Bind(user); err != nil {
		return c.String(http.StatusBadRequest, "login required")
	}
	if err = c.Validate(user); err != nil {
		return err
	}

	q, err := s.userRepo.GetByLogin(user.Login)
	if err != nil {
		s.logger.Error("cannot check if user exists", zap.Error(err))
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}
	if q > 0 {
		return c.String(http.StatusBadRequest, "user with such login already exists")
	}

	err = s.userRepo.Create(user)
	if err != nil {
		s.logger.Error("cannot create user", zap.Error(err))
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}
	return c.JSON(http.StatusOK, user)
}

func (s *UsersAction) Delete(c echo.Context) (err error) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		s.logger.Error("wrong user ID", zap.Error(err))
		return c.String(http.StatusBadRequest, "wrong user ID")
	}

	_, err = s.userRepo.GetById(id)
	if errors.Is(err, domain.ErrNoRows) {
		return c.String(http.StatusNotFound, "user not found")
	} else if err != nil {
		s.logger.Error("cannot get user by ID", zap.Error(err))
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}

	err = s.userRepo.Delete(id)
	if err != nil {
		s.logger.Error("cannot delete user", zap.Error(err))
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}
	return c.JSON(http.StatusOK, "OK")
}
