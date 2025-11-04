package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/JackalLabs/jindexer/database"
	"github.com/JackalLabs/jindexer/utils"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func main() {
	utils.InitLogger("Starting API")

	d, err := database.NewDatabase()
	if err != nil {
		panic(err)
	}

	// Initialize provider cache
	providerCache := NewProviderCache()

	// Set up Gin router
	r := gin.Default()

	// Query endpoint for proofs by merkle and date range
	r.GET("/query", func(c *gin.Context) {
		merkle := c.Query("merkle")
		if merkle == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "merkle parameter is required"})
			return
		}

		// Parse optional start and end dates, default to 30 days from current time
		now := time.Now()
		endTime := now
		startTime := now.AddDate(0, 0, -30)

		if startDateStr := c.Query("start_date"); startDateStr != "" {
			parsedStart, err := time.Parse(time.RFC3339, startDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use RFC3339 (e.g., 2006-01-02T15:04:05Z07:00)"})
				return
			}
			startTime = parsedStart
		}

		if endDateStr := c.Query("end_date"); endDateStr != "" {
			parsedEnd, err := time.Parse(time.RFC3339, endDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use RFC3339 (e.g., 2006-01-02T15:04:05Z07:00)"})
				return
			}
			endTime = parsedEnd
		}

		// Get proofs from database
		proofs, err := d.ListProofsByMerkleAndTimeRange(merkle, startTime, endTime)
		if err != nil {
			log.Err(err).Msg("failed to query proofs")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query database"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"merkle":     merkle,
			"start_date": startTime,
			"end_date":   endTime,
			"proofs":     proofs,
			"count":      len(proofs),
		})
	})

	// Recent proofs endpoint - lists most recent proofs ordered by block date with a limit
	r.GET("/recent", func(c *gin.Context) {
		// Parse limit parameter, default to 100 if not provided
		limitStr := c.DefaultQuery("limit", "100")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter, must be a positive integer"})
			return
		}

		// Get recent proofs from database
		proofs, err := d.ListRecentProofs(limit)
		if err != nil {
			log.Err(err).Msg("failed to query recent proofs")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query database"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"limit":  limit,
			"proofs": proofs,
			"count":  len(proofs),
		})
	})

	// Proofs endpoint - lists proofs ordered by ID (most recent first) with a limit
	r.GET("/proofs", func(c *gin.Context) {
		// Parse limit parameter, default to 100 if not provided
		limitStr := c.DefaultQuery("limit", "100")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter, must be a positive integer"})
			return
		}

		// Get proofs from database ordered by ID
		proofs, err := d.ListProofsByID(limit)
		if err != nil {
			log.Err(err).Msg("failed to query proofs")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query database"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"limit":  limit,
			"proofs": proofs,
			"count":  len(proofs),
		})
	})

	// Provider endpoint - returns the IP/domain for a given Jackal address
	r.GET("/provider/:address", func(c *gin.Context) {
		address := c.Param("address")
		if address == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "address parameter is required"})
			return
		}

		ip, err := providerCache.GetProviderIP(address)
		if err != nil {
			log.Err(err).Str("address", address).Msg("failed to get provider IP")
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "provider not found",
				"address": address,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"address": address,
			"ip":      ip,
		})
	})

	log.Info().Msg("Starting API server on :9797")
	if err := r.Run(":9797"); err != nil {
		panic(err)
	}
}
