package filter

import (
	"strconv"
	"strings"

	"github.com/fitstar/falcore"
)

// fixme: probably should use net.SplitHostPort
func SplitHostPort(hostPort string, defaultPort int) (string, int) {
	parts := strings.Split(hostPort, ":")
	upstreamHost := parts[0]
	upstreamPort := defaultPort
	if len(parts) > 1 {
		var err error
		upstreamPort, err = strconv.Atoi(parts[1])
		if err != nil {
			upstreamPort = defaultPort
			falcore.Error("UpstreamPool Error converting port to int for", upstreamHost, ":", err)
		}
	}
	return upstreamHost, upstreamPort
}
