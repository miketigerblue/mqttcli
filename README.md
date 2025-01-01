# mqttcli

A simple, **scalable**, and **container-friendly** MQTT command-line tool, written in Go.  
Use it to connect, subscribe, and publish to any MQTT broker—including AWS IoT Core—via secure TLS or plain TCP.  

- Build Status: [![CI](https://github.com/miketigerblue/mqttcli/actions/workflows/ci.yml/badge.svg)](https://github.com/miketigerblue/mqttcli/actions/workflows/ci.yml)  
- Go Report Card: [![Go Report Card](https://goreportcard.com/badge/github.com/miketigerblue/mqttcli)](https://goreportcard.com/report/github.com/miketigerblue/mqttcli)  

## Features

- **Mutual TLS Support**  
  Connect to AWS IoT Core or any secure broker using CA, client certificate, and private key.  
- **JSON or CLI Configuration**  
  Specify broker settings and credentials via JSON config file, command-line flags, or both.  
- **Flexible**  
  Subscribe and (optionally) publish to multiple topics with configurable QoS levels.  
- **Container & Cloud Ready**  
  Multi-stage Dockerfile included for easy deployment to Kubernetes or other container environments.  
- **Graceful Shutdown**  
  Cleanly closes MQTT sessions on SIGINT/SIGTERM signals (Ctrl+C).  

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Examples](#examples)
  - [Basic Local Broker](#basic-local-broker)
  - [AWS IoT Core](#aws-iot-core)
  - [JSON Config File](#json-config-file)
- [Building from Source](#building-from-source)
- [Docker Usage](#docker-usage)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)

## Installation

### Precompiled Binaries (Coming Soon)
_Planned for future releases_.  

### From Source

1. Ensure you have **Go 1.20+** installed. 

2. Run:
   
        git clone https://github.com/miketigerblue/mqttcli.git
        cd mqttcli
        go build -o mqttcli ./cmd/mqttcli

Optional: move mqttcli to your $PATH:

        mv mqttcli /usr/local/bin/

You’re all set!

Quick Start

      ./mqttcli --broker "tcp://localhost:1883" \
            --clientid "testClient" \
            --topic "my/test/topic" \
            --qos 1

This subscribes to my/test/topic on a local broker using client ID testClient and QoS level 1.

## Configuration

You can configure mqttcli in two ways:

CLI Flags

    --broker        (string)  MQTT broker URL (e.g. "tcp://localhost:1883", "ssl://host:8883")
    --clientid      (string)  Unique MQTT client ID
    --username      (string)  MQTT username (optional)
    --password      (string)  MQTT password (optional)
    --topic         (string)  Topic to subscribe (and optionally publish) to
    --cafile        (string)  Path to CA certificate file
    --certfile      (string)  Path to client certificate
    --keyfile       (string)  Path to client key
    --qos           (int)     QoS level: 0, 1, or 2
    --insecure      (bool)    Skip server cert validation (NOT recommended)
    --quiet         (bool)    Suppress incoming message logs
    --verbose-errors (bool)   Print more detailed errors
    --config        (string)  Path to a JSON config file

JSON Config

    {
    "broker_url": "ssl://<aws-iot-endpoint>.amazonaws.com:8883",
    "client_id": "myAwsThing",
    "ca_file": "/path/to/AmazonRootCA1.pem",
    "cert_file": "/path/to/deviceCert.crt",
    "key_file": "/path/to/deviceKey.key",
    "topic": "iot/gnss/myThing/data",
    "qos": 1
    }

Invoke via --config /path/to/config.json.
CLI flags override any matching JSON fields.

## Examples

Basic Local Broker

    ./mqttcli \
        --broker "tcp://localhost:1883" \
        --clientid "localTest" \
        --topic "test/topic" \
        --qos 1

Connects via plain TCP (no TLS).
Subscribes to test/topic.


AWS IoT Core

    ./mqttcli \
        --broker "ssl://<your-endpoint>.amazonaws.com:8883" \
        --clientid "myThing" \
        --cafile "/path/to/AmazonRootCA1.pem" \
        --certfile "/path/to/deviceCert.crt" \
        --keyfile "/path/to/deviceKey.key" \
        --topic "iot/gnss/myThing/data" \
        --qos 1

Make sure your AWS IoT policy allows iot:Connect, iot:Publish, and (if needed) iot:Subscribe on the target topics.
This tool will subscribe to iot/gnss/myThing/data and log incoming messages.

JSON Config File

    {
    "broker_url": "ssl://<aws-iot-endpoint>.amazonaws.com:8883",
    "client_id": "myAwsThing",
    "ca_file": "/path/to/AmazonRootCA1.pem",
    "cert_file": "/path/to/deviceCert.crt",
    "key_file": "/path/to/deviceKey.key",
    "topic": "iot/gnss/myThing/data",
    "qos": 1
    }

## Usage:

    ./mqttcli --config config.json

You can still pass CLI flags like --qos 2 to override the JSON setting.

### Building from Source

Clone the repo:

    git clone https://github.com/miketigerblue/mqttcli.git
    cd mqttcli

Compile:

    go build -o mqttcli ./cmd/mqttcli

Test:

    go test ./...
    (Add tests in internal/, pkg/, or wherever appropriate.)

## Docker Usage

A multi-stage Dockerfile is included.

Build and run:

    docker build -t tigerblue/mqttcli .
    docker run --rm tigerblue/mqttcli \
        --broker "tcp://test.mosquitto.org:1883" \
        --topic "mqttcli/example" \
        --clientid "dockerClient"

For AWS IoT usage, you can mount CA/cert/key into the container as volumes and reference them with --cafile, --certfile, etc.

## Roadmap

 Publishing Support for sending messages (payload, intervals) from CLI.
 Multiple Subscriptions for different QoS levels.
 WebSocket connections, e.g., AWS IoT with SigV4 auth.
 Load Balanced / Cluster Mode for high-availability message bridging.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

Fork the project & create your feature branch: git checkout -b feature/my-new-feature
Commit your changes: git commit -am 'Add some feature'
Push to the branch: git push origin feature/my-new-feature
Create a new Pull Request

## License

This project is licensed under the Apache License 2.0