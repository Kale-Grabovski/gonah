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

var configTpl = `
loglevel: debug
listen: 8877
db:
  dsn: {{.ConnString}}
kafka:
  host: {{.KafkaHost}}
`

type configArgs struct {
	ConnString string
	KafkaHost  string
}

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
	pool.MaxWait = 120 * time.Second

	kafkaHost := startKafka(pool, logger)
	dbConnString := startPostgreSQL(pool, logger)
	confFilename := createConfig(dbConnString, kafkaHost, logger)
	startAPI(m, confFilename, logger)
}

func startAPI(m *testing.M, confFilename string, logger domain.Logger) {
	// We should change directory, otherwise the service will not find `migrations` directory
	err := os.Chdir("../..")
	if err != nil {
		logger.Panic("os.Chdir failed", zap.Error(err))
	}

	cmd := exec.Command("./gonah", "api", "--config", confFilename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
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
		_ = cmd.Process.Kill()
		logger.Panic("REST API is unavailable")
		return
	}

	// Run all tests
	code := m.Run()

	_ = cmd.Process.Signal(syscall.SIGTERM)
	os.Exit(code)
}

func startPostgreSQL(pool *dockertest.Pool, logger domain.Logger) string {
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
		logger.Panic("Could not start resource", zap.Error(err))
	}

	databaseUrl := "postgres://usr:secret@" + resource.GetHostPort("5432/tcp") + "/dbname?sslmode=disable"

	if err = pool.Retry(func() error {
		db, err := pgx.Connect(context.Background(), databaseUrl)
		if err != nil {
			return err
		}
		return db.Ping(context.Background())
	}); err != nil {
		logger.Panic("Could not connect to docker", zap.Error(err))
	}

	return databaseUrl
}

func startKafka(pool *dockertest.Pool, logger domain.Logger) (host string) {
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "bashj79/kafka-kraft",
		Hostname:   "kafka",
		PortBindings: map[docker.Port][]docker.PortBinding{
			"9092/tcp": {{HostIP: "localhost", HostPort: "9092/tcp"}},
		},
		ExposedPorts: []string{"9092/tcp"},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		logger.Error("could not start kafka", zap.Error(err))
		return
	}
	host = resource.GetHostPort("9092/tcp")

	if err = pool.Retry(func() error {
		conn, err := kafka.DialLeader(context.Background(), "tcp", host, "shit-topic", 0)
		if err != nil {
			return err
		}
		defer conn.Close()

		message := kafka.Message{Value: []byte("Hello World")}
		_, err = conn.WriteMessages(message)
		return err
	}); err != nil {
		logger.Error("could not connect to kafka", zap.Error(err))
	}
	return
}

func createConfig(dbConn, kafkaConn string, logger domain.Logger) string {
	tmpl, err := template.New("config").Parse(configTpl)
	if err != nil {
		logger.Panic("template.Parse failed", zap.Error(err))
	}

	args := configArgs{
		ConnString: dbConn,
		KafkaHost:  kafkaConn,
	}
	var configBuff bytes.Buffer
	err = tmpl.Execute(&configBuff, args)
	if err != nil {
		logger.Panic("tmpl.Execute failed", zap.Error(err))
	}

	confFile, err := os.CreateTemp("", "config.*.yaml")
	if err != nil {
		logger.Panic("ioutil.TempFile failed", zap.Error(err))
	}

	_, err = confFile.WriteString(configBuff.String())
	if err != nil {
		logger.Panic("confFile.WriteString failed", zap.Error(err))
	}

	err = confFile.Close()
	if err != nil {
		logger.Panic("confFile.Close failed", zap.Error(err))
	}
	return confFile.Name()
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
