package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/go-git/go-git/v5"

	"github.com/simonnik/GB_Best_CourseWork_GO/services/parser"
	"github.com/simonnik/GB_Best_CourseWork_GO/services/scanner"

	"github.com/simonnik/GB_Best_CourseWork_GO/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	accessLogger := newLogger("logs/access.log")
	errorLogger := newLogger("logs/error.log")
	printLastCommit(errorLogger)
	configFile := flag.String("config", "config.yaml", "set path to configuration file")
	flag.Parse()
	cfg, err := config.NewConfig(configFile)
	if err != nil {
		errorLogger.Fatalf("Cannot read configuration")
		return
	}
	accessLogger.Info("Start")
	scan := bufio.NewScanner(os.Stdin)
	fmt.Print("\nВведите строку запроса:\r\n")
	scan.Scan()
	query := scan.Text()
	accessLogger.Infof("got sql query: %s", query)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.App.Timeout)*time.Second)
	p := parser.NewParser(query)
	q, err := p.Parse()
	if err != nil {
		errorLogger.Error(err)
		cancel()
		return
	}
	s, err := scanner.NewScanner(q.TableName)
	if err != nil {
		errorLogger.Error(err)
		cancel()
		log.Fatal(err)
	}
	defer s.File.Close()
	h, err := s.GetHeaders()
	if err != nil {
		errorLogger.Error(err)
		cancel()
		log.Print(err)
		return
	}
	accessLogger.With(zap.Strings("Headers", h)).Info("Getting headers from file")
	go s.Scan(ctx, *q)
	if err != nil {
		errorLogger.Error(err)
		cancel()
		log.Print(err)
		return
	}

	sigCh := make(chan os.Signal, 1) // Создаем канал для приема сигналов, SIGINT
	for {
		select {
		case <-ctx.Done(): // Если всё завершили - выходим
			msg := "Вышло время выполнения\n"
			errorLogger.Info(msg)
			log.Print(msg)
			return
		case sig := <-sigCh:
			if sig == syscall.SIGINT {
				msg := "Получили сигнал SIGINT"
				errorLogger.Error(msg)
				cancel() // Если пришёл сигнал SigInt - завершаем контекст
				log.Print(msg)
				return
			}
		case msg := <-s.ChanResult():
			switch {
			case msg.Err != nil:
				err := fmt.Sprintf("Error: %s\n", msg.Err.Error())
				errorLogger.Error(err)
				cancel()
				log.Print(err)
				return
			case len(msg.Results) > 0:
				fmt.Printf("%s\n", msg.Results)
			case msg.Finished:
				log.Print("Finished\n")
				accessLogger.Info("Finished")
				cancel()
				return
			}
		}
	}
}

func printLastCommit(errorLogger *zap.SugaredLogger) {
	r, err := git.PlainOpen("./")
	if err != nil {
		errorLogger.Error("Error open repository.",
			zap.String("repository", "./"),
			zap.Error(err),
		)
	}
	ref, err := r.Head()
	if err != nil {
		errorLogger.Error("Error get HEAD in git.",
			zap.String("repository", "./"),
			zap.Error(err),
		)
	}
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		errorLogger.Error("Error get last commit.",
			zap.String("repository", "./"),
			zap.Error(err),
		)
	}

	fmt.Printf("Executable file: %s.\nBuild on commit %s\n", os.Args[0], commit.Hash)
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
