package telemetry

import (
	"net/http"

	"github.com/segmentio/stats/v4"
	"github.com/segmentio/stats/v4/prometheus"
	"github.com/sirupsen/logrus"
)

// Engine is the engine used by global helper functions for this module.
var Engine = stats.DefaultEngine

var serverAddr = ":2112"
var serverPattern = "/metrics"

//Activate configure and starts the telemetry server exposing a Prometheus resource
func Activate(metricPrefix string) {
	// Configure telemetry engine
	// Create a Prometheus default handler
	ph := prometheus.DefaultHandler
	// Replace the default stats engine with an engine that prepends the "daemon" prefix to all metrics
	Engine = stats.WithPrefix(metricPrefix)
	// Register the handler so it receives metrics from the default engine.
	Engine.Register(ph)
	// Flush the default stats engine on return to ensure all buffered
	// metrics are sent to the server.
	defer Engine.Flush()
	// move everything inside commands and search for setting up a common prefix for all metrics sent!
	logrus.Infof("Setting up Prometheus telemetry on %s%s", serverAddr, serverPattern)
	go func() {
		http.Handle(serverPattern, ph)
		logrus.Error(http.ListenAndServe(serverAddr, nil))
	}()

}
