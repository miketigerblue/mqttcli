package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Config holds all the MQTT connection and subscription details.
type Config struct {
	// MQTT connection details
	BrokerURL string `json:"broker_url"` // e.g. "ssl://your-iot-endpoint.amazonaws.com:8883" or "tcp://localhost:1883"
	ClientID  string `json:"client_id"`  // e.g. "myTestClient"
	Username  string `json:"username"`   // optional for AWS IoT; sometimes used for other brokers
	Password  string `json:"password"`   // optional for AWS IoT; sometimes used for other brokers
	CAFile    string `json:"ca_file"`    // path to root CA cert (e.g. AmazonRootCA1.pem)
	CertFile  string `json:"cert_file"`  // path to device/client certificate
	KeyFile   string `json:"key_file"`   // path to private key
	Insecure  bool   `json:"insecure"`   // skip server cert validation (not recommended in production)

	// Subscription details
	Topic       string `json:"topic"`        // e.g. "iot/gnss/+/data"
	QoS         byte   `json:"qos"`          // 0, 1, or 2
	Quiet       bool   `json:"quiet"`        // if true, don’t print incoming messages
	PrintErrors bool   `json:"print_errors"` // if true, log or print errors verbosely

	// Optional: Publish details (could be extended to allow a publish payload, etc.)
}

// loadConfig reads a JSON file into a Config struct.
func loadConfig(configPath string) (*Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// overrideWithFlags sets any non-zero CLI flags into the Config struct to allow easy overrides.
func overrideWithFlags(cfg *Config, flags *cliFlags) {
	if flags.BrokerURL != "" {
		cfg.BrokerURL = flags.BrokerURL
	}
	if flags.ClientID != "" {
		cfg.ClientID = flags.ClientID
	}
	if flags.Username != "" {
		cfg.Username = flags.Username
	}
	if flags.Password != "" {
		cfg.Password = flags.Password
	}
	if flags.Topic != "" {
		cfg.Topic = flags.Topic
	}
	if flags.CAFile != "" {
		cfg.CAFile = flags.CAFile
	}
	if flags.CertFile != "" {
		cfg.CertFile = flags.CertFile
	}
	if flags.KeyFile != "" {
		cfg.KeyFile = flags.KeyFile
	}
	if flags.QoS >= 0 {
		cfg.QoS = byte(flags.QoS)
	}
	if flags.Insecure {
		cfg.Insecure = true
	}
	if flags.Quiet {
		cfg.Quiet = true
	}
	if flags.PrintErrors {
		cfg.PrintErrors = true
	}
}

type cliFlags struct {
	ConfigPath  string
	BrokerURL   string
	ClientID    string
	Username    string
	Password    string
	Topic       string
	CAFile      string
	CertFile    string
	KeyFile     string
	QoS         int
	Insecure    bool
	Quiet       bool
	PrintErrors bool
}

// initCLIFlags defines our command-line flags with usage text.
func initCLIFlags() *cliFlags {
	var f cliFlags

	flag.StringVar(&f.ConfigPath, "config", "", "Path to JSON config file (optional). If provided, this file is loaded first.")
	flag.StringVar(&f.BrokerURL, "broker", "", "Broker URL, e.g. 'ssl://<endpoint>:8883' or 'tcp://localhost:1883'")
	flag.StringVar(&f.ClientID, "clientid", "", "MQTT client ID (must be unique per broker).")
	flag.StringVar(&f.Username, "username", "", "MQTT username if broker requires it.")
	flag.StringVar(&f.Password, "password", "", "MQTT password if broker requires it.")
	flag.StringVar(&f.Topic, "topic", "", "MQTT topic to subscribe to.")
	flag.StringVar(&f.CAFile, "cafile", "", "Path to root CA certificate file (e.g. AmazonRootCA1.pem).")
	flag.StringVar(&f.CertFile, "certfile", "", "Path to client certificate file (x.509).")
	flag.StringVar(&f.KeyFile, "keyfile", "", "Path to client private key file.")
	flag.IntVar(&f.QoS, "qos", -1, "QoS level for subscription (0, 1, or 2).")
	flag.BoolVar(&f.Insecure, "insecure", false, "Skip TLS server cert verification (NOT recommended).")
	flag.BoolVar(&f.Quiet, "quiet", false, "If set, do not print incoming messages.")
	flag.BoolVar(&f.PrintErrors, "verbose-errors", false, "Print errors verbosely if set.")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			`Usage: %s [options]

This utility subscribes to an MQTT topic using Eclipse Paho, supporting optional TLS for
AWS IoT Core or other brokers. Configuration can come from both a JSON file and CLI flags.
CLI flags override JSON values.

Options:
`, filepath.Base(os.Args[0]))
		flag.PrintDefaults()

		fmt.Println(`
Examples:

  # Basic local broker usage:
  mqttcli --broker "tcp://localhost:1883" --clientid "testClient" \
          --topic "my/test/topic" --qos 1

  # Using AWS IoT Core with mutual TLS:
  mqttcli --broker "ssl://<endpoint>.amazonaws.com:8883" \
          --clientid "myThing" \
          --cafile "AmazonRootCA1.pem" \
          --certfile "deviceCert.crt" \
          --keyfile "deviceKey.key" \
          --topic "iot/gnss/myThing/data" --qos 1

  # JSON config usage:
  mqttcli --config /path/to/config.json
`)
	}

	return &f
}

// messageHandler prints incoming messages (unless quiet).
func messageHandler(cfg *Config) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		if !cfg.Quiet {
			fmt.Printf("[MSG RECEIVED] Topic=%s QoS=%d Payload=%s\n",
				msg.Topic(), msg.Qos(), msg.Payload())
		}
	}
}

// connectMQTT sets up and connects an MQTT client based on the provided Config.
func connectMQTT(cfg *Config) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.BrokerURL)
	opts.SetClientID(cfg.ClientID)
	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}

	// Set up TLS config if using ssl://
	if err := configureTLS(opts, cfg); err != nil {
		return nil, err
	}

	// OnConnectionLost
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		if cfg.PrintErrors {
			log.Printf("[ERROR] MQTT connection lost: %v", err)
		}
	}

	// Create and start connection
	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	if err := token.Error(); err != nil {
		return nil, err
	}

	return client, nil
}

func configureTLS(opts *mqtt.ClientOptions, cfg *Config) error {
	// Only configure TLS if scheme is "ssl" or user provided CA/cert files
	isSSL := false
	if len(cfg.BrokerURL) > 5 {
		isSSL = (cfg.BrokerURL[0:5] == "ssl://")
	}

	if isSSL || cfg.CAFile != "" || cfg.CertFile != "" || cfg.KeyFile != "" {
		tlsConfig, err := NewTLSConfig(cfg.CAFile, cfg.CertFile, cfg.KeyFile, cfg.Insecure)
		if err != nil {
			return err
		}
		opts.SetTLSConfig(tlsConfig)
	}
	return nil
}

// subscribeToTopic subscribes to the configured topic and waits for messages.
func subscribeToTopic(client mqtt.Client, cfg *Config, handler mqtt.MessageHandler) error {
	token := client.Subscribe(cfg.Topic, cfg.QoS, handler)
	token.Wait()
	return token.Error()
}

func main() {
	// 1. Parse CLI flags
	flags := initCLIFlags()
	flag.Parse()

	// 2. Load config file if provided
	var cfg Config
	if flags.ConfigPath != "" {
		loadedCfg, err := loadConfig(flags.ConfigPath)
		if err != nil {
			log.Fatalf("[ERROR] could not load config file: %v\n", err)
		}
		cfg = *loadedCfg
	}

	// 3. Override config with CLI flags (if set)
	overrideWithFlags(&cfg, flags)

	// 4. Validate minimal required fields
	if cfg.BrokerURL == "" {
		log.Fatalf("[ERROR] Broker URL is not set. Provide via --broker or config file.")
	}
	if cfg.ClientID == "" {
		log.Fatalf("[ERROR] Client ID is not set. Provide via --clientid or config file.")
	}
	if cfg.Topic == "" {
		log.Fatalf("[ERROR] Topic is not set. Provide via --topic or config file.")
	}
	// For QoS, if not set, default to 0.
	if cfg.QoS != 0 && cfg.QoS != 1 && cfg.QoS != 2 {
		cfg.QoS = 0
	}

	// 5. Connect to MQTT broker
	client, err := connectMQTT(&cfg)
	if err != nil {
		log.Fatalf("[ERROR] MQTT connection failed: %v", err)
	}
	defer client.Disconnect(250)

	log.Printf("[INFO] Connected to %s as clientID='%s'", cfg.BrokerURL, cfg.ClientID)

	// 6. Subscribe to topic
	if err := subscribeToTopic(client, &cfg, messageHandler(&cfg)); err != nil {
		log.Fatalf("[ERROR] Failed to subscribe to topic '%s': %v\n", cfg.Topic, err)
	}
	log.Printf("[INFO] Subscribed to topic '%s' with QoS=%d", cfg.Topic, cfg.QoS)

	// 7. Handle graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("[INFO] Shutting down...")
	// Optional cleanup, e.g. unsubscribe:
	// client.Unsubscribe(cfg.Topic).Wait()

	// Wait briefly to ensure final logs/messages are handled
	time.Sleep(1 * time.Second)
	log.Println("[INFO] Exiting.")
}
