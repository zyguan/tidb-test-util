package log

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type glogEncoder struct {
	*zapcore.EncoderConfig
	internal zapcore.Encoder
}

func (enc glogEncoder) AddArray(key string, arr zapcore.ArrayMarshaler) error {
	return enc.internal.AddArray(key, arr)
}

func (enc glogEncoder) AddObject(key string, obj zapcore.ObjectMarshaler) error {
	return enc.internal.AddObject(key, obj)
}

func (enc glogEncoder) AddBinary(key string, val []byte) {
	enc.internal.AddBinary(key, val)
}

func (enc glogEncoder) AddByteString(key string, val []byte) {
	enc.internal.AddByteString(key, val)
}

func (enc glogEncoder) AddBool(key string, val bool) {
	enc.internal.AddBool(key, val)
}

func (enc glogEncoder) AddComplex128(key string, val complex128) {
	enc.internal.AddComplex128(key, val)
}

func (enc glogEncoder) AddComplex64(key string, val complex64) {
	enc.internal.AddComplex64(key, val)
}

func (enc glogEncoder) AddDuration(key string, val time.Duration) {
	enc.internal.AddDuration(key, val)
}

func (enc glogEncoder) AddFloat64(key string, val float64) {
	enc.internal.AddFloat64(key, val)
}

func (enc glogEncoder) AddFloat32(key string, val float32) {
	enc.internal.AddFloat32(key, val)
}

func (enc glogEncoder) AddInt(key string, val int) {
	enc.internal.AddInt(key, val)
}

func (enc glogEncoder) AddInt64(key string, val int64) {
	enc.internal.AddInt64(key, val)
}

func (enc glogEncoder) AddInt32(key string, val int32) {
	enc.internal.AddInt32(key, val)
}

func (enc glogEncoder) AddInt16(key string, val int16) {
	enc.internal.AddInt16(key, val)
}

func (enc glogEncoder) AddInt8(key string, val int8) {
	enc.internal.AddInt8(key, val)
}

func (enc glogEncoder) AddString(key, val string) {
	enc.internal.AddString(key, val)
}

func (enc glogEncoder) AddTime(key string, val time.Time) {
	enc.internal.AddTime(key, val)
}

func (enc glogEncoder) AddUint(key string, val uint) {
	enc.internal.AddUint(key, val)
}

func (enc glogEncoder) AddUint64(key string, val uint64) {
	enc.internal.AddUint64(key, val)
}

func (enc glogEncoder) AddUint32(key string, val uint32) {
	enc.internal.AddUint32(key, val)
}

func (enc glogEncoder) AddUint16(key string, val uint16) {
	enc.internal.AddUint16(key, val)
}

func (enc glogEncoder) AddUint8(key string, val uint8) {
	enc.internal.AddUint8(key, val)
}

func (enc glogEncoder) AddUintptr(key string, val uintptr) {
	enc.internal.AddUintptr(key, val)
}

func (enc glogEncoder) AddReflected(key string, val interface{}) error {
	return enc.internal.AddReflected(key, val)
}

func (enc glogEncoder) OpenNamespace(key string) {
	enc.internal.OpenNamespace(key)
}

func (enc glogEncoder) Clone() zapcore.Encoder {
	return glogEncoder{
		EncoderConfig: enc.EncoderConfig,
		internal:      enc.internal.Clone(),
	}
}

func (enc glogEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf := bufpool.Get()

	switch ent.Level {
	case zapcore.DebugLevel:
		buf.AppendByte('D')
	case zapcore.InfoLevel:
		buf.AppendByte('I')
	case zapcore.WarnLevel:
		buf.AppendByte('W')
	case zapcore.ErrorLevel:
		buf.AppendByte('E')
	case zapcore.DPanicLevel, zap.PanicLevel:
		buf.AppendByte('P')
	case zapcore.FatalLevel:
		buf.AppendByte('F')
	default:
		buf.AppendString(fmt.Sprintf("%d!", ent.Level))
	}

	_, month, day := ent.Time.Date()
	hour, minute, second := ent.Time.Clock()

	appendTwoDigit(buf, int(month))
	appendTwoDigit(buf, day)
	buf.AppendByte(' ')
	appendTwoDigit(buf, hour)
	buf.AppendByte(':')
	appendTwoDigit(buf, minute)
	buf.AppendByte(':')
	appendTwoDigit(buf, second)
	buf.AppendByte('.')
	appendTwoDigit(buf, ent.Time.Nanosecond()/1e7%100)
	appendTwoDigit(buf, ent.Time.Nanosecond()/1e5%100)
	appendTwoDigit(buf, ent.Time.Nanosecond()/1e3%100)

	k := 8
	for n := pid; n > 0; k-- {
		n /= 10
	}
	for ; k > 0; k-- {
		buf.AppendByte(' ')
	}
	buf.AppendInt(int64(pid))
	buf.AppendByte(' ')
	buf.AppendString(ent.Caller.TrimmedPath())
	buf.AppendString("] ")
	if len(ent.LoggerName) > 0 && len(enc.NameKey) == 0 {
		buf.AppendString("[" + ent.LoggerName + "] ")
	}
	buf.AppendString(ent.Message)

	json, _ := enc.internal.EncodeEntry(ent, fields)
	str := strings.TrimSpace(json.String())
	if len(str) > 2 {
		buf.AppendByte('\t')
		buf.AppendString(str)
	}
	json.Free()

	if ent.Stack != "" && enc.StacktraceKey != "" {
		buf.AppendByte('\n')
		buf.AppendString(ent.Stack)
	}

	buf.AppendString(enc.LineEnding)

	return buf, nil
}

func appendTwoDigit(buf *buffer.Buffer, n int) {
	buf.AppendByte(digists[n/10%10])
	buf.AppendByte(digists[n%10])
}

const digists = "0123456789"

var (
	pid     = os.Getpid()
	bufpool = buffer.NewPool()
)

func NewGLogEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	cfgCopy := cfg
	cfgCopy.MessageKey = ""
	cfgCopy.TimeKey = ""
	cfgCopy.LevelKey = ""
	cfgCopy.CallerKey = ""
	return glogEncoder{
		EncoderConfig: &cfg,
		internal:      zapcore.NewJSONEncoder(cfgCopy),
	}
}

func NewGLogDevConfig() zap.Config {
	return zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:      true,
		Encoding:         "glog",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

func NewGLogDev(options ...zap.Option) (*zap.Logger, error) {
	return NewGLogDevConfig().Build(options...)
}

func init() {
	zap.RegisterEncoder("glog", func(config zapcore.EncoderConfig) (encoder zapcore.Encoder, e error) {
		return NewGLogEncoder(config), nil
	})
}
