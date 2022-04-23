package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/caarlos0/env/v6"
	"github.com/taiyoh/toyhose"
)

var (
	version string = "HEAD"
	commit  string = "not-registered"
)

func main() {
	if err := toyhose.SetupLogger(); err != nil {
		panic(err)
	}
	log := toyhose.Logger()

	var conf toyhoseConfig
	if err := env.Parse(&conf); err != nil {
		log.Fatal().Err(err).Msg("invalid environment variables")
	}
	awsConf := aws.NewConfig().
		WithRegion(conf.Region).
		WithCredentials(credentials.NewStaticCredentials(conf.AccessKeyID, conf.SecretKey, ""))
	d := toyhose.NewDispatcher(&toyhose.DispatcherConfig{
		AWSConf: awsConf,
		S3InjectedConf: toyhose.S3InjectedConf{
			SizeInMBs:         conf.S3SizeInMBs,
			IntervalInSeconds: conf.S3IntervalInSeconds,
			EndPoint:          conf.S3EndPoint,
			DisableBuffering:  conf.S3DisableBuffering,
		},
		KinesisInjectedConf: toyhose.KinesisInjectedConf{
			Endpoint: conf.KinesisEndpoint,
		},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/", d.Dispatch)

	srv := &http.Server{
		Handler: mux,
		Addr:    fmt.Sprintf(":%d", conf.Port),
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		os.Interrupt)

	defer cancel()

	go func() {
		<-ctx.Done()
		sdCtx, sdCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer sdCancel()
		log.Info().Msgf("shutdown process...")
		if err := srv.Shutdown(sdCtx); err != nil {
			log.Error().Err(err).Msg("Shutdown process failed")
		}
	}()

	log.Info().Int("port", conf.Port).
		Str("version", version).
		Str("commit", commit).Msgf("starting toyhose server")

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error().Err(err).Msg("ListenAndServe failed")
	}
}

type toyhoseConfig struct {
	AccessKeyID         string  `env:"AWS_ACCESS_KEY_ID,true"`
	SecretKey           string  `env:"AWS_SECRET_ACCESS_KEY,true"`
	Region              string  `env:"AWS_REGION"            envDefault:"us-east-1"`
	Port                int     `env:"PORT"                  envDefault:"4573"` // inspired by localstack
	S3DisableBuffering  bool    `env:"S3_DISABLE_BUFFERING"  envDefault:"false"`
	S3SizeInMBs         *int    `env:"S3_BUFFERING_HINTS_SIZE_IN_MBS"`
	S3IntervalInSeconds *int    `env:"S3_BUFFERING_HINTS_INTERVAL_IN_SECONDS"`
	S3EndPoint          *string `env:"S3_ENDPOINT_URL"`
	KinesisEndpoint     *string `env:"KINESIS_STREAM_ENDPOINT_URL"`
}
