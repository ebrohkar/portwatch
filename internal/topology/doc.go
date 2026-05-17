// Package topology maps port numbers to named services and logical tiers
// (frontend, backend, database, internal). It is used by the enrichment
// and alerting layers to attach human-readable context to port events,
// making it easier to assess the impact of an unexpected open or closed port.
package topology
