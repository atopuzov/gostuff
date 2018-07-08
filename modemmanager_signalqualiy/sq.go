// -*- coding: utf-8 -*-
// The MIT License (MIT)

// Copyright (c) 2014 Aleksandar Topuzović <aleksandar.topuzovic@gmail.com>

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"flag"
	"log"
	"time"

	"github.com/guelfey/go.dbus"
	InfluxDB "github.com/influxdata/influxdb/client/v2"
)

const (
	modemManager  = "org.freedesktop.ModemManager1"
	signalQuality = "org.freedesktop.ModemManager1.Modem.SignalQuality"
	getObjects    = "org.freedesktop.DBus.ObjectManager.GetManagedObjects"
)

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

	pt, err := InfluxDB.NewPoint("E3276", tags, fields, time.Now())
	if err != nil {
		log.Fatalln("Unable to create new point: ", err)
	}

	bp.AddPoint(pt)

	err = c.Write(bp)
	if err != nil {
		log.Fatalln("Unable to write the metric: ", err)
	}
}

func readSignalQuality(conn *dbus.Conn, modemPath dbus.ObjectPath) uint32 {
	busObject := conn.Object(modemManager, modemPath)
	prop, err := busObject.GetProperty(signalQuality)

	if err != nil {
		log.Fatal(err)
	}

	value := prop.Value().([]interface{})
	quality := value[0]

	return quality.(uint32)
}

func publishModemSQ(influx InfluxDB.Client, dbname string) {
	conn, err := dbus.SystemBus()

	if err != nil {
		log.Fatal(err)
	}

	busObject := conn.Object(modemManager, "/org/freedesktop/ModemManager1")

	objects := map[dbus.ObjectPath]map[string]map[string]dbus.Variant{}
	err = busObject.Call(getObjects, 0).Store(&objects)
	if err != nil {
		log.Fatal(err)
	}

	for objPath, _ := range objects {
		signalquality := readSignalQuality(conn, objPath)
		log.Printf("%s: %d", objPath, signalquality)

		fields := map[string]interface{}{
			"signalquality": signalquality,
		}

		tags := map[string]string{
			"host": "localhost",
		}
		publishMetric(influx, dbname, tags, fields)
	}
}

func main() {
	dbserver := flag.String("db-server", "http://localhost:8086", "InfluxDB server")
	dbuser := flag.String("db-user", "admin", "InfluxDB username")
	dbpass := flag.String("db-pass", "admin", "InfluxDB password")
	dbname := flag.String("db-name", "modem", "InfluxDB database")
	flag.Parse()

	influx := influxDBClient(*dbserver, *dbuser, *dbpass)

	publishModemSQ(influx, *dbname)
}
