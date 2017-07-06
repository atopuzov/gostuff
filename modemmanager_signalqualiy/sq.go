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
	"log"
	"time"

	"github.com/guelfey/go.dbus"
	"github.com/influxdata/influxdb/client/v2"
)

const (
	mm_bus_name    = "org.freedesktop.ModemManager1"
	mm_object_path = "/org/freedesktop/ModemManager1/Modem/0"
	mm_iface       = "org.freedesktop.ModemManager1.Modem.SignalQuality"
	database       = "modem"
	username       = "admin"
	password       = "admin"
)

func influxDBClient() client.Client {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://localhost:8086",
		Username: username,
		Password: password,
	})

	if err != nil {
		log.Fatal(err)
	}

	return c
}

func readSignalQuality() uint32 {
	conn, err := dbus.SystemBus()

	if err != nil {
		log.Fatal(err)
	}

	busObject := conn.Object(mm_bus_name, mm_object_path)
	prop, err := busObject.GetProperty(mm_iface)

	if err != nil {
		log.Fatal(err)
	}

	value := prop.Value().([]interface{})
	quality := value[0]

	return quality.(uint32)
}

func createMetrics(c client.Client) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
		Precision: "s",
	})

	if err != nil {
		log.Fatal(err)
	}

	signalquality := readSignalQuality()

	fields := map[string]interface{}{
		"signalquality": signalquality,
	}

	tags := map[string]string{
		"host": "localhost",
	}

	pt, err := client.NewPoint("E3276", tags, fields, time.Now())
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
