package competitive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/competitive/models"
)

// ShoppingGraphClient queries the Shopping Graph for competitor pricing data.
type ShoppingGraphClient struct {
	baseURL string
	client  *http.Client
	cache   *priceCache
}

// NewShoppingGraphClient creates a new Shopping Graph HTTP client.
//
// Parameters:
//   - baseURL: Shopping Graph base URL (e.g., "http://localhost:9000")
func NewShoppingGraphClient(baseURL string) *ShoppingGraphClient {
	return &ShoppingGraphClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 1 * time.Second,
		},
		cache: newPriceCache(10 * time.Second),
	}
}

// GetLowestPrice returns the lowest competitor price for a product.
// Implements CompetitorPriceSource interface.
func (c *ShoppingGraphClient) GetLowestPrice(productID string) (price int, merchantID string, err error) {
	// Check cache first
	if cached, ok := c.cache.get(productID); ok {
		return cached.Price, cached.MerchantID, nil
	}

	// Search by product ID
	results, err := c.search(productID, 10)
	if err != nil {
		return 0, "", fmt.Errorf("shopping graph search failed: %w", err)
	}

	if len(results) == 0 {
		return 0, "", fmt.Errorf("no results found for product %s", productID)
	}

	// Find lowest price among in-stock items
	lowestPrice := -1
	lowestMerchant := ""

	for _, result := range results {
		if !result.InStock {
			continue // Skip out-of-stock items
		}

		if lowestPrice == -1 || result.Price < lowestPrice {
			lowestPrice = result.Price
			lowestMerchant = result.MerchantID
		}
	}

	if lowestPrice == -1 {
		return 0, "", fmt.Errorf("no in-stock results for product %s", productID)
	}

	// Cache the result
	c.cache.set(productID, cachedPrice{
		Price:      lowestPrice,
		MerchantID: lowestMerchant,
		Timestamp:  time.Now(),
	})

	return lowestPrice, lowestMerchant, nil
}

// GetCompetitorPrices returns all competitor prices for a product.
// Implements the new models.CompetitorPriceSource interface.
func (c *ShoppingGraphClient) GetCompetitorPrices(productID string) ([]models.CompetitorPrice, error) {
	log.Printf("[DEBUG ShoppingGraph] Searching for product: %s", productID)
	results, err := c.search(productID, 50)
	if err != nil {
		log.Printf("[DEBUG ShoppingGraph] Search FAILED: %v", err)
		return nil, fmt.Errorf("shopping graph search failed: %w", err)
	}

	log.Printf("[DEBUG ShoppingGraph] Got %d results from search", len(results))

	prices := make([]models.CompetitorPrice, 0, len(results))
	now := time.Now()
	for i, result := range results {
		log.Printf("[DEBUG ShoppingGraph] Result %d: MerchantID=%s, MerchantName=%s, Price=%d, InStock=%v, DiscountHints=%v",
			i, result.MerchantID, result.MerchantName, result.Price, result.InStock, result.DiscountHints)

		// Estimate effective price after best discount
		effectivePrice := estimateEffectivePrice(result.Price, result.DiscountHints)

		log.Printf("[DEBUG ShoppingGraph] Effective price: %d (discount hints applied)", effectivePrice)

		prices = append(prices, models.CompetitorPrice{
			MerchantID:     result.MerchantID,
			MerchantName:   result.MerchantName,
			Price:          result.Price,
			InStock:        result.InStock,
			Timestamp:      now,
			DiscountHints:  result.DiscountHints,
			EffectivePrice: effectivePrice,
		})
	}

	log.Printf("[DEBUG ShoppingGraph] Returning %d competitor prices", len(prices))
	return prices, nil
}

// GetCompetitorPricesLegacy returns all competitor prices using the legacy type.
// This is for backward compatibility with the old CompetitivePricingAgent.
func (c *ShoppingGraphClient) GetCompetitorPricesLegacy(productID string) ([]CompetitorPrice, error) {
	results, err := c.search(productID, 50)
	if err != nil {
		return nil, fmt.Errorf("shopping graph search failed: %w", err)
	}

	prices := make([]CompetitorPrice, 0, len(results))
	for _, result := range results {
		prices = append(prices, CompetitorPrice{
			MerchantID:   result.MerchantID,
			MerchantName: result.MerchantName,
			ProductID:    result.ProductID,
			Price:        result.Price,
			InStock:      result.InStock,
		})
	}

	return prices, nil
}

// search performs a Shopping Graph search query.
func (c *ShoppingGraphClient) search(query string, limit int) ([]SearchResult, error) {
	reqBody := map[string]interface{}{
		"query": query,
		"limit": limit,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + "/search"
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Results []SearchResult `json:"results"`
		Total   int            `json:"total"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Results, nil
}

// priceCache provides simple in-memory caching of price lookups.
type priceCache struct {
	mu      sync.RWMutex
	entries map[string]cachedPrice
	ttl     time.Duration
}

type cachedPrice struct {
	Price      int
	MerchantID string
	Timestamp  time.Time
}

func newPriceCache(ttl time.Duration) *priceCache {
	return &priceCache{
		entries: make(map[string]cachedPrice),
		ttl:     ttl,
	}
}

func (pc *priceCache) get(productID string) (cachedPrice, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	entry, ok := pc.entries[productID]
	if !ok {
		return cachedPrice{}, false
	}

	// Check if expired
	if time.Since(entry.Timestamp) > pc.ttl {
		return cachedPrice{}, false
	}

	return entry, true
}

func (pc *priceCache) set(productID string, price cachedPrice) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.entries[productID] = price
}

func (pc *priceCache) clear() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.entries = make(map[string]cachedPrice)
}

// estimateEffectivePrice estimates the final price after applying the best discount hint.
// Parses discount code patterns to estimate percentage/fixed discounts.
func estimateEffectivePrice(basePrice int, discountHints []string) int {
	if len(discountHints) == 0 {
		return basePrice
	}

	bestPrice := basePrice

	for _, code := range discountHints {
		estimatedPrice := basePrice

		// Pattern matching for common discount code formats:
		// - WELCOME10, SAVE20, etc. → percentage discount
		// - FIXED500 → $5 fixed discount (500 cents)
		// - 10OFF, 20OFF → percentage discount

		// Try to extract percentage from code name
		var percent int
		var fixed int

		// Check for patterns like "WELCOME10", "SAVE20", "10OFF"
		if len(code) >= 2 {
			lastTwo := code[len(code)-2:]
			if lastTwo[0] >= '0' && lastTwo[0] <= '9' && lastTwo[1] >= '0' && lastTwo[1] <= '9' {
				// Two digits at the end, likely a percentage
				percent = int(lastTwo[0]-'0')*10 + int(lastTwo[1]-'0')
				log.Printf("[DEBUG estimateEffectivePrice] Code %s → estimated %d%% discount", code, percent)
			} else if lastTwo[1] >= '0' && lastTwo[1] <= '9' {
				// One digit at the end
				percent = int(lastTwo[1] - '0')
				log.Printf("[DEBUG estimateEffectivePrice] Code %s → estimated %d%% discount", code, percent)
			}
		}

		// Check for FIXED pattern (e.g., FIXED500 = $5 off)
		if len(code) > 5 && code[:5] == "FIXED" {
			// Parse remaining digits
			fixedStr := code[5:]
			fixed = 0
			for _, ch := range fixedStr {
				if ch >= '0' && ch <= '9' {
					fixed = fixed*10 + int(ch-'0')
				}
			}
			log.Printf("[DEBUG estimateEffectivePrice] Code %s → estimated $%.2f fixed discount", code, float64(fixed)/100)
		}

		// Apply the discount
		if percent > 0 {
			estimatedPrice = basePrice * (100 - percent) / 100
		} else if fixed > 0 {
			estimatedPrice = basePrice - fixed
		} else {
			// Unknown pattern, assume 10% default
			estimatedPrice = basePrice * 90 / 100
			log.Printf("[DEBUG estimateEffectivePrice] Code %s → unknown pattern, assuming 10%% discount", code)
		}

		// Keep track of best (lowest) price
		if estimatedPrice < bestPrice {
			bestPrice = estimatedPrice
		}
	}

	return bestPrice
}
