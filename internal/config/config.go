package config

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var (
	defaultMaxLogins           = 10
	defaultMaxPasswords        = 10
	defaultMaxIPs              = 10
	defaultCacheSize           = 10
	defaultCacheTTL            = time.Minute
	defaultBucketTTL           = time.Minute * 5
	defaultLogPreset           = "prod"
	defaultLogLevel            = "info"
	defaultLogEncoding         = "json"
	defaultLogOutputPaths      = []string{"stderr"}
	defaultLogErrorOutputPaths = []string{"stderr"}
	defaultDBSSLMode           = "require"
	defaultDBTimeout           = time.Second * 3
	defaultAPIServerHost       = "localhost"
	defaultAPIServerPort       = 8080
)

type Config struct {
	Main      *MainConf
	Logger    *LoggerConf
	DB        *DBConf
	APIServer *ServerConf
}

type MainConf struct {
	MaxLogins    int
	MaxPasswords int
	MaxIPs       int
	CacheSize    int
	CacheTTL     time.Duration
	BucketTTL    time.Duration
}

type LoggerConf struct {
	Preset           string
	Level            string
	Encoding         string
	OutputPaths      []string
	ErrorOutputPaths []string
}

type DBConf struct {
	Host     string
	Port     int
	DBName   string
	User     string
	Password string
	SSLMode  string
	Timeout  time.Duration
}

type ServerConf struct {
	Host string
	Port int
}

func New(filePath string, envPrefix string) (*Config, error) {
	config := Config{}
	v := viper.New()

	dir, name := path.Split(filePath)
	fParts := strings.SplitN(name, ".", 2)

	v.SetConfigName(fParts[0])
	v.SetConfigType(fParts[1])
	v.AddConfigPath(dir)
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()

	config.Main = processMainConf(v)

	config.Logger, err = processLoggerConf(v)
	if err != nil {
		return nil, err
	}

	config.DB, err = processDBConf(v)
	if err != nil {
		return nil, err
	}

	config.APIServer = processAPIServerConf(v)

	return &config, nil
}

func inStrArray(needle string, arr []string) bool {
	for _, v := range arr {
		if needle == v {
			return true
		}
	}
	return false
}

func getAllowedStringVal(v *viper.Viper, field string, allowed []string) (string, error) {
	value := v.GetString(field)
	if !inStrArray(value, allowed) {
		return "", fmt.Errorf(`invalid %s value: "%s", allowed values are %v`, field, value, allowed)
	}
	return value, nil
}

func processMainConf(v *viper.Viper) *MainConf {
	v.SetDefault("main.maxLogins", defaultMaxLogins)
	v.SetDefault("main.maxPasswords", defaultMaxPasswords)
	v.SetDefault("main.maxIPs", defaultMaxIPs)
	v.SetDefault("main.cacheSize", defaultCacheSize)
	v.SetDefault("main.cacheTTL", defaultCacheTTL)
	v.SetDefault("main.bucketTTL", defaultBucketTTL)

	conf := MainConf{}
	conf.MaxLogins = v.GetInt("main.maxLogins")
	conf.MaxPasswords = v.GetInt("main.maxPasswords")
	conf.MaxIPs = v.GetInt("main.maxIPs")
	conf.CacheSize = v.GetInt("main.cacheSize")
	conf.CacheTTL = v.GetDuration("main.cacheTTL")
	conf.BucketTTL = v.GetDuration("main.bucketTTL")

	return &conf
}

func processLoggerConf(v *viper.Viper) (*LoggerConf, error) {
	v.SetDefault("logger.preset", defaultLogPreset)
	v.SetDefault("logger.level", defaultLogLevel)
	v.SetDefault("logger.encoding", defaultLogEncoding)
	v.SetDefault("logger.outputPaths", defaultLogOutputPaths)
	v.SetDefault("logger.errorOutputPaths", defaultLogErrorOutputPaths)

	conf := LoggerConf{}

	val, err := getAllowedStringVal(v, "logger.preset", []string{"dev", "prod"})
	if err != nil {
		return nil, err
	}
	conf.Preset = val

	val, err = getAllowedStringVal(
		v,
		"logger.level",
		[]string{"debug", "info", "warn", "error", "dpanic", "panic", "fatal"},
	)
	if err != nil {
		return nil, err
	}
	conf.Level = val

	val, err = getAllowedStringVal(v, "logger.encoding", []string{"console", "json"})
	if err != nil {
		return nil, err
	}
	conf.Encoding = val

	conf.OutputPaths = v.GetStringSlice("logger.outputPaths")
	conf.ErrorOutputPaths = v.GetStringSlice("logger.errorOutputPaths")

	return &conf, nil
}

func processDBConf(v *viper.Viper) (*DBConf, error) {
	v.SetDefault("DBSSLMode", defaultDBSSLMode)
	v.SetDefault("DBTimeout", defaultDBTimeout)

	conf := DBConf{}

	if !v.IsSet("DBHost") || !v.IsSet("DBName") || !v.IsSet("DBUser") || !v.IsSet("DBPassword") {
		return nil, fmt.Errorf("not all database requisites are set, need set dbhost, dbname, dbuser, dbpassword")
	}

	conf.Host = v.GetString("DBHost")
	conf.Port = v.GetInt("DBPort")
	conf.DBName = v.GetString("DBName")
	conf.User = v.GetString("DBUser")
	conf.Password = v.GetString("DBPassword")
	conf.SSLMode = v.GetString("DBSSLMode")
	conf.Timeout = v.GetDuration("DBTimeout")

	return &conf, nil
}

func processAPIServerConf(v *viper.Viper) *ServerConf {
	v.SetDefault("api-server.host", defaultAPIServerHost)
	v.SetDefault("api-server.port", defaultAPIServerPort)

	conf := ServerConf{}
	conf.Host = v.GetString("api-server.host")
	conf.Port = v.GetInt("api-server.port")

	return &conf
}
