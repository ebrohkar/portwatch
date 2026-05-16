// Package trend provides a sliding-window alert frequency tracker for portwatch.
//
// Trend buckets incoming alert recordings by a configurable time interval and
// compares successive buckets to determine whether alert volume for a given
// port is rising, falling, or stable. This information can be used by
// downstream pipeline stages or summary reporters to surface emerging threats
// before they trigger hard thresholds.
//
// Example usage:
//
//	tr := trend.New(time.Minute, 5)
//	tr.Record(8080)
//	fmt.Println(tr.Summary(8080)) // "port 8080: stable (current bucket count: 1)"
package trend
