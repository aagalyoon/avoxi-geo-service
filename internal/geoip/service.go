package geoip

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/oschwald/geoip2-golang"
	"github.com/sirupsen/logrus"
)

// Service provides IP geolocation functionality
type Service struct {
	db       *geoip2.Reader
	logger   *logrus.Logger
	mu       sync.RWMutex
	demoMode bool
}

// NewService creates a new GeoIP service
func NewService(dbPath string, logger *logrus.Logger) (*Service, error) {
	service := &Service{
		logger:   logger,
		demoMode: false,
	}

	// Check if demo mode is requested
	if dbPath == "" || dbPath == "demo" {
		logger.Info("Running in demo mode - using sample data")
		service.demoMode = true
		return service, nil
	}

	// Try to open the database
	db, err := geoip2.Open(dbPath)
	if err != nil {
		logger.Warnf("Failed to open GeoIP database: %v. Running in demo mode.", err)
		service.demoMode = true
		return service, nil
	}

	service.db = db
	logger.Infof("GeoIP database loaded from %s", dbPath)
	return service, nil
}

// Close closes the GeoIP database
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// ValidateIP checks if an IP address is from one of the allowed countries
func (s *Service) ValidateIP(ipStr string, allowedCountries []string) (bool, string, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, "", fmt.Errorf("invalid IP address: %s", ipStr)
	}

	country, err := s.getCountry(ip)
	if err != nil {
		return false, "", err
	}

	// Check if the country is in the allowed list
	allowed := false
	for _, allowedCountry := range allowedCountries {
		if strings.EqualFold(country, allowedCountry) {
			allowed = true
			break
		}
	}

	return allowed, country, nil
}

// getCountry retrieves the country code for an IP address
func (s *Service) getCountry(ip net.IP) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Demo mode returns sample data
	if s.demoMode {
		return s.getDemoCountry(ip.String()), nil
	}

	if s.db == nil {
		return "", fmt.Errorf("GeoIP database not available")
	}

	record, err := s.db.Country(ip)
	if err != nil {
		return "", fmt.Errorf("failed to lookup IP: %w", err)
	}

	return record.Country.IsoCode, nil
}

// getDemoCountry returns sample country data for demo mode
func (s *Service) getDemoCountry(ip string) string {
	// Sample mappings for demo mode
	demoData := map[string]string{
		"8.8.8.8":       "US",
		"8.8.4.4":       "US",
		"1.1.1.1":       "AU",
		"134.195.196.1": "GB",
		"200.148.32.1":  "BR",
		"142.250.80.46": "US",
		"185.60.216.35": "RU",
		"31.13.72.36":   "IE",
		"157.240.12.35": "US",
		"151.101.1.140": "US",
	}

	if country, ok := demoData[ip]; ok {
		return country
	}

	// Default to US for unknown IPs in demo mode
	return "US"
}

// UpdateDatabase updates the GeoIP database with a new file
func (s *Service) UpdateDatabase(newDBPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Open new database
	newDB, err := geoip2.Open(newDBPath)
	if err != nil {
		return fmt.Errorf("failed to open new database: %w", err)
	}

	// Close old database if exists
	if s.db != nil {
		s.db.Close()
	}

	// Replace with new database
	s.db = newDB
	s.demoMode = false
	s.logger.Info("GeoIP database updated successfully")

	return nil
}