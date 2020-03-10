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
	"github.com/kelseyhightower/envconfig"
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

	conf := &toyhoseConfig{}
	if err := envconfig.Process("", conf); err != nil {
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh,
			syscall.SIGTERM,
			syscall.SIGINT,
			os.Interrupt)
		for {
			select {
			case <-ctx.Done():
				sdCtx, sdCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer sdCancel()
				log.Info().Msgf("shutdown process...")
				if err := srv.Shutdown(sdCtx); err != nil {
					log.Error().Err(err).Msg("Shutdown process failed")
				}
				return
			case c := <-sigCh:
				log.Info().Interface("signal", c).Msgf("caught signal")
				switch c {
				case syscall.SIGTERM, syscall.SIGINT, os.Interrupt:
					cancel()
				}
			}
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
	AccessKeyID         string  `envconfig:"AWS_ACCESS_KEY_ID"     required:"true"`
	SecretKey           string  `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`
	Region              string  `envconfig:"AWS_REGION"            default:"us-east-1"`
	Port                int     `envconfig:"PORT"                  default:"4573"` // inspired by localstack
	S3DisableBuffering  bool    `envconfig:"S3_DISABLE_BUFFERING"  default:"false"`
	S3SizeInMBs         *int    `envconfig:"S3_BUFFERING_HINTS_SIZE_IN_MBS"`
	S3IntervalInSeconds *int    `envconfig:"S3_BUFFERING_HINTS_INTERVAL_IN_SECONDS"`
	S3EndPoint          *string `envconfig:"S3_ENDPOINT_URL"`
	KinesisEndpoint     *string `envconfig:"KINESIS_STREAM_ENDPOINT_URL"`
}
