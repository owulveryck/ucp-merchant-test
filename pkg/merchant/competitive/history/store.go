// Package history provides price history storage and trend analysis.
package history

import (
	"math"
	"sort"
	"sync"
	"time"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/models"
)

// InMemoryHistoryStore stores price history in memory.
type InMemoryHistoryStore struct {
	mu     sync.RWMutex
	prices map[string][]models.PricePoint // productID -> []PricePoint
}

// NewInMemoryHistoryStore creates a new in-memory history store.
func NewInMemoryHistoryStore() *InMemoryHistoryStore {
	return &InMemoryHistoryStore{
		prices: make(map[string][]models.PricePoint),
	}
}

// RecordPrice records a price observation.
func (s *InMemoryHistoryStore) RecordPrice(productID string, price int, timestamp time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	point := models.PricePoint{
		ProductID: productID,
		Price:     price,
		Timestamp: timestamp,
	}

	s.prices[productID] = append(s.prices[productID], point)

	// Keep only last 1000 points per product
	if len(s.prices[productID]) > 1000 {
		s.prices[productID] = s.prices[productID][len(s.prices[productID])-1000:]
	}

	return nil
}

// GetPriceHistory returns recent price history.
func (s *InMemoryHistoryStore) GetPriceHistory(productID string, limit int) ([]models.PricePoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	points, ok := s.prices[productID]
	if !ok || len(points) == 0 {
		return []models.PricePoint{}, nil
	}

	// Return last N points
	start := 0
	if len(points) > limit {
		start = len(points) - limit
	}

	result := make([]models.PricePoint, len(points)-start)
	copy(result, points[start:])

	return result, nil
}

// GetTrend analyzes price trend over a duration.
func (s *InMemoryHistoryStore) GetTrend(productID string, duration time.Duration) (models.Trend, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	points, ok := s.prices[productID]
	if !ok || len(points) == 0 {
		return models.Trend{
			Direction:  "stable",
			Period:     duration,
			DataPoints: 0,
		}, nil
	}

	// Filter points within duration
	cutoff := time.Now().Add(-duration)
	var relevant []models.PricePoint
	for _, p := range points {
		if p.Timestamp.After(cutoff) {
			relevant = append(relevant, p)
		}
	}

	if len(relevant) < 2 {
		return models.Trend{
			Direction:  "stable",
			Period:     duration,
			DataPoints: len(relevant),
		}, nil
	}

	// Sort by timestamp
	sort.Slice(relevant, func(i, j int) bool {
		return relevant[i].Timestamp.Before(relevant[j].Timestamp)
	})

	// Calculate trend
	firstPrice := float64(relevant[0].Price)
	lastPrice := float64(relevant[len(relevant)-1].Price)
	percentChange := ((lastPrice - firstPrice) / firstPrice) * 100

	// Calculate volatility (standard deviation)
	volatility := calculateVolatility(relevant)

	// Determine direction
	direction := "stable"
	if percentChange > 2 {
		direction = "up"
	} else if percentChange < -2 {
		direction = "down"
	}

	return models.Trend{
		Direction:     direction,
		PercentChange: percentChange,
		Period:        duration,
		DataPoints:    len(relevant),
		Volatility:    volatility,
	}, nil
}

// calculateVolatility calculates price volatility (coefficient of variation).
func calculateVolatility(points []models.PricePoint) float64 {
	if len(points) < 2 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for _, p := range points {
		sum += float64(p.Price)
	}
	mean := sum / float64(len(points))

	// Calculate standard deviation
	sumSquaredDiff := 0.0
	for _, p := range points {
		diff := float64(p.Price) - mean
		sumSquaredDiff += diff * diff
	}
	stdDev := math.Sqrt(sumSquaredDiff / float64(len(points)))

	// Return coefficient of variation (stdDev / mean)
	if mean == 0 {
		return 0
	}
	return (stdDev / mean) * 100
}
