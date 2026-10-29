// Package core defines the platform-neutral domain model and the public
// contracts used by the yatori SDK.
//
// The package deliberately contains no network implementation. Platform
// behaviour lives in internal/driver and is exposed through the Client
// interface. This keeps the public surface stable while allowing the
// legacy api/ and aggregation/ implementations to evolve independently.
package core
