package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/inenagl/anti-brute-force/internal/api"
	"github.com/inenagl/anti-brute-force/internal/app"
	lrucache "github.com/inenagl/anti-brute-force/internal/cache"
	"github.com/inenagl/anti-brute-force/internal/config"
	"github.com/inenagl/anti-brute-force/internal/logger"
	bucketstorage "github.com/inenagl/anti-brute-force/internal/storage/bucket"
	bwliststorage "github.com/inenagl/anti-brute-force/internal/storage/bwlist"
	_ "github.com/jackc/pgx/stdlib"
)

const envPrefix = "GOABF"

var configFile string

func main() {
	flag.StringVar(&configFile, "config", "/etc/abf/config.yaml", "Path to configuration file")
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	cfg, err := config.New(configFile, envPrefix)
	if err != nil {
		log.Fatalln(err)
	}

	lg, err := logger.New(
		cfg.Logger.Preset,
		cfg.Logger.Level,
		cfg.Logger.Encoding,
		cfg.Logger.OutputPaths,
		cfg.Logger.ErrorOutputPaths,
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer lg.Sync()

	bwListStorage := bwliststorage.New(
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.DBName,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.SSLMode,
		cfg.DB.Timeout,
	)
	if err = bwListStorage.Connect(); err != nil {
		lg.Sync()
		log.Fatalln(err) //nolint:gocritic
	}
	defer bwListStorage.Close()

	loginBucketStorage := bucketstorage.New(cfg.Main.BucketTTL)
	passwordBucketStorage := bucketstorage.New(cfg.Main.BucketTTL)
	ipBucketStorage := bucketstorage.New(cfg.Main.BucketTTL)
	cache := lrucache.New(cfg.Main.CacheSize, cfg.Main.CacheTTL)
	core := app.New(
		*lg,
		bwListStorage,
		loginBucketStorage,
		passwordBucketStorage,
		ipBucketStorage,
		cache,
		cfg.Main.MaxLogins,
		cfg.Main.MaxPasswords,
		cfg.Main.MaxIPs,
	)

	server := api.NewServer(cfg.APIServer.Host, cfg.APIServer.Port, *lg, core)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP,
	)
	defer cancel()

	go func() {
		<-ctx.Done()

		if err := server.Stop(); err != nil {
			lg.Error("failed to stop GRPC server: " + err.Error())
		}
	}()

	for _, storage := range []*bucketstorage.Storage{loginBucketStorage, passwordBucketStorage, ipBucketStorage} {
		go func(s *bucketstorage.Storage) {
			tk := time.NewTicker(cfg.Main.BucketTTL)
			for {
				select {
				case <-ctx.Done():
					lg.Debug("stop clearing of old buckets in storage")
					return
				case <-tk.C:
					lg.Debug("clear old buckets in storage")
					s.ClearByTTL()
				}
			}
		}(storage)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Start(); err != nil {
			lg.Error("failed to start GRPC server: " + err.Error())
			cancel()
		}
	}()

	wg.Wait()
}
