// Package reporter provides the Reporter type, which orchestrates a complete
// port-monitoring cycle:
//
//  1. Scan the configured port range for open TCP ports.
//  2. Diff the results against the previously persisted state.
//  3. Evaluate each changed port against the configured rule set.
//  4. Emit an Alert via the Notifier for every rule violation.
//  5. Persist the new port state so the next cycle has a baseline.
//
// Reporter is intentionally stateless between calls to Run; all persistence
// is delegated to the state.Store.
package reporter
