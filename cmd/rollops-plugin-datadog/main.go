// Command rollops-plugin-datadog is a Rollops metric-provider plugin backed by
// Datadog. Build it, pin its sha256, and point a rollout's analysis.plugin at
// the binary.
package main

import (
	"fmt"
	"os"

	datadog "github.com/klarlabs-studio/rollops-plugin-datadog"
	"go.klarlabs.de/rollops/pkg/plugin"
)

// version is overwritten at build time via -ldflags.
var version = "dev"

func main() {
	safety := plugin.Safety{
		NetworkHosts: []string{"api.datadoghq.com:443"},
		EnvVars:      []string{"DD_API_URL", "DD_API_KEY", "DD_APP_KEY", "DD_WINDOW"},
		RiskClass:    plugin.RiskPassive, // reads metrics only
	}
	if err := plugin.ServeMetricProvider("klarlabs/datadog", version, datadog.FromEnv(), safety); err != nil {
		fmt.Fprintln(os.Stderr, "rollops-plugin-datadog:", err)
		os.Exit(1)
	}
}
