package pkg

import "go.uber.org/zap"

var Logger *zap.SugaredLogger

func init() {
	var logger, _ = zap.NewProduction()
	var sugar = logger.Sugar()
	Logger = sugar
	defer logger.Sync()
}
