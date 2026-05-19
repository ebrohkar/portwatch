// Package trend provides a sliding-window alert frequency tracker for portwatch.
//
// Trend buckets incoming alert recordings by a configurable time interval and
// compares successive buckets to determine whether alert volume for a given
// port is rising, falling, or stable. This information can be used by
// downstream pipeline stages or summary reporters to surface emerging threats
// before they trigger hard thresholds.
//
// # Architecture
//
// A Tracker maintains a circular buffer of time buckets per port. Each call to
// Record increments the count for the current bucket. When Summary is called,
// the tracker compares the two most recent non-empty buckets to classify the
// trend direction:
//
//   - Rising:  current bucket count exceeds the previous bucket count
//   - Falling: current bucket count is below the previous bucket count
//   - Stable:  counts are equal, or fewer than two buckets have data
//
// # Example usage
//
//	tr := trend.New(time.Minute, 5)
//	tr.Record(8080)
//	fmt.Println(tr.Summary(8080)) // "port 8080: stable (current bucket count: 1)"
package trend
