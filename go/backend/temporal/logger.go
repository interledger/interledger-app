package temporal

import (
	"github.com/interledger/interledger-app/go/log"
	"go.uber.org/zap"
)

// temporalZapLogger adapts our zap logger to Temporal's logging interface
// so worker/client logs stay JSON-formatted and consistent.
func newTemporalLogger() temporalZapLogger {
	return temporalZapLogger{logger: log.Logger().Sugar()}
}

// temporalZapLogger satisfies go.temporal.io/sdk/log.Logger.
type temporalZapLogger struct {
	logger *zap.SugaredLogger
}

func (l temporalZapLogger) Debug(msg string, keyvals ...interface{}) {
	l.logger.Debugw(msg, keyvals...)
}

func (l temporalZapLogger) Info(msg string, keyvals ...interface{}) {
	l.logger.Infow(msg, keyvals...)
}

func (l temporalZapLogger) Warn(msg string, keyvals ...interface{}) {
	l.logger.Warnw(msg, keyvals...)
}

func (l temporalZapLogger) Error(msg string, keyvals ...interface{}) {
	l.logger.Errorw(msg, keyvals...)
}
