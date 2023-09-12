package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	utils "fable/utils"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var (
	cfg *ProducerConfig
	kw  *kafka.Writer
	Log *zap.SugaredLogger
)

type ProducerConfig struct {
	Broker         string `mapstructure:"broker"`
	Topic          string `mapstructure:"topic"`
	ServerPort     string `mapstructure:"serverport"`
	MaxSize        int64  `mapstructure:"maxSize"`
	MaxTimeAllowed int64  `mapstructure:"maxTimeAllowed"`
}

var (
	toBeWrittenToKafka int
	writtenToKafka     int
)

func main() {

	cfg = new(ProducerConfig)
	configPath := flag.String("config", "./producer.json", "")
	flag.Parse()
	err := utils.LoadConfig(*configPath, cfg)
	if err != nil {
		return
	}

	exitCh := make(chan error, 1)
	go utils.ShutDownHandler(exitCh)

	//configure the kafka producer
	kw = &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Broker),
		Topic:                  cfg.Topic,
		WriteBackoffMax:        time.Duration(cfg.MaxTimeAllowed) * time.Second,
		WriteBackoffMin:        time.Duration(2 * time.Second),
		BatchSize:              int(cfg.MaxTimeAllowed) * 1000,
		BatchBytes:             cfg.MaxSize * 1 << 10 * 1 << 10, //10MB
		RequiredAcks:           1,
		AllowAutoTopicCreation: true,
	}
	defer kw.Close()

	//configure http server
	router := gin.Default()
	router.POST("/log", handleRequest)
	log.Println("starting server...")
	go router.Run(":" + cfg.ServerPort)
	<-exitCh
	// printStats()
}

func printStats() {
	fmt.Println("to be writtent to kakfa: ", toBeWrittenToKafka)
	fmt.Println("written to kafka: ", writtenToKafka)
}

func handleRequest(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.String(http.StatusMethodNotAllowed, "request not allowed")
		return
	}

	toBeWrittenToKafka++
	logInput := new(utils.LogInput)
	err := c.ShouldBindJSON(logInput)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid json request: "+err.Error())
		return
	}

	err = writeToKafka(logInput)
	if err != nil {
		c.String(http.StatusInternalServerError, "internal server error, please try again later...")
		return
	}

	writtenToKafka++
	c.String(http.StatusOK, "received the request, thanks")
}

func writeToKafka(lip *utils.LogInput) error {

	km, err := json.Marshal(lip)
	if err != nil {
		return err
	}
	err = kw.WriteMessages(context.Background(), kafka.Message{
		Value: km,
	})
	if err != nil {
		return err
	}
	return nil

}
