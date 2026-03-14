// SPDX-License-Identifier: GPL-3.0-or-later

package httpx

import (
	"crypto/tls"
	"net/http"
	"strings"
	"time"
)

func APIClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

func NoProxyClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: &http.Transport{Proxy: nil},
	}
}

func VaultClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:       timeout,
		CheckRedirect: NoRedirect,
	}
}

func VaultInsecureClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:       timeout,
		CheckRedirect: NoRedirect,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

func TruncateBody(body []byte) string {
	const maxLen = 200
	s := strings.TrimSpace(string(body))
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

func NoRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}
