package main

import (
	"net/http"
	"time"

	"github.com/JackalLabs/jindexer/database"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// ReportRequest represents the request body for the /report endpoint
type ReportRequest struct {
	Merkles   []string `json:"merkles" binding:"required,min=1"`
	StartTime string   `json:"start_time" binding:"required"`
	EndTime   string   `json:"end_time" binding:"required"`
}

// ProofWindow represents a 12-hour window with proof status
type ProofWindow struct {
	Start         time.Time `json:"start"`
	End           time.Time `json:"end"`
	AllProven     bool      `json:"all_proven"`
	ProvenMerkles []string  `json:"proven_merkles"`
	MissedMerkles []string  `json:"missed_merkles"`
}

// ReportSummary contains aggregate statistics
type ReportSummary struct {
	TotalWindows       int `json:"total_windows"`
	FullyProvenWindows int `json:"fully_proven_windows"`
	MissedWindows      int `json:"missed_windows"`
}

// ReportResponse represents the response body for the /report endpoint
type ReportResponse struct {
	Merkles []string      `json:"merkles"`
	Windows []ProofWindow `json:"windows"`
	Summary ReportSummary `json:"summary"`
}

const windowDuration = 12 * time.Hour

// RegisterReportEndpoint adds the /report POST endpoint to the router
func RegisterReportEndpoint(r *gin.Engine, d *database.Database) {
	r.POST("/report", func(c *gin.Context) {
		var req ReportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Parse times
		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format, use RFC3339"})
			return
		}

		endTime, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format, use RFC3339"})
			return
		}

		if endTime.Before(startTime) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "end_time must be after start_time"})
			return
		}

		// Generate report
		response, err := generateReport(d, req.Merkles, startTime, endTime)
		if err != nil {
			log.Err(err).Msg("failed to generate report")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
			return
		}

		c.JSON(http.StatusOK, response)
	})
}

// generateReport creates the proof report by analyzing 12-hour windows
func generateReport(d *database.Database, merkles []string, startTime, endTime time.Time) (*ReportResponse, error) {
	// Build merkle proof map: merkle -> list of proof times
	merkleProofTimes := make(map[string][]time.Time)
	for _, merkle := range merkles {
		merkleProofTimes[merkle] = []time.Time{}
	}

	// Query proofs for each merkle
	for _, merkle := range merkles {
		proofs, err := d.ListProofsByMerkleAndTimeRange(merkle, startTime, endTime)
		if err != nil {
			return nil, err
		}
		for _, proof := range proofs {
			merkleProofTimes[merkle] = append(merkleProofTimes[merkle], proof.Block.Time)
		}
	}

	// Generate windows
	var windows []ProofWindow
	windowStart := startTime

	for windowStart.Before(endTime) {
		windowEnd := windowStart.Add(windowDuration)
		if windowEnd.After(endTime) {
			windowEnd = endTime
		}

		window := analyzeWindow(merkles, merkleProofTimes, windowStart, windowEnd)
		windows = append(windows, window)

		windowStart = windowEnd
	}

	// Calculate summary
	fullyProven := 0
	missed := 0
	for _, w := range windows {
		if w.AllProven {
			fullyProven++
		} else {
			missed++
		}
	}

	return &ReportResponse{
		Merkles: merkles,
		Windows: windows,
		Summary: ReportSummary{
			TotalWindows:       len(windows),
			FullyProvenWindows: fullyProven,
			MissedWindows:      missed,
		},
	}, nil
}

// analyzeWindow checks which merkles have proofs in the given time window
func analyzeWindow(merkles []string, merkleProofTimes map[string][]time.Time, windowStart, windowEnd time.Time) ProofWindow {
	var provenMerkles []string
	var missedMerkles []string

	for _, merkle := range merkles {
		proofTimes := merkleProofTimes[merkle]
		hasProofInWindow := false

		for _, proofTime := range proofTimes {
			if !proofTime.Before(windowStart) && proofTime.Before(windowEnd) {
				hasProofInWindow = true
				break
			}
		}

		if hasProofInWindow {
			provenMerkles = append(provenMerkles, merkle)
		} else {
			missedMerkles = append(missedMerkles, merkle)
		}
	}

	return ProofWindow{
		Start:         windowStart,
		End:           windowEnd,
		AllProven:     len(missedMerkles) == 0,
		ProvenMerkles: provenMerkles,
		MissedMerkles: missedMerkles,
	}
}

