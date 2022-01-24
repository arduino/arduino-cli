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

package metrics

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/arduino/arduino-cli/inventory"
	"github.com/segmentio/stats/v4"
	"github.com/segmentio/stats/v4/prometheus"
	"github.com/sirupsen/logrus"
)

// serverPattern is the metrics endpoint resource path for consume metrics
var serverPattern = "/metrics"

// Activate configures and starts the metrics server exposing a Prometheus resource
func Activate(metricPrefix, serverAddr string) {
	// Create a Prometheus default handler
	ph := prometheus.DefaultHandler
	// Create a new stats engine with an engine that prepends the "daemon" prefix to all metrics
	// and includes the installationID as a tag, then replace the default stats engine
	stats.DefaultEngine = stats.WithPrefix(metricPrefix, stats.T("installationID",
		inventory.Store.GetString("installation.id")))
	// Register the handler so it receives metrics from the default engine.
	stats.Register(ph)

	logrus.Infof("Setting up Prometheus metrics on %s%s", serverAddr, serverPattern)
	go func() {
		http.Handle(serverPattern, ph)
		logrus.Error(http.ListenAndServe(serverAddr, nil))
	}()

}

// Sanitize uses config generated UUID (installation.secret) as an HMAC secret to sanitize and anonymize
// a string, maintaining it distinguishable from a different string from the same Installation
func Sanitize(s string) string {
	logrus.Infof("inventory.Store.ConfigFileUsed() %s", inventory.Store.ConfigFileUsed())
	installationSecret := inventory.Store.GetString("installation.secret")
	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(installationSecret))
	// Write Data to it
	h.Write([]byte(s))
	// Get result and encode as hexadecimal string
	return hex.EncodeToString(h.Sum(nil))
}
