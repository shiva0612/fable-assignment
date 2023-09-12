package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"time"

	utils "fable/utils"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var (
	kr                             *kafka.Conn
	cfg                            *ConsumerConfig
	Log                            *zap.SugaredLogger
	toBeInserted, inserted, failed int64
)

type ConsumerConfig struct {
	Broker   string `mapstructure:"broker"`
	Topic    string `mapstructure:"topic"`
	Dbhost   string `mapstructure:"dbHost"`
	DbPort   string `mapstructure:"dbPort"`
	DbUser   string `mapstructure:"dbUser"`
	DbPsw    string `mapstructure:"dbPsw"`
	DbName   string `mapstructure:"dbName"`
	LogLevel string `mapstructure:"logLevel"`
}

func createConsumer() error {

	var err error
	time.Sleep(2 * time.Second)
	kr, err = kafka.DialLeader(context.Background(), "tcp", cfg.Broker, cfg.Topic, 0)
	if err != nil {
		Log.Error("failed to dial leader: " + err.Error())
		return err
	}
	return nil
}

func main() {

	cfg = new(ConsumerConfig)
	configPath := flag.String("config", "./consumer.json", "")
	flag.Parse()
	err := utils.LoadConfig(*configPath, cfg)
	if err != nil {
		return
	}
	Log = utils.ConfigureZapLogger(cfg.LogLevel)

	err = configureDatabase()
	if err != nil {
		return
	}
	defer db.Close()

	err = createConsumer()
	if err != nil {
		return
	}
	defer kr.Close()

	ctx, cancel := context.WithCancel(context.Background())
	exitch := make(chan error)
	go utils.ShutDownHandler(exitch)

	go read_msgs(exitch, ctx)

	<-exitch
	Log.Info("tobeInserted = ", toBeInserted)
	Log.Info("inserted = ", inserted)
	Log.Info("failed = ", failed)
	cancel()
}

func read_msgs(exitch chan error, ctx context.Context) {

	Log.Info("reading msgs...")

Loop:
	for {
		batch := kr.ReadBatch(1<<10, 10*1<<10*1<<10)
		if batch.Err() != nil && batch.Err() != io.EOF {
			Log.Debug("error reading batch: ", batch.Err().Error())
			continue
		}
		lips := []utils.LogInput{}
		for {
			lip := new(utils.LogInput)
			km, err := batch.ReadMessage()
			if err != nil {
				if err == context.Canceled {
					Log.Debug("context cancelled exiting from kafka reading: " + err.Error())
					break Loop
				}
				if err != io.EOF {
					Log.Debug("error reading batch: ", err.Error())
				}
				break
			}
			err = json.Unmarshal(km.Value, lip)
			if err != nil {
				Log.Debug("json marshalling error: " + err.Error())
				continue
			}
			lips = append(lips, *lip)
		}
		batch.Close()

		Log.Debug("writing %d msgs to db: \n", len(lips))
		writeToDB(ctx, lips)
	}

	Log.Info("exiting reading kafka msgs...")
}
