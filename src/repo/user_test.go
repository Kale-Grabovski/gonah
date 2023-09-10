package repo

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"go.uber.org/zap"

	"github.com/Kale-Grabovski/gonah/src/domain"
	"github.com/Kale-Grabovski/gonah/src/service/migrate"
)

var db *pgxpool.Pool

func TestMain(m *testing.M) {
	logger, err := domain.NewLogger()

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		logger.Error("Could not construct pool", zap.Error(err))
		return
	}

	err = pool.Client.Ping()
	if err != nil {
		logger.Error("Could not connect to Docker", zap.Error(err))
		return
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15",
		Env: []string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_USER=user_name",
			"POSTGRES_DB=dbname",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		logger.Error("Could not start resource", zap.Error(err))
		return
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://user_name:secret@%s/dbname?sslmode=disable", hostAndPort)

	logger.Info("Connecting to database on url: " + databaseUrl)

	resource.Expire(60) // Tell docker to hard kill the container in 60 seconds

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 60 * time.Second
	if err = pool.Retry(func() error {
		db, err = pgxpool.Connect(context.Background(), databaseUrl)
		return err
	}); err != nil {
		logger.Error("Could not connect to docker", zap.Error(err))
		return
	}

	migrate.Run("../../migrations", db, logger)

	//Run tests
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err = pool.Purge(resource); err != nil {
		logger.Error("Could not purge resource", zap.Error(err))
		return
	}
	os.Exit(code)
}

func TestUser(t *testing.T) {
	rep := NewUserRepository(db)
	user, err := rep.Create("shit")
	if err != nil {
		t.Errorf("can't create user: %v", err)
	}
	if user.Login != "shit" {
		t.Errorf("wrong user: %v", err)
	}
}
