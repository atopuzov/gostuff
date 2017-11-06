package main

import (
	"fmt"
	"log"
	"time"

	"github.com/d2r2/go-dht"
	"github.com/influxdata/influxdb/client/v2"
)

const (
	database = "temperature"
	username = "admin"
	password = "admin"
)

func influxDBClient() client.Client {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://localhost:8086",
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalln("Error: ", err)
	}
	return c
}

func readTemperatureandHumidity() (float32, float32) {
	sensorType := dht.DHT22
	temperature, humidity, retried, err :=
		dht.ReadDHTxxWithRetry(sensorType, 10, false, 10)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Sensor = %v: Temperature = %v*C, Humidity = %v%% (retried %d times)\n",
		sensorType, temperature, humidity, retried)
	return temperature, humidity
}

func createMetrics(c client.Client) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
		Precision: "s",
	})

	if err != nil {
		log.Fatalln("Error: ", err)
	}

	temperature, humidity := readTemperatureandHumidity()

	fields := map[string]interface{}{
		"temperature": temperature,
		"humidity":    humidity,
	}

	tags := map[string]string{
		"host": "localhost",
	}

	pt, err := client.NewPoint("dht22", tags, fields, time.Now())
	if err != nil {
		log.Fatal(err)
	}

	bp.AddPoint(pt)

	err = c.Write(bp)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	c := influxDBClient()
	createMetrics(c)
}
