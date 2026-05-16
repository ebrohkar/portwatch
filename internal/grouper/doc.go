// Package grouper provides alert grouping by named port sets.
//
// Alerts are accumulated via Add and retrieved in bulk with Flush, which
// resets each group after returning its contents. This allows periodic
// batch processing of related port events without per-alert overhead.
package grouper
