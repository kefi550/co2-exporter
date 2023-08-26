package main

import (
	"fmt"
	"net/http"

	"github.com/alecthomas/kingpin"
	mhz16 "github.com/kefi550/mh-z16-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type prometheusExporter struct {
	co2            *prometheus.Desc
	sensorPortName string
}

func NewPrometheusExporter(sensorPortName string) *prometheusExporter {
	return &prometheusExporter{
		co2: prometheus.NewDesc(
			"co2_ppm",
			"co2 concentration ppm",
			nil, nil,
		),
		sensorPortName: sensorPortName,
	}
}

func (pe *prometheusExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- pe.co2
}

func (pe *prometheusExporter) Collect(ch chan<- prometheus.Metric) {
	value, err := getCo2Command(pe.sensorPortName)
	if err != nil {
		fmt.Println(err)
		return
	}
	ch <- prometheus.MustNewConstMetric(pe.co2, prometheus.GaugeValue, value)
}

func getCo2Command(sensorPortName string) (float64, error) {
	sensor, err := mhz16.Open(sensorPortName)
	if err != nil {
		return 0, fmt.Errorf("cant open mhz16 %w", err)
	}
	defer sensor.Close()

	out, err := sensor.GetCo2()
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	result := float64(out)
	return result, nil
}

func main() {
	portName := kingpin.Arg("port", "portName").Required().ExistingFile()

	kingpin.Parse()

	registry := prometheus.NewRegistry()

	exporter := NewPrometheusExporter(*portName)
	registry.MustRegister(exporter)

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.ListenAndServe(":2112", nil)
}
