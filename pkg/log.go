package pkg

import "go.uber.org/zap"

var Logger *zap.SugaredLogger

func InitKafkaLogger() {
	var logger, _ = zap.NewProduction()
	var sugar = logger.Sugar()
	Logger = sugar
	defer logger.Sync()
}
