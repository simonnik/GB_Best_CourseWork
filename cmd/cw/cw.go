package main

import (
	"bufio"
	"context"
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

	cfg, err := config.NewConfig()
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
		return
	}
	defer s.File.Close()
	h, err := s.GetHeaders()
	if err != nil {
		errorLogger.Error(err)
		cancel()
		return
	}
	accessLogger.With(zap.Strings("Headers", h)).Info("Getting headers from file")
	go s.Scan(ctx, q)
	if err != nil {
		errorLogger.Error(err)
		cancel()
		return
	}

	sigCh := make(chan os.Signal, 1) // Создаем канал для приема сигналов, SIGINT
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
		case msg := <-s.ChanResult():
			switch {
			case msg.Err != nil:
				err := fmt.Sprintf("Error: %s\n", msg.Err.Error())
				log.Print(err)
				errorLogger.Error(err)
				cancel()
				return
			case len(msg.Results) > 0:
				fmt.Printf("%s\n", msg.Results)
			case msg.Finished:
				fmt.Print("Finished\n")
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
