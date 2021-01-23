package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amimof/huego"
	"github.com/pion/dtls/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Lightshow is the main Lightshow object
type Lightshow struct {
	Bridge            *huego.Bridge
	Config            Config
	Logger            *logrus.Entry
	Context           context.Context
	ContextCancelFunc context.CancelFunc
	Group             *huego.Group
	DTLSConn          *dtls.Conn
	MessageBytes      []byte
}

// Config is where config gets marshalled/unmarshalled from/to
type Config struct {
	Hue struct {
		Address     string
		User        string
		Secret      string
		FrequencyHz float64
	}
	Lights []struct {
		ID          int
		Coordinates struct {
			X float64
			Y float64
		}
	}
	Log struct {
		Level string
	}
}

var lightshow Lightshow

func main() {

	// Set up logger
	logger := PrepareLogger()

	// Parse config
	config, err := PrepareConfig(logger)
	if err != nil {
		os.Exit(1)
	}

	logLevel, err := logrus.ParseLevel(config.Log.Level)
	if err != nil {
		logger.WithError(err).Warn("Failed to parse log level, staying at default level of info")
	} else {
		logger.Logger.SetLevel(logLevel)
	}

	ctx, ctxCancel := SetupContext(logger)

	// Instantiate object
	lightshow = Lightshow{
		Config:            *config,
		Logger:            logger,
		Context:           ctx,
		ContextCancelFunc: ctxCancel,
	}
	defer lightshow.SaveConfig()

	// Connect to API
	err = lightshow.ConnectToHueAPI()
	if err != nil {
		os.Exit(1)
	}

	// Configure Entertainment Zone
	lightshow.ConfigureEntertainmentZone()
	defer lightshow.DeleteEntertainmentZone()

	// Start streaming
	lightshow.StartDTLSStreaming()
	defer lightshow.Group.DisableStreaming()
	defer lightshow.DTLSConn.Close()

	go func() {
		r, g, b := 0, 0, 0
		bri := 1
		for {
			if (r == 0) && (g == 0) && (b == 0) {
				r = 1
				bri = 1
			} else if (r == 1) && (g == 0) && (b == 0) {
				r = 0
				g = 1
			} else if (r == 0) && (g == 1) && (b == 0) {
				g = 0
				b = 1
			} else if (r == 0) && (g == 0) && (b == 1) {
				b = 0
				bri = 0
			}
			for _, light := range []int{7, 5, 8, 13, 15, 11, 10, 12, 14, 9} {
				lightshow.SetLights([]int{light}, r*255, g*255, b*255, float32(bri))
				time.Sleep(time.Millisecond * 100)
			}
		}
	}()

	// Loop until we get #cancelled
	<-ctx.Done()
}

// PrepareLogger sets up logrus
func PrepareLogger() *logrus.Entry {
	logrus.SetReportCaller(true)

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	return logrus.WithFields(logrus.Fields{
		"app": "lightshow",
	})
}

// PrepareConfig reads config from the config file
func PrepareConfig(log *logrus.Entry) (*Config, error) {
	var config *Config

	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	viper.SetDefault("log.level", "info")
	viper.SetDefault("hue.frequencyhz", 12.5)

	err := viper.ReadInConfig()
	if err != nil {
		log.WithError(err).Error("Failed to read in config")
		return nil, err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.WithError(err).Error("Failed to unmarshal config")
		return nil, err
	}

	return config, nil
}

// SaveConfig writes back changes to the config file
func (l *Lightshow) SaveConfig() error {

	// Convert config to JSON
	jsonBytes, err := json.Marshal(l.Config)
	if err != nil {
		l.Logger.WithError(err).Error("Failed to marshal configuration")
		return err
	}

	// Merge config
	byteReader := bytes.NewReader(jsonBytes)
	err = viper.MergeConfig(byteReader)
	if err != nil {
		l.Logger.WithError(err).Error("Failed to merge configuration")
		return err
	}

	// Write the config file
	err = viper.WriteConfig()
	if err != nil {
		l.Logger.WithError(err).Error("Failed to write config file")
		return err
	}

	return nil
}

// SetupContext configures the overall context that is passed down through the program
func SetupContext(log *logrus.Entry) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	signals := []os.Signal{
		syscall.SIGABRT,
		syscall.SIGALRM,
		syscall.SIGBUS,
		syscall.SIGHUP,
		syscall.SIGTERM,
		os.Interrupt,
		syscall.SIGILL,
		syscall.SIGQUIT,
		syscall.SIGSEGV,
		syscall.SIGTRAP,
	}

	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, signals...)

	go func() {
		sig := <-signalChan
		log.Infof("Caught signal %s, aborting execution", sig)
		cancel()
	}()

	return ctx, cancel
}
