package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/rs/zerolog/log"
	"os"
	"sync"
	"time"
)

type Config struct {
	ServerPort string
	Timezone   string
}

type Secrets struct {
	RequestToken          string `json:"REQUEST_TOKEN"`
	ArgoAdminUserName     string `json:"ARGO_ADMIN_USERNAME"`
	ProdArgoAdminPassword string `json:"PROD_ARGO_ADMIN_PASSWORD"`
	DevArgoAdminPassword  string `json:"DEV_ARGO_ADMIN_PASSWORD"`
}

type SecretLoader struct {
	mu         sync.RWMutex
	region     string
	secretName string
	client     *secretsmanager.Client
	secrets    *Secrets
	config     *Config
	loaded     bool
}

var (
	secretLoaderInstance *SecretLoader
	once                 sync.Once
	defaultPort          = "8080"
)

func getSecretLoader(region, secretName string) (*SecretLoader, error) {
	log.Debug().Msg("=====> Getting Secret Loader")
	var initErr error
	once.Do(func() {
		cfg, err := awsconfig.LoadDefaultConfig(
			context.Background(),
			awsconfig.WithRegion(region),
		)

		if err != nil {
			initErr = fmt.Errorf("failed to load SDK config: %w", err)
			return
		}

		// singleton 인스턴스 생성
		secretLoaderInstance = &SecretLoader{
			region:     region,
			secretName: secretName,
			client:     secretsmanager.NewFromConfig(cfg),
			secrets:    &Secrets{},
			config: &Config{
				ServerPort: defaultPort,
				Timezone:   "Asia/Seoul",
			},
			loaded: false,
		}
	})

	if initErr != nil {
		return nil, initErr
	}

	return secretLoaderInstance, nil
}

func (sl *SecretLoader) loadSecrets() error {
	log.Debug().Msg("=====> Loading Secrets")
	sl.mu.Lock()
	defer sl.mu.Unlock()

	if sl.loaded {
		return nil
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(sl.secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}

	output, err := sl.client.GetSecretValue(context.Background(), input)
	if err != nil {
		return fmt.Errorf("LoadSecrets | failed to get secret value: %w", err)
	}

	if output.SecretString == nil {
		return errors.New("LoadSecrets | loaded secrets are empty")
	}

	if err := json.Unmarshal([]byte(*output.SecretString), &sl.secrets); err != nil {
		return fmt.Errorf("LoadSecrets | failed to unmarshal secret value: %w", err)
	}

	if port := os.Getenv("SERVER_PORT"); port != "" {
		sl.config.ServerPort = port
	}

	// 환경변수에서 타임존 설정 가져오기 (설정되어 있지 않으면 기본값 사용)
	if tz := os.Getenv("TIMEZONE"); tz != "" {
		sl.config.Timezone = tz
	}

	sl.loaded = true
	return nil
}

func (sl *SecretLoader) setEnvironmentVariables() error {
	log.Debug().Msg("=====> Setting Environment Variables")
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	if !sl.loaded {
		return errors.New("SetEnvironmentVariables | secret does not exist")
	}

	serviceEnv := os.Getenv("APP_ENV")
	var adminPassword string

	switch serviceEnv {
	case "dev":
		adminPassword = sl.secrets.DevArgoAdminPassword
	case "prod":
		adminPassword = sl.secrets.ProdArgoAdminPassword
	}

	envVars := map[string]string{
		"REQUEST_TOKEN":       sl.secrets.RequestToken,
		"ARGO_ADMIN_USERNAME": sl.secrets.ArgoAdminUserName,
		"ARGO_ADMIN_PASSWORD": adminPassword,
	}

	for k, v := range envVars {
		if err := os.Setenv(k, v); err != nil {
			return fmt.Errorf("SetEnvironmentVariables | failed to set environment variable %s: %w", k, err)
		}
	}
	return nil
}

func (sl *SecretLoader) setTimezone() error {
	log.Debug().Msg("=====> Setting Timezone")
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	if !sl.loaded {
		return errors.New("setTimezone | config not loaded yet")
	}

	location, err := time.LoadLocation(sl.config.Timezone)
	if err != nil {
		return fmt.Errorf("setTimezone | failed to load timezone %s: %w", sl.config.Timezone, err)
	}

	time.Local = location
	return nil
}

func (sl *SecretLoader) GetConfig() (*Config, error) {
	log.Debug().Msg("=====> Getting Config")
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	if !sl.loaded {
		return nil, errors.New("GetConfig | secret does not exist")
	}

	return sl.config, nil
}

func Setting() *Config {
	log.Debug().Msg("=====> Initialize Secret Loader")
	// Instance 생성
	loader, err := getSecretLoader("ap-northeast-2", "/secret/devops")
	if err != nil {
		log.Fatal().Err(err).Msg("Setting | Failed to get secret loader")
		return nil
	}

	// Secret 로드
	if err := loader.loadSecrets(); err != nil {
		log.Fatal().Err(err).Msg("Setting | Failed to load secrets")
		return nil
	}

	// 환경 변수 설정
	if err := loader.setEnvironmentVariables(); err != nil {
		log.Fatal().Err(err).Msg("Setting | Failed to set environment variables")
		return nil
	}

	// Timezone 설정
	if err := loader.setTimezone(); err != nil {
		log.Fatal().Err(err).Msg("Setting | Failed to set timezone")
	}

	// 설정 반환
	config, err := loader.GetConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Setting | Failed to get config")
		return nil
	}

	return config
}
