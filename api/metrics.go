package main

import (
	"math"
	"time"

	"github.com/JackalLabs/jindexer/database"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

const (
	// Proof window duration in seconds (12 hours)
	proofWindowSeconds = 12 * 60 * 60
	// Critical threshold in seconds (24 hours)
	criticalWindowSeconds = 24 * 60 * 60
)

var (
	// Aggregate metrics - these don't have high cardinality labels

	// TotalMerklesTracked is the total number of unique merkles in the database
	TotalMerklesTracked = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "jindexer_merkles_total",
			Help: "Total number of unique merkles being tracked",
		},
	)

	// TotalProofsIndexed is the total number of proofs in the database
	TotalProofsIndexed = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "jindexer_proofs_total",
			Help: "Total number of proofs indexed in the database",
		},
	)

	// MerklesHealthy is the count of merkles with proofs within the 12h window
	MerklesHealthy = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "jindexer_merkles_healthy",
			Help: "Number of merkles with proofs within the 12-hour window",
		},
	)

	// MerklesMissed is the count of merkles that have missed the 12h proof window
	MerklesMissed = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "jindexer_merkles_missed",
			Help: "Number of merkles that have missed the 12-hour proof window",
		},
	)

	// MerklesCritical is the count of merkles that have missed the 24h window
	MerklesCritical = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "jindexer_merkles_critical",
			Help: "Number of merkles that have missed the 24-hour proof window (critical)",
		},
	)

	// ProofAgeHistogram tracks the distribution of proof ages
	ProofAgeHistogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "jindexer_proof_age_seconds",
			Help:    "Histogram of proof ages in seconds",
			Buckets: []float64{3600, 7200, 10800, 14400, 21600, 43200, 64800, 86400, 172800}, // 1h, 2h, 3h, 4h, 6h, 12h, 18h, 24h, 48h
		},
	)

	// OldestProofAge tracks the age of the oldest proof (worst case)
	OldestProofAge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "jindexer_oldest_proof_age_seconds",
			Help: "Age of the oldest proof in seconds (worst case merkle)",
		},
	)

	// NewestProofAge tracks the age of the newest proof
	NewestProofAge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "jindexer_newest_proof_age_seconds",
			Help: "Age of the most recent proof in seconds",
		},
	)
)

func init() {
	// Register aggregate metrics with Prometheus
	prometheus.MustRegister(TotalMerklesTracked)
	prometheus.MustRegister(TotalProofsIndexed)
	prometheus.MustRegister(MerklesHealthy)
	prometheus.MustRegister(MerklesMissed)
	prometheus.MustRegister(MerklesCritical)
	prometheus.MustRegister(ProofAgeHistogram)
	prometheus.MustRegister(OldestProofAge)
	prometheus.MustRegister(NewestProofAge)
}

// RegisterMetricsEndpoint adds the /metrics endpoint to the router
func RegisterMetricsEndpoint(r *gin.Engine) {
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

// InitializeMetricsFromDatabase loads existing proof data into Prometheus metrics
// This is called at startup to populate the metrics with historical data
// It also starts a background goroutine to periodically refresh metrics
func InitializeMetricsFromDatabase(d *database.Database) {
	log.Info().Msg("Initializing Prometheus metrics from database...")

	// Initial load
	refreshMetricsFromDatabase(d)

	// Start background refresh goroutine
	// Since indexer and API are separate containers, we need to poll the database
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			refreshMetricsFromDatabase(d)
		}
	}()

	log.Info().Msg("Prometheus metrics refresh goroutine started (30s interval)")
}

// refreshMetricsFromDatabase queries the database and computes aggregate metrics
func refreshMetricsFromDatabase(d *database.Database) {
	// Use SQL aggregates to get the latest proof time per merkle instead of
	// loading individual rows into Go memory.
	merkleProofs, err := d.GetMerkleLastProofTimes()
	if err != nil {
		log.Err(err).Msg("failed to refresh metrics from database")
		return
	}

	totalProofs, err := d.GetTotalProofCount()
	if err != nil {
		log.Err(err).Msg("failed to get total proof count")
		return
	}

	now := time.Now().Unix()

	totalMerkles := len(merkleProofs)
	var healthy, missed, critical int
	var oldestAge, newestAge int64 = 0, math.MaxInt64

	for _, mp := range merkleProofs {
		age := now - mp.LastProofTime.Unix()

		ProofAgeHistogram.Observe(float64(age))

		if age > oldestAge {
			oldestAge = age
		}
		if age < newestAge {
			newestAge = age
		}

		if age <= proofWindowSeconds {
			healthy++
		} else if age <= criticalWindowSeconds {
			missed++
		} else {
			critical++
		}
	}

	if totalMerkles == 0 {
		newestAge = 0
	}

	TotalMerklesTracked.Set(float64(totalMerkles))
	TotalProofsIndexed.Set(float64(totalProofs))
	MerklesHealthy.Set(float64(healthy))
	MerklesMissed.Set(float64(missed))
	MerklesCritical.Set(float64(critical))
	OldestProofAge.Set(float64(oldestAge))
	NewestProofAge.Set(float64(newestAge))

	log.Debug().
		Int("total_merkles", totalMerkles).
		Int("healthy", healthy).
		Int("missed", missed).
		Int("critical", critical).
		Msg("refreshed proof metrics")
}
