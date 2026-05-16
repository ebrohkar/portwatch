// Package dedup implements alert deduplication for portwatch.
//
// A Deduplicator tracks recently seen alerts by a composite key of
// (port, event, severity). If the same combination is seen again within
// the configured TTL, Allow returns false and the alert is dropped.
//
// Use Flush periodically to evict expired entries and bound memory usage.
package dedup
