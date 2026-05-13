// Package baseline provides a persistent snapshot of ports that are
// considered "normal" for the monitored host.
//
// A Baseline is populated once (either manually or via an initial scan)
// and then used by the reporter to classify subsequent port-change
// alerts as expected or unexpected. The snapshot can be saved to and
// restored from a JSON file so that portwatch survives restarts without
// treating every known-open port as a new anomaly.
package baseline
