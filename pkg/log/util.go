package log

import "go.uber.org/zap"

var (
	l0 *zap.Logger
	l1 *zap.Logger
	s1 *zap.SugaredLogger
)

func Use(l *zap.Logger) {
	l0 = l
	if l0 != nil {
		l1 = l0.WithOptions(zap.AddCallerSkip(1))
		s1 = l1.Sugar()
	} else {
		l1 = nil
		s1 = nil
	}
}

func UseGLog(options ...zap.Option) {
	log, _ := NewGLog(options...)
	Use(log)
}

func UseGLogDev(options ...zap.Option) {
	log, _ := NewGLogDev(options...)
	Use(log)
}

func UseProduction(options ...zap.Option) {
	log, _ := zap.NewProduction(options...)
	Use(log)
}

func UseDevelopment(options ...zap.Option) {
	log, _ := zap.NewDevelopment(options...)
	Use(log)
}

func L() *zap.Logger {
	if l0 == nil {
		Use(zap.L())
	}
	return l0
}

func l() *zap.Logger {
	if l0 == nil {
		Use(zap.L())
	}
	return l1
}

func s() *zap.SugaredLogger {
	if l0 == nil {
		Use(zap.L())
	}
	return s1
}

func Debug(msg string, fields ...zap.Field) { l().Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)  { l().Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { l().Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { l().Error(msg, fields...) }
func Panic(msg string, fields ...zap.Field) { l().Panic(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { l().Fatal(msg, fields...) }

func Debugf(msg string, args ...interface{}) { s().Debugf(msg, args...) }
func Infof(msg string, args ...interface{})  { s().Infof(msg, args...) }
func Warnf(msg string, args ...interface{})  { s().Warnf(msg, args...) }
func Errorf(msg string, args ...interface{}) { s().Errorf(msg, args...) }
func Panicf(msg string, args ...interface{}) { s().Panicf(msg, args...) }
func Fatalf(msg string, args ...interface{}) { s().Fatalf(msg, args...) }

func Debugw(msg string, keysAndValues ...interface{}) { s().Debugw(msg, keysAndValues...) }
func Infow(msg string, keysAndValues ...interface{})  { s().Infow(msg, keysAndValues...) }
func Warnw(msg string, keysAndValues ...interface{})  { s().Warnw(msg, keysAndValues...) }
func Errorw(msg string, keysAndValues ...interface{}) { s().Errorw(msg, keysAndValues...) }
func Panicw(msg string, keysAndValues ...interface{}) { s().Panicw(msg, keysAndValues...) }
func Fatalw(msg string, keysAndValues ...interface{}) { s().Fatalw(msg, keysAndValues...) }
