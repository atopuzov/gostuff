package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/exp/io/i2c"
	"github.com/quhar/bme280"
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

func readTemperatureandHumidityPresure() (float64, float64, float64) {
	d, err := i2c.Open(&i2c.Devfs{Dev: "/dev/i2c-1"}, 0x76)
	
	if err != nil {
		log.Fatal(err)
	}

	b := bme280.New(d)
	err = b.Init()
	if err != nil {
		log.Fatal(err)
	}

	temperature, presure, humidity, err := b.EnvData()
	fmt.Printf("Sensor = BME280: Temperature = %v*C, Humidity = %v%%, Presure = %vhPa\n",
		temperature, humidity, presure)
		
	return temperature, humidity, presure
}

func createMetrics(c client.Client) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
		Precision: "s",
	})

	if err != nil {
		log.Fatalln("Error: ", err)
	}

	temperature, humidity, presure := readTemperatureandHumidityPresure()

	fields := map[string]interface{}{
		"temperature": temperature,
		"humidity":    humidity,
		"presure":     presure,
	}

	tags := map[string]string{
		"host": "localhost",
	}

	pt, err := client.NewPoint("bme280", tags, fields, time.Now())
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
