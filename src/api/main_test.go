package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"text/template"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/Kale-Grabovski/gonah/src/domain"
)

// http.Client wrapper for adding new methods, particularly sendJsonReq
type httpClient struct {
	parent http.Client
}

func TestMain(m *testing.M) {
	logger, err := domain.NewLogger()
	if err != nil {
		panic("cannot init logger: " + err.Error())
	}

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		logger.Panic("Could not construct pool", zap.Error(err))
	}

	err = pool.Client.Ping()
	if err != nil {
		logger.Panic("Could not connect to Docker", zap.Error(err))
	}

	var confPath string
	var stopDB func() // cannot use defer because of os.Exit
	confPath, stopDB = startPostgreSQL(pool, logger)
	startKafka(pool, logger)

	// We should change directory, otherwise the service will not find `migrations` directory
	err = os.Chdir("../..")
	if err != nil {
		stopDB()
		logger.Panic("os.Chdir failed", zap.Error(err))
	}

	cmd := exec.Command("./gonah", "api", "--config", confPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		stopDB()
		logger.Panic("cmd.Start failed", zap.Error(err))
	}

	// We have to make sure the migration is finished and REST API is available before running any tests.
	// Otherwise, there might be a race condition - the test see that API is unavailable and terminates,
	// pruning Docker container in the process which was running a migration.
	attempt := 0
	ok := false
	client := httpClient{}
	for attempt < 20 {
		attempt++
		_, _, err = client.sendJsonReq("GET", "http://localhost:8877/api/v1/users", []byte{})
		if err != nil {
			logger.Error("client.sendJsonReq failed: %v, waiting...", zap.Error(err))
			time.Sleep(1 * time.Second)
			continue
		}

		ok = true
		break
	}

	if !ok {
		stopDB()
		_ = cmd.Process.Kill()
		logger.Panic("REST API is unavailable")
		return
	}

	// Run all tests
	code := m.Run()

	_ = cmd.Process.Signal(syscall.SIGTERM)
	stopDB()
	os.Exit(code)
}

func startPostgreSQL(pool *dockertest.Pool, logger domain.Logger) (confPath string, cleaner func()) {
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15",
		Env: []string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_USER=usr",
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

	connString := "postgres://usr:secret@" + resource.GetHostPort("5432/tcp") + "/dbname?sslmode=disable"
	return waitForDBMS(pool, resource, connString, logger)
}

func startKafka(pool *dockertest.Pool, logger domain.Logger) {
	_, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "bashj79/kafka-kraft",
		Hostname:   "kafka",
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		logger.Error("could not start kafka", zap.Error(err))
		return
	}

	retryFn := func() error {
		conn, err := kafka.DialLeader(context.Background(), "tcp", "kafka:9092", "shit", 0)
		if err != nil {
			return err
		}
		defer conn.Close()

		message := kafka.Message{Value: []byte("Hello World")}
		_, err = conn.WriteMessages(message)
		return err
	}

	if err = pool.Retry(retryFn); err != nil {
		logger.Error("could not connect to kafka", zap.Error(err))
	}
}

func waitForDBMS(
	pool *dockertest.Pool,
	resource *dockertest.Resource,
	connString string,
	logger domain.Logger,
) (confPath string, cleaner func()) {
	// DBMS needs some time to start.
	// Port forwarding always works, thus net.Dial can't be used here.
	attempt := 0
	ok := false
	for attempt < 20 {
		attempt++
		conn, err := pgx.Connect(context.Background(), connString)
		if err != nil {
			logger.Info("pgx.Connect failed", zap.Error(err))
			time.Sleep(1 * time.Second)
			continue
		}

		_ = conn.Close(context.Background())
		ok = true
		break
	}

	if !ok {
		_ = pool.Purge(resource)
		logger.Panic("couldn't connect to PostgreSQL")
	}

	tmpl, err := template.New("config").Parse(`
loglevel: debug
listen: 8877
db:
  dsn: {{.ConnString}}
kafka:
  host: kafka:9092
`)
	if err != nil {
		_ = pool.Purge(resource)
		logger.Panic("template.Parse failed", zap.Error(err))
	}

	configArgs := struct {
		ConnString string
	}{
		ConnString: connString,
	}
	var configBuff bytes.Buffer
	err = tmpl.Execute(&configBuff, configArgs)
	if err != nil {
		_ = pool.Purge(resource)
		logger.Panic("tmpl.Execute failed", zap.Error(err))
	}

	confFile, err := os.CreateTemp("", "config.*.yaml")
	if err != nil {
		_ = pool.Purge(resource)
		logger.Panic("ioutil.TempFile failed", zap.Error(err))
	}

	_, err = confFile.WriteString(configBuff.String())
	if err != nil {
		_ = pool.Purge(resource)
		logger.Panic("confFile.WriteString failed", zap.Error(err))
	}

	err = confFile.Close()
	if err != nil {
		_ = pool.Purge(resource)
		logger.Panic("confFile.Close failed", zap.Error(err))
	}

	cleanerFunc := func() {
		// purge the container
		err := pool.Purge(resource)
		if err != nil {
			logger.Panic("pool.Purge failed", zap.Error(err))
		}

		err = os.Remove(confFile.Name())
		if err != nil {
			logger.Panic("os.Remove failed", zap.Error(err))
		}
	}

	return confFile.Name(), cleanerFunc
}

func (cl *httpClient) sendJsonReq(method, url string, reqBody []byte) (resp *http.Response, resBody []byte, err error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err = cl.parent.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	resBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return resp, resBody, nil
}
