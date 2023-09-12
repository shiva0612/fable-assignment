package utils

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ConfigureZapLogger(level string) *zap.SugaredLogger {
	config := zap.NewProductionConfig()
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		log.Fatalln("invalid log level: ", level)
	}
	config.Level.SetLevel(lvl)
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	config.Encoding = "console"
	zlog, _ := config.Build(zap.AddStacktrace(zap.ErrorLevel)) //prints stacktrace only for error level and above
	return zlog.Sugar()
}

type LogInput struct {
	Id        int64  `json:"id"`
	TimeStamp int64  `json:"unix_ts"`
	UserId    int64  `json:"user_id"`
	Event     string `json:"event_name"`
}

func ShutDownHandler(exitch chan error) {
	ch := make(chan os.Signal, 10)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-ch
	signal.Stop(ch)
	close(ch)
	exitch <- errors.New("graceful shutdown received")
	close(exitch)
}

func LoadConfig(path string, cls interface{}) error {
	viper.AddConfigPath(filepath.Dir(path))
	viper.SetConfigName(filepath.Base(path))
	viper.SetConfigType(filepath.Ext(path)[1:]) //.json

	err := viper.ReadInConfig()
	if err != nil {
		log.Println("error while reading the config: " + err.Error())
		return err
	}

	err = viper.Unmarshal(cls)
	if err != nil {
		log.Println("error while unmarshalling config.json: " + err.Error())
		return err
	}
	return nil
}
