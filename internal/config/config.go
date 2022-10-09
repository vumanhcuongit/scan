package config

import (
	"io/ioutil"
	"os"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type App struct {
	EnvConfig    *EnvConfig         `yaml:"common"`
	DB           *DatabaseConfig    `yaml:"db"`
	MessageQueue MessageQueueConfig `yaml:"message_queue"`
	GRPCPort     string             `yaml:"grpc_port"`
	HTTPAddr     string             `yaml:"http_addr"`
}

type EnvConfig struct {
	Environment string `yaml:"environment"`
	SentryDSN   string `yaml:"sentry_dsn"`
}

type MessageQueueConfig struct {
	Broker       string `yaml:"broker"`
	TopicRequest string `yaml:"topic_request"`
	TopicReply   string `yaml:"topic_reply"`
	GroupID      string `yaml:"group_id"`
}

type DatabaseConfig struct {
	GlobalServiceID            uint16 `yaml:"global_service_id"`
	DriverName                 string `yaml:"driver_name"`
	DataSource                 string `yaml:"data_source"`
	MaxOpenConns               int    `yaml:"max_open_conns"`
	MaxIdleConns               int    `yaml:"max_idle_conns"`
	ConnMaxLifeTimeMiliseconds int64  `yaml:"conn_max_life_time_ms"`
	MigrationConnURL           string `yaml:"migration_conn_url"`
	IsDevMode                  bool   `yaml:"is_dev_mode"`
}

type WebhookConfig struct {
	TimeOutInSeconds int `yaml:"timeout_in_seconds"`
}

type TNSLServiceConfig struct {
	BaseURL                    string `yaml:"base_url"`
	TimeOutInSeconds           int    `yaml:"timeout_in_seconds"`
	AccessToken                string `yaml:"access_token"`
	WebhookSecretToken         string `yaml:"webhook_secret_token"`
	SwitchingStatusAccessToken string `yaml:"switching_status_access_token"`
}

type ServiceDiscovery struct {
	CredentialDSN string `yaml:"credential_dsn"`
	OrderDSN      string `yaml:"order_dsn"`
}

// Load load config from file and environment variables.
func Load(filePath string) (*App, error) {
	if filePath == "" {
		filePath = os.Getenv("CONFIG_FILE")
	}

	fields := []interface{}{
		"func", "config.readFromFile",
		"filePath", filePath,
	}

	zapLogger, _ := zap.NewProduction()
	defer func() {
		_ = zapLogger.Sync()
	}()
	log := zapLogger.Sugar()

	log.With(fields...).Debug("Load config")

	log.Debugf("CONFIG_FILE=%v", filePath)

	configBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.With(fields...).Error("Failed to load config file")
		return nil, err
	}

	configBytes = []byte(os.ExpandEnv(string(configBytes)))
	cfg := &App{}
	err = yaml.Unmarshal(configBytes, cfg)
	if err != nil {
		log.With(fields...).Error("Failed to parse config file")
		return nil, err
	}

	return cfg, nil
}

func (c *EnvConfig) IsLocalEnvironment() bool {
	return c.Environment == "local"
}

// IsProductionEnvironment ...
func (c *EnvConfig) IsProductionEnvironment() bool {
	return c.Environment == "production"
}
