package geolocation

import (
	"net"
	"net/http"
	"strings"
)

// ExtractClientIP extracts the real client IP from HTTP request headers
// It handles common proxy headers like X-Forwarded-For and X-Real-IP
func ExtractClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (most common)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs separated by commas
		// The first one is typically the original client IP
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if isValidIP(ip) {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		ip := strings.TrimSpace(xri)
		if isValidIP(ip) {
			return ip
		}
	}

	// Check CF-Connecting-IP (Cloudflare)
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		ip := strings.TrimSpace(cfip)
		if isValidIP(ip) {
			return ip
		}
	}

	// Check X-Client-IP header
	if xcip := r.Header.Get("X-Client-IP"); xcip != "" {
		ip := strings.TrimSpace(xcip)
		if isValidIP(ip) {
			return ip
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

// isValidIP checks if the provided string is a valid IP address
// and excludes private/local addresses
func isValidIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Exclude private IP ranges and localhost
	if parsedIP.IsLoopback() || parsedIP.IsPrivate() {
		return false
	}

	return true
}

// IsPrivateOrLocalIP checks if the IP is private or local
func IsPrivateOrLocalIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return true // Assume private if can't parse
	}

	return parsedIP.IsLoopback() || parsedIP.IsPrivate()
}