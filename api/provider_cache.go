package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/rs/zerolog/log"
)

type ProviderCache struct {
	mu     sync.RWMutex
	cache  map[string]string // address -> IP
	apiURL string
}

type Provider struct {
	Address string `json:"address"`
	IP      string `json:"ip"`
}

type ProviderResponse struct {
	Provider Provider `json:"provider"`
}

func NewProviderCache() *ProviderCache {
	return &ProviderCache{
		cache:  make(map[string]string),
		apiURL: "https://api.jackalprotocol.com/jackal/canine-chain/storage/providers",
	}
}

// GetProviderIP returns the IP/domain for a given Jackal address.
// If not in cache, it fetches from the API and updates the cache.
func (pc *ProviderCache) GetProviderIP(address string) (string, error) {
	// Check cache first
	pc.mu.RLock()
	if ip, found := pc.cache[address]; found {
		pc.mu.RUnlock()
		return ip, nil
	}
	pc.mu.RUnlock()

	// Not in cache, fetch from API
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine might have updated it)
	if ip, found := pc.cache[address]; found {
		return ip, nil
	}

	// Fetch provider from API by address
	provider, err := pc.fetchProviderByAddress(address)
	if err != nil {
		return "", fmt.Errorf("failed to fetch provider: %w", err)
	}

	// Update cache with fetched provider
	pc.cache[provider.Address] = provider.IP

	return provider.IP, nil
}

// fetchProviderByAddress fetches a single provider from the API by address.
func (pc *ProviderCache) fetchProviderByAddress(address string) (*Provider, error) {
	reqURL := fmt.Sprintf("%s/%s", pc.apiURL, address)

	log.Info().Str("url", reqURL).Str("address", address).Msg("Fetching provider from Jackal API")

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("provider not found: %s", address)
		}
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var providerResp ProviderResponse
	if err := json.NewDecoder(resp.Body).Decode(&providerResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	provider := providerResp.Provider
	log.Info().Str("address", address).Str("ip", provider.IP).Msg("Fetched provider from Jackal API")
	return &provider, nil
}
