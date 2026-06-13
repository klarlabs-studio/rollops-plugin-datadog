# rollops-plugin-datadog

A [Rollops](https://github.com/klarlabs-studio/rollops) metric-provider plugin
backed by [Datadog](https://www.datadoghq.com/). It resolves a Datadog metrics
query to a single scalar so rollout analysis can gate a canary on Datadog
metrics, the way it does on Prometheus.

## How it works

Rollops calls the plugin's `query_metric` tool with a Datadog metrics query. The
plugin queries Datadog's `/api/v1/query` endpoint over a lookback window
(default 5m) and returns the most recent point of the first series. Wire those
values into a CEL condition in the analysis block:

```yaml
analysis:
  provider: datadog
  plugin: ~/.rollops/plugins/datadog
  sha256: <pin>
  metrics:
    - name: errorRate
      query: "avg:trace.http.request.errors{service:checkout}.as_rate()"
    - name: p99
      query: "p99:trace.http.request.duration{service:checkout}"
  condition: "errorRate < 0.05 && p99 < 0.8"
  count: 3
  interval: 30s
```

## Configuration

Credentials come from the plugin's own environment, never from the Rollops
target spec (which carries only the query string):

| Env var      | Required | Default                     | Description                 |
|--------------|----------|-----------------------------|-----------------------------|
| `DD_API_URL` | no       | `https://api.datadoghq.com` | Use the EU site if applicable |
| `DD_API_KEY` | yes      | —                           | Datadog API key             |
| `DD_APP_KEY` | yes      | —                           | Datadog application key     |
| `DD_WINDOW`  | no       | `5m`                        | Lookback window (Go duration) |

## Install

```sh
rollops plugin install datadog
```

## License

MIT
