package logger

import "go.uber.org/zap"

var Log *zap.Logger = zap.NewNop()

func SetupLogger(level string) error {
	lvl, err := zap.ParseAtomicLevel(zap.InfoLevel.String())
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}
