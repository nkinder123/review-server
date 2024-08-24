package pkg

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func InitLog() (*zap.Logger, error) {
	//zap将日志以json的方式存入文件
	files, err := os.OpenFile("./review.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic("create log file has error")
	}
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	sync := zapcore.AddSync(files)
	core := zapcore.NewCore(encoder, sync, zapcore.DebugLevel)
	z := zap.New(core)
	return z, nil
}
