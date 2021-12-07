package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/simonnik/GB_Best_CourseWork_GO/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	var BuildGitHash string
	fmt.Printf("Executable file: %s.\nBuild on commit %s\n", os.Args[0], BuildGitHash)
	accessLogger := newLogger("logs/access.log")
	errorLogger := newLogger("logs/error.log")

	cfg, err := config.NewConfig()
	if err != nil {
		errorLogger.Fatalf("Cannot read configuration")
	}
	accessLogger.Info("Initialize")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.App.Timeout)*time.Second)

	sigCh := make(chan os.Signal, 1) // Создаем канал для приема сигналов, SIGUSR1, (lint: gocritic)
	for {
		select {
		case <-ctx.Done(): // Если всё завершили - выходим
			msg := "Вышло время выполнения\n"
			errorLogger.Info(msg)
			fmt.Print(msg)
			return
		case sig := <-sigCh:
			if sig == syscall.SIGINT {
				errorLogger.Error("Получили сигнал SIGINT")
				cancel() // Если пришёл сигнал SigInt - завершаем контекст
				return
			}
		}
	}
}

func newLogger(logFile string) *zap.SugaredLogger {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{logFile}
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	logger, err := cfg.Build()
	if err != nil {
		log.Fatalf("logging %s not initialize: %v", logFile, err)
	}

	return logger.Sugar()
}
