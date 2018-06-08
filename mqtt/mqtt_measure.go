package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	InfluxDB "github.com/influxdata/influxdb/client/v2"
)

type Measurement struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Client      string  `json:"client"`
}

func influxDBClient(server string, username string, password string) InfluxDB.Client {
	c, err := InfluxDB.NewHTTPClient(InfluxDB.HTTPConfig{
		Addr:     server,
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalln("Unable to connect to InfluxDB: ", err)
	}
	return c
}

func publishMetric(c InfluxDB.Client, database string, tags map[string]string, fields map[string]interface{}) {
	bp, err := InfluxDB.NewBatchPoints(InfluxDB.BatchPointsConfig{
		Database:  database,
		Precision: "s",
	})

	if err != nil {
		log.Fatalln("Unable to create batch points: ", err)
	}

	pt, err := InfluxDB.NewPoint("sht30", tags, fields, time.Now())
	if err != nil {
		log.Fatalln("Unable to create new point: ", err)
	}

	bp.AddPoint(pt)

	err = c.Write(bp)
	if err != nil {
		log.Fatalln("Unable to write the metric: ", err)
	}
}

func main() {
	mqttserver := flag.String("mqtt-server", "tcp://127.0.0.1:1883", "The full URL of the MQTT server to connect to")
	mqtttopic := flag.String("mqtt-topic", "outTopic", "Topic on which to listen for messages")
	dbserver := flag.String("db-server", "http://localhost:8086", "InfluxDB server")
	dbuser := flag.String("db-user", "admin", "InfluxDB username")
	dbpass := flag.String("db-pass", "admin", "InfluxDB password")
	dbname := flag.String("db-name", "temperature", "InfluxDB database")
	flag.Parse()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	influx := influxDBClient(*dbserver, *dbuser, *dbpass)

	opts := MQTT.NewClientOptions().AddBroker(*mqttserver)
	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	opts.SetTLSConfig(tlsConfig)

	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(*mqtttopic, 0,
			func(client MQTT.Client, message MQTT.Message) {
				var measurement = Measurement{}
				if err := json.Unmarshal(message.Payload(), &measurement); err != nil {
					log.Printf("Unable to decode message '%s': %s\n", message.Payload(), err)
				}
				fmt.Println(measurement)

				fields := map[string]interface{}{
					"temperature": measurement.Temperature,
					"humidity":    measurement.Humidity,
				}

				tags := map[string]string{
					"client": measurement.Client,
				}

				publishMetric(influx, *dbname, tags, fields)
			}); token.Wait() && token.Error() != nil {
			log.Fatalln("Unable to subscribe to MQTT topic: ", token.Error())
		}
	}

	client := MQTT.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalln("Unable to connect to the MQT server:", token.Error())
	} else {
		fmt.Println("Connected...")
	}

	<-c
}
