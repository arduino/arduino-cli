package telemetry

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"path/filepath"

	"github.com/segmentio/stats/v4"
	"github.com/segmentio/stats/v4/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Engine is the engine used by global helper functions for this module.
var Engine = stats.DefaultEngine

// Activate configure and starts the telemetry server exposing a Prometheus resource
func Activate(metricPrefix, installationID string) {
	// Configure telemetry engine
	// Create a Prometheus default handler
	ph := prometheus.DefaultHandler
	// Create a new stats engine with an engine that prepends the "daemon" prefix to all metrics
	// and includes the installationID as a tag
	Engine = stats.WithPrefix(metricPrefix, stats.T("installationID", installationID))
	// Register the handler so it receives metrics from the default engine.
	Engine.Register(ph)

	// Configure using viper settings
	serverAddr := viper.GetString("telemetry.addr")
	serverPattern := viper.GetString("telemetry.pattern")
	logrus.Infof("Setting up Prometheus telemetry on %s%s", serverAddr, serverPattern)
	go func() {
		http.Handle(serverPattern, ph)
		logrus.Error(http.ListenAndServe(serverAddr, nil))
	}()

}

// SanitizeSketchPath uses config generated UUID (installation.secret) as an HMAC secret to sanitize and anonymize the sketch
// name maintaining it distinguishable from a different sketch from the same Installation
func SanitizeSketchPath(sketchPath string) string {
	installationSecret := viper.GetString("installation.secret")
	sketchName := filepath.Base(sketchPath)
	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(installationSecret))
	// Write Data to it
	h.Write([]byte(sketchName))
	// Get result and encode as hexadecimal string
	return hex.EncodeToString(h.Sum(nil))
}
