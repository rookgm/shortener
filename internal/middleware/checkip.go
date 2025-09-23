package middleware

import (
	"net"
	"net/http"
	"strings"
)

// CheckTrustedSubNet checks trusted subnet
func CheckTrustedSubNet(s string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if trusted subnet is empty, then return forbidden
		if s == "" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		// parse CIDR
		_, ipv4Net, err := net.ParseCIDR(s)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		// get client ip from request header value
		v := r.Header.Get("X-Real-IP")
		// try parse ip
		clientIP := net.ParseIP(v)
		if clientIP == nil {
			ips := r.Header.Get("X-Forwarded-For")
			ipStrs := strings.Split(ips, ",")
			ipStr := ipStrs[0]
			clientIP = net.ParseIP(ipStr)
			if clientIP == nil {
				http.Error(w, "invalid client ip", http.StatusBadRequest)
				return
			}
		}
		// check trusted subnet includes client ip
		if !ipv4Net.Contains(clientIP) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
