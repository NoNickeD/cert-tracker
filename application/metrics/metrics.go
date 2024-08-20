package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

var (
	certExpiryDays = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cert_expiry_days",
			Help: "Days until SSL certificate expires",
		},
		[]string{"domain", "expiry_date", "issued_date", "ca", "http_status", "error"}, // Removed `days_remaining` from labels
	)
)

func init() {
	// Register the custom metrics with Prometheus' default registry
	prometheus.MustRegister(certExpiryDays)
}

// UpdateCertMetrics updates the Prometheus metrics for certificate information
func UpdateCertMetrics(domain, expiryDate, issuedDate string, daysRemaining int, ca, httpStatus, error string) {
	// Log metric update information for debugging purposes
	log.Printf("Updating metrics for domain: %s, expiry: %s, issued: %s, days remaining: %d, CA: %s, status: %s, error: %s",
		domain, expiryDate, issuedDate, daysRemaining, ca, httpStatus, error)

	certExpiryDays.With(prometheus.Labels{
		"domain":      domain,
		"expiry_date": expiryDate,
		"issued_date": issuedDate,
		"ca":          ca,
		"http_status": httpStatus,
		"error":       error,
	}).Set(float64(daysRemaining)) // Set the value of the metric to `daysRemaining`
}

// PrometheusHandler returns an HTTP handler for serving Prometheus metrics
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}
