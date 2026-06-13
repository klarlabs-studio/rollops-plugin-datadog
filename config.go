package datadog

import (
	"os"
	"time"
)

// FromEnv builds a Provider from the plugin's environment. Secrets and endpoint
// come from the plugin process, never from the Rollops target spec (Rollops
// passes only the query string).
//
//	DD_API_URL  base URL (default https://api.datadoghq.com; use the EU site if applicable)
//	DD_API_KEY  Datadog API key (required)
//	DD_APP_KEY  Datadog application key (required)
//	DD_WINDOW   lookback window as a Go duration (default 5m)
func FromEnv() Provider {
	base := os.Getenv("DD_API_URL")
	if base == "" {
		base = "https://api.datadoghq.com"
	}
	win, _ := time.ParseDuration(os.Getenv("DD_WINDOW"))
	return Provider{
		BaseURL: base,
		APIKey:  os.Getenv("DD_API_KEY"),
		AppKey:  os.Getenv("DD_APP_KEY"),
		Window:  win,
	}
}
