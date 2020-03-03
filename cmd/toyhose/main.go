package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/taiyoh/toyhose"
)

func main() {
	setupLogger()
	conf := &toyhoseConfig{}
	if err := envconfig.Process("", conf); err != nil {
		log.Fatal().Err(err).Msg("invalid environment variables")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", toyhose.NewDispatcher(conf.Dispatcher).Dispatch)

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

	log.Info().Int("port", conf.Port).Msg("starting toyhose server")

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error().Err(err).Msg("ListenAndServe failed")
	}
}

type toyhoseConfig struct {
	Dispatcher *toyhose.DispatcherConfig
	Port       int `envconfig:"PORT" default:"4573"`
}

func setupLogger() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = time.RFC3339Nano
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05.000000"}
	output.FormatLevel = func(i interface{}) string {
		return fmt.Sprintf("%-6s", i)
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf(" %s ", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s:", i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}
	output.FormatCaller = func(i interface{}) string {
		t := fmt.Sprintf("%s", i)
		s := strings.Split(t, ":")
		if 2 != len(s) {
			return t
		}
		f := filepath.Base(s[0])
		return f + ":" + s[1]
	}
	log.Logger = zerolog.New(output).With().Timestamp().Caller().Logger()
}
