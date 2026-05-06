// Package filter provides port filtering for portwatch scan results.
//
// A Filter can be configured with an explicit include list, an exclude list,
// or both. When the include list is empty all ports are considered candidates;
// the exclude list is always applied last and takes precedence.
//
// Typical usage:
//
//	f, err := filter.New(includePorts, excludePorts)
//	if err != nil { ... }
//	allowedPorts := f.Apply(scannedPorts)
package filter
