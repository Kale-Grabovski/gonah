package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/jackc/pgx/v4"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"go.uber.org/zap"

	"github.com/Kale-Grabovski/gonah/src/domain"
)

const containersExpireSec = 30

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
		logger.Panic("could not construct pool", zap.Error(err))
	}
	err = pool.Client.Ping()
	if err != nil {
		logger.Panic("could not connect to Docker", zap.Error(err))
	}
	pool.MaxWait = containersExpireSec * time.Second

	var (
		kafkaConn string
		dbConn    string
		wg        sync.WaitGroup
	)
	wg.Add(2)

	go func() {
		kafkaConn = startKafka(pool, logger)
		wg.Done()
	}()
	go func() {
		dbConn = startPostgreSQL(pool, logger)
		wg.Done()
	}()
	wg.Wait()

	startAPI(m, logger, dbConn, kafkaConn)
}

func startAPI(m *testing.M, logger domain.Logger, dbConn, kafkaConn string) {
	// We should change directory, otherwise the service will not find `migrations` directory
	err := os.Chdir("../..")
	if err != nil {
		logger.Panic("os.Chdir failed", zap.Error(err))
	}

	cmd := exec.Command("./gonah", "api")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, domain.EnvPrefix+"_APIPORT=8877")
	cmd.Env = append(cmd.Env, domain.EnvPrefix+"_DB_DSN="+dbConn)
	cmd.Env = append(cmd.Env, domain.EnvPrefix+"_KAFKA_HOST="+kafkaConn)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		logger.Panic("failed to start api", zap.Error(err))
	}
	waitAPI(cmd, logger)

	// Run all tests
	code := m.Run()
	_ = cmd.Process.Signal(syscall.SIGKILL)
	os.Exit(code)
}

// We have to make sure the migration is finished and REST API is available before running any tests.
// Otherwise, there might be a race condition - the test see that API is unavailable and terminates,
// pruning Docker container in the process which was running a migration.
func waitAPI(cmd *exec.Cmd, logger domain.Logger) {
	ok := false
	attempt := 0
	client := httpClient{}
	for attempt < 20 {
		attempt++
		_, _, err := client.sendJsonReq(http.MethodGet, "http://localhost:8877/up", []byte{})
		if err == nil {
			ok = true
			break
		}
		logger.Warn("client.sendJsonReq failed: %v, waiting...", zap.Error(err))
		time.Sleep(200 * time.Millisecond)
	}
	if !ok {
		_ = cmd.Process.Kill()
		logger.Panic("REST API is unavailable")
	}
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
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432/tcp": {{HostIP: "localhost", HostPort: "5435/tcp"}},
		},
		ExposedPorts: []string{"5432/tcp"},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		logger.Panic("could not start postgres", zap.Error(err))
	}
	resource.Expire(containersExpireSec)

	databaseUrl := "postgres://usr:secret@" + resource.GetHostPort("5432/tcp") + "/dbname?sslmode=disable"

	if err = pool.Retry(func() error {
		db, err := pgx.Connect(context.Background(), databaseUrl)
		if err != nil {
			return err
		}
		return db.Ping(context.Background())
	}); err != nil {
		logger.Panic("could not connect to postgres", zap.Error(err))
	}

	return databaseUrl
}

func startKafka(pool *dockertest.Pool, logger domain.Logger) (kafkaHost string) {
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "bitnami/kafka",
		Tag:        "3.5",
		Hostname:   "kafka",
		Env: []string{
			"KAFKA_CFG_NODE_ID=0",
			"KAFKA_CFG_PROCESS_ROLES=controller,broker",
			"KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=0@kafka:9093",
			"KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093",
			"KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://:9092",
			"KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT",
			"KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER",
			"KAFKA_CFG_INTER_BROKER_LISTENER_NAME=PLAINTEXT",
		},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"9092/tcp": {{HostIP: "localhost", HostPort: "9099/tcp"}},
		},
		ExposedPorts: []string{"9092/tcp"},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		logger.Panic("could not start kafka", zap.Error(err))
	}
	resource.Expire(containersExpireSec)
	kafkaHost = resource.GetHostPort("9092/tcp")

	if err = pool.Retry(func() error {
		p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kafkaHost})
		if err != nil {
			return err
		}
		topic := "shit"
		return p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte("shit"),
		}, nil)
	}); err != nil {
		logger.Panic("could not connect to kafka", zap.Error(err))
	}
	return
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
	return resp, resBody, err
}
