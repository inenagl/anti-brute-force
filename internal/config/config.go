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

	dir, name := path.Split(filePath)
	fParts := strings.SplitN(name, ".", 2)

	viper.SetConfigName(fParts[0])
	viper.SetConfigType(fParts[1])
	viper.AddConfigPath(dir)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()

	config.Main = processMainConf()

	config.Logger, err = processLoggerConf()
	if err != nil {
		return nil, err
	}

	config.DB, err = processDBConf()
	if err != nil {
		return nil, err
	}

	config.APIServer = processAPIServerConf()

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

func getAllowedStringVal(field string, allowed []string) (string, error) {
	value := viper.GetString(field)
	if !inStrArray(value, allowed) {
		return "", fmt.Errorf(`invalid %s value: "%s", allowed values are %v`, field, value, allowed)
	}
	return value, nil
}

func processMainConf() *MainConf {
	viper.SetDefault("main.maxLogins", defaultMaxLogins)
	viper.SetDefault("main.maxPasswords", defaultMaxPasswords)
	viper.SetDefault("main.maxIPs", defaultMaxIPs)
	viper.SetDefault("main.cacheSize", defaultCacheSize)
	viper.SetDefault("main.cacheTTL", defaultCacheTTL)
	viper.SetDefault("main.bucketTTL", defaultBucketTTL)

	conf := MainConf{}
	conf.MaxLogins = viper.GetInt("main.maxLogins")
	conf.MaxPasswords = viper.GetInt("main.maxPasswords")
	conf.MaxIPs = viper.GetInt("main.maxIPs")
	conf.CacheSize = viper.GetInt("main.cacheSize")
	conf.CacheTTL = viper.GetDuration("main.cacheTTL")
	conf.BucketTTL = viper.GetDuration("main.bucketTTL")

	return &conf
}

func processLoggerConf() (*LoggerConf, error) {
	viper.SetDefault("logger.preset", defaultLogPreset)
	viper.SetDefault("logger.level", defaultLogLevel)
	viper.SetDefault("logger.encoding", defaultLogEncoding)
	viper.SetDefault("logger.outputPaths", defaultLogOutputPaths)
	viper.SetDefault("logger.errorOutputPaths", defaultLogErrorOutputPaths)

	conf := LoggerConf{}

	val, err := getAllowedStringVal("logger.preset", []string{"dev", "prod"})
	if err != nil {
		return nil, err
	}
	conf.Preset = val

	val, err = getAllowedStringVal(
		"logger.level",
		[]string{"debug", "info", "warn", "error", "dpanic", "panic", "fatal"},
	)
	if err != nil {
		return nil, err
	}
	conf.Level = val

	val, err = getAllowedStringVal("logger.encoding", []string{"console", "json"})
	if err != nil {
		return nil, err
	}
	conf.Encoding = val

	conf.OutputPaths = viper.GetStringSlice("logger.outputPaths")
	conf.ErrorOutputPaths = viper.GetStringSlice("logger.errorOutputPaths")

	return &conf, nil
}

func processDBConf() (*DBConf, error) {
	viper.SetDefault("DBSSLMode", defaultDBSSLMode)
	viper.SetDefault("DBTimeout", defaultDBTimeout)

	conf := DBConf{}

	if !viper.IsSet("DBHost") || !viper.IsSet("DBName") || !viper.IsSet("DBUser") || !viper.IsSet("DBPassword") {
		return nil, fmt.Errorf("not all database requisites are set, need set dbhost, dbname, dbuser, dbpassword")
	}

	conf.Host = viper.GetString("DBHost")
	conf.Port = viper.GetInt("DBPort")
	conf.DBName = viper.GetString("DBName")
	conf.User = viper.GetString("DBUser")
	conf.Password = viper.GetString("DBPassword")
	conf.SSLMode = viper.GetString("DBSSLMode")
	conf.Timeout = viper.GetDuration("DBTimeout")

	return &conf, nil
}

func processAPIServerConf() *ServerConf {
	viper.SetDefault("api-server.host", defaultAPIServerHost)
	viper.SetDefault("api-server.port", defaultAPIServerPort)

	conf := ServerConf{}
	conf.Host = viper.GetString("api-server.host")
	conf.Port = viper.GetInt("api-server.port")

	return &conf
}
