package logger

import (
	"os"
	"strconv"
	"time"
)

// EnvironmentConfig cria configurações baseadas em variáveis de ambiente
func EnvironmentConfig() *Config {
	config := DefaultConfig()

	// LOG_LEVEL
	if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
		switch levelStr {
		case "debug", "DEBUG":
			config.Level = DebugLevel
		case "info", "INFO":
			config.Level = InfoLevel
		case "warn", "WARN", "warning", "WARNING":
			config.Level = WarnLevel
		case "error", "ERROR":
			config.Level = ErrorLevel
		case "fatal", "FATAL":
			config.Level = FatalLevel
		case "panic", "PANIC":
			config.Level = PanicLevel
		}
	}

	// LOG_FORMAT
	if formatStr := os.Getenv("LOG_FORMAT"); formatStr != "" {
		switch formatStr {
		case "json", "JSON":
			config.Format = JSONFormat
		case "console", "CONSOLE":
			config.Format = ConsoleFormat
		case "text", "TEXT":
			config.Format = TextFormat
		}
	}

	// SERVICE_NAME
	if serviceName := os.Getenv("SERVICE_NAME"); serviceName != "" {
		config.ServiceName = serviceName
	}

	// SERVICE_VERSION
	if serviceVersion := os.Getenv("SERVICE_VERSION"); serviceVersion != "" {
		config.ServiceVersion = serviceVersion
	}

	// ENVIRONMENT
	if environment := os.Getenv("ENVIRONMENT"); environment != "" {
		config.Environment = environment
	} else if env := os.Getenv("ENV"); env != "" {
		config.Environment = env
	}

	// LOG_ADD_SOURCE
	if addSourceStr := os.Getenv("LOG_ADD_SOURCE"); addSourceStr != "" {
		if addSource, err := strconv.ParseBool(addSourceStr); err == nil {
			config.AddSource = addSource
		}
	}

	// LOG_ADD_STACKTRACE
	if addStackStr := os.Getenv("LOG_ADD_STACKTRACE"); addStackStr != "" {
		if addStack, err := strconv.ParseBool(addStackStr); err == nil {
			config.AddStacktrace = addStack
		}
	}

	// LOG_TIME_FORMAT
	if timeFormat := os.Getenv("LOG_TIME_FORMAT"); timeFormat != "" {
		config.TimeFormat = timeFormat
	}

	// Configuração de sampling
	if samplingInitial := os.Getenv("LOG_SAMPLING_INITIAL"); samplingInitial != "" {
		if initial, err := strconv.Atoi(samplingInitial); err == nil {
			if config.SamplingConfig == nil {
				config.SamplingConfig = &SamplingConfig{}
			}
			config.SamplingConfig.Initial = initial
		}
	}

	if samplingThereafter := os.Getenv("LOG_SAMPLING_THEREAFTER"); samplingThereafter != "" {
		if thereafter, err := strconv.Atoi(samplingThereafter); err == nil {
			if config.SamplingConfig == nil {
				config.SamplingConfig = &SamplingConfig{}
			}
			config.SamplingConfig.Thereafter = thereafter
		}
	}

	if samplingTick := os.Getenv("LOG_SAMPLING_TICK"); samplingTick != "" {
		if tick, err := time.ParseDuration(samplingTick); err == nil {
			if config.SamplingConfig == nil {
				config.SamplingConfig = &SamplingConfig{}
			}
			config.SamplingConfig.Tick = tick
		}
	}

	return config
}

// DevelopmentConfig retorna uma configuração otimizada para desenvolvimento
func DevelopmentConfig() *Config {
	return &Config{
		Level:         InfoLevel,
		Format:        JSONFormat,
		Output:        os.Stdout,
		TimeFormat:    time.RFC3339,
		AddSource:     false,
		AddStacktrace: false,
		Fields:        make(map[string]any),
	}
}

// StatementConfig retorna uma configuração otimizada para homologação
func StatementConfig() *Config {
	return &Config{
		Level:         InfoLevel,
		Format:        JSONFormat,
		Output:        os.Stdout,
		TimeFormat:    time.RFC3339,
		AddSource:     false,
		AddStacktrace: false,
		Fields:        make(map[string]any),
		SamplingConfig: &SamplingConfig{
			Initial:    1000,
			Thereafter: 100,
			Tick:       time.Second,
		},
	}
}

// ProductionConfig retorna uma configuração otimizada para produção
func ProductionConfig() *Config {
	return &Config{
		Level:         WarnLevel,
		Format:        JSONFormat,
		Output:        os.Stdout,
		TimeFormat:    time.RFC3339,
		AddSource:     false,
		AddStacktrace: false,
		Fields:        make(map[string]any),
		SamplingConfig: &SamplingConfig{
			Initial:    1000,
			Thereafter: 100,
			Tick:       time.Second,
		},
	}
}

// TestingConfig retorna uma configuração otimizada para testes
func TestingConfig() *Config {
	return &Config{
		Level:         DebugLevel,
		Format:        JSONFormat,
		Output:        os.Stdout,
		TimeFormat:    time.RFC3339,
		AddSource:     true,
		AddStacktrace: true,
		Fields:        make(map[string]any),
	}
}

// ConfigFromEnvironment cria e configura automaticamente baseado no ambiente
func ConfigFromEnvironment() error {
	config := EnvironmentConfig()

	// Detecta automaticamente o melhor provider
	provider := "zap" // padrão

	if providerEnv := os.Getenv("LOG_PROVIDER"); providerEnv != "" {
		provider = providerEnv
	} else {
		// Auto-detecta baseado no ambiente
		switch config.Environment {
		case "production":
			provider = "zap" // Zap é mais performático para produção
		case "homolog":
			provider = "zerolog" // Zerolog tem boa DX para desenvolvimento
		case "development":
			provider = "slog" // Zerolog tem boa DX para desenvolvimento
		default:
			provider = "zap" // padrão
		}
		// Caso contrário, usa zap como padrão
	}

	return SetProvider(provider, config)
}

// MustConfigFromEnvironment similar ao ConfigFromEnvironment mas entra em pânico em caso de erro
func MustConfigFromEnvironment() {
	if err := ConfigFromEnvironment(); err != nil {
		panic("Failed to configure logging from environment: " + err.Error())
	}
}
