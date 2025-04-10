// Package config пакет с конфигом приложения.
package config

// Config структура конфиг файла.
type Config struct {
	Env      string         `yaml:"env"      env-default:"local"`
	RPC      RPCConfig      `yaml:"rpc"      env-required:"true"`
	Database DatabaseConfig `yaml:"database" env-required:"true"`
	Logging  LoggingConfig  `yaml:"logging"  env-required:"false"`
	OTP      OTPConfig      `yaml:"otp"`
	Security SecurityConfig `yaml:"security"`
}

// RPCConfig структура конфига для RPC сервера.
type RPCConfig struct {
	Address string `yaml:"address" env:"GK_GRPC_ADDR" env-default:":50051"`
}

// DatabaseConfig структура конфига для базы данных.
type DatabaseConfig struct {
	URI                 string `yaml:"uri"                   env:"GK_DATABASE_URI"             env-required:"true"`
	ExternalStoragePath string `yaml:"external_storage_path"                                   env-required:"true"`
	MigrationsPath      string `yaml:"migrations_path"       env:"GK_DATABASE_MIGRATIONS_PATH" env-required:"true"`
	MaxCons             int    `yaml:"max_cons"              env:"GK_DATABASE_MAX_CONS" env-default:"20"`
}

// LoggingConfig структура конфига для логгера.
type LoggingConfig struct {
	Level string `yaml:"level" env:"GK_LOGGING_LEVEL" env-default:"info"`
}

// SecurityConfig структура конфига параметров безопасности.
type SecurityConfig struct {
	Pepper string `yaml:"pepper" env-required:"true"`
}

// OTPConfig структура конфига для OTP паролей.
type OTPConfig struct {
	Algorithm string `yaml:"algorithm" env-default:"SHA1"`
	Digits    int    `yaml:"digits"    env-default:"6"`
	Period    int    `yaml:"period"    env-default:"30"`
}
