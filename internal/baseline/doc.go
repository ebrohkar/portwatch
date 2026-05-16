// Package baseline provides a persistent snapshot of ports that are
// considered "normal" for the monitored host.
//
// A Baseline is populated once (either manually or via an initial scan)
// and then used by the reporter to classify subsequent port-change
// alerts as expected or unexpected. The snapshot can be saved to and
// restored from a JSON file so that portwatch survives restarts without
// treating every known-open port as a new anomaly.
//
// Typical usage:
//
//	b, err := baseline.LoadFromFile("baseline.json")
//	if err != nil {
//		// No existing baseline; create one from the current scan.
//		b = baseline.New(initialPorts)
//		_ = b.SaveToFile("baseline.json")
//	}
//	// Later, compare a new scan against the baseline.
//	added, removed := b.Diff(currentPorts)
package baseline
