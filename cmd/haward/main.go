package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Feresey/haward/rules"
	"github.com/Feresey/haward/session"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type flags struct {
	logsDir      string
	outputFile   string
	rulesFile    string
	yourNickname string
}

func main() {
	var f flags

	// TODO корпорацию того же чела не считать
	// TODO парсить правила
	// TODO раскидать нормально этот файл
	// TODO переименованные жопа c кланом

	flag.StringVar(&f.logsDir, "dir", ".local/share/starconflict/logs", "Path to logs directory")
	flag.StringVar(&f.outputFile, "o", "out.csv", "Path to the output file")
	flag.StringVar(&f.rulesFile, "rules", "rules.txt", "Path to the rules file")
	flag.StringVar(&f.yourNickname, "nick", "ZiroTwo", "Your nickname")
	flag.Parse()

	lc := zap.NewDevelopmentConfig()
	lc.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	lc.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

	logger, err := lc.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	rulesFile, err := os.Open(f.rulesFile)
	if err != nil {
		logger.Fatal("", zap.Error(err))
	}
	defer rulesFile.Close()

	rules, err := rules.NewRules(rulesFile)
	if err != nil {
		logger.Fatal("parse rules", zap.Error(err))
	}

	p := &Parser{
		f:      f,
		logger: logger,
		rules:  rules,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGILL, syscall.SIGTERM)
	defer cancel()

	logger.Info("start parse")
	if err := p.run(ctx); err != nil {
		logger.Fatal("", zap.Error(err))
	}
}

type Parser struct {
	f flags

	logger *zap.Logger
	rules  *rules.Rules
}

func (p *Parser) run(ctx context.Context) error {
	sessions, err := getSessionList(p.f.logsDir)
	if err != nil {
		return fmt.Errorf("scan sessions: %w", err)
	}
	p.logger.Debug("session list", zap.Strings("sessions", sessions))

	// TODO
	output, err := os.Create(p.f.outputFile)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer output.Close()

	w := csv.NewWriter(output)

	if err := w.Write(NewReportIter(nil).Header()); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}
	w.Flush()
	p.logger.Info("write csv header")

	for _, session := range sessions {
		p.logger.Info("start process session", zap.String("session", session))

		sessionReport, err := p.parseSession(ctx, session)
		if err != nil {
			return fmt.Errorf("parse session: %s :%w", session, err)
		}

		p.logger.Info("write report")
		ri := NewReportIter(sessionReport)
		for ri.Next() {
			err := w.Write(ri.Line())
			if err != nil {
				return fmt.Errorf("write result linte: %w", err)
			}
		}
		p.logger.Info("flush level result")
		w.Flush()
	}

	p.logger.Info("flush output")
	w.Flush()
	return w.Error()
}

const sessionTimeFormat = "2006.01.02 15.04.05.999"

func (p *Parser) parseSession(ctx context.Context, sessionName string) (*SessionReport, error) {
	startedAt, err := time.Parse(sessionTimeFormat, sessionName)
	if err != nil {
		return nil, fmt.Errorf("parse session date: %q: %w", sessionName, err)
	}

	combat, err := os.Open(filepath.Join(p.f.logsDir, sessionName, "combat.log"))
	if err != nil {
		return nil, fmt.Errorf("open combat log: %w", err)
	}
	game, err := os.Open(filepath.Join(p.f.logsDir, sessionName, "game.log"))
	if err != nil {
		return nil, fmt.Errorf("open game log: %w", err)
	}

	parser := session.NewParser(p.f.yourNickname, combat, game, p.rules)

	done := make(chan error, 1)
	levelReports := make(chan *session.LevelReport)

	p.logger.Debug("start session parser")
	go func() {
		err := parser.Parse(ctx, levelReports)
		lvl := zapcore.WarnLevel
		if errors.Is(err, io.EOF) {
			err = nil
			lvl = zapcore.DebugLevel
		}
		p.logger.Check(lvl, "goroutine stopped").Write(zap.Error(err))
		done <- err
		close(levelReports)
	}()

	var s SessionReport

	s.StartedAt = startedAt

	for levelReport := range levelReports {
		lvl := zapcore.DebugLevel
		if len(levelReport.Score) != 0 {
			s.Levels = append(s.Levels, levelReport)
			lvl = zapcore.InfoLevel
		}
		p.logger.Check(lvl, "got level report").Write(zap.Int("length", len(levelReport.Score)))
	}

	p.logger.Debug("wait for error")
	return &s, <-done
}

func getSessionList(logsDir string) ([]string, error) {
	sessions, err := os.ReadDir(logsDir)
	if err != nil {
		return nil, err
	}

	var res []string

	for _, session := range sessions {
		if session.IsDir() {
			res = append(res, session.Name())
		}
	}

	return res, nil
}
