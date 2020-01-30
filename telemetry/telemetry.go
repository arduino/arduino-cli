// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package telemetry

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"path/filepath"

	"github.com/arduino/arduino-cli/repertory"
	"github.com/segmentio/stats/v4"
	"github.com/segmentio/stats/v4/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Activate configures and starts the telemetry server exposing a Prometheus resource
func Activate(metricPrefix string) {
	// Create a Prometheus default handler
	ph := prometheus.DefaultHandler
	// Create a new stats engine with an engine that prepends the "daemon" prefix to all metrics
	// and includes the installationID as a tag, then replace the default stats engine
	stats.DefaultEngine = stats.WithPrefix(metricPrefix, stats.T("installationID",
		repertory.Store.GetString("installation.id")))
	// Register the handler so it receives metrics from the default engine.
	stats.Register(ph)

	// Configure using viper settings
	serverAddr := viper.GetString("telemetry.addr")
	serverPattern := viper.GetString("telemetry.pattern")
	logrus.Infof("Setting up Prometheus telemetry on %s%s", serverAddr, serverPattern)
	go func() {
		http.Handle(serverPattern, ph)
		logrus.Error(http.ListenAndServe(serverAddr, nil))
	}()

}

// SanitizeSketchPath uses config generated UUID (installation.secret) as an HMAC secret to sanitize and anonymize
// the sketch name maintaining it distinguishable from a different sketch from the same Installation
func SanitizeSketchPath(sketchPath string) string {
	logrus.Infof("repertory.Store.ConfigFileUsed() %s", repertory.Store.ConfigFileUsed())
	installationSecret := repertory.Store.GetString("installation.secret")
	sketchName := filepath.Base(sketchPath)
	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(installationSecret))
	// Write Data to it
	h.Write([]byte(sketchName))
	// Get result and encode as hexadecimal string
	return hex.EncodeToString(h.Sum(nil))
}
