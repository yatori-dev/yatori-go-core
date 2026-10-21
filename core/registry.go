package core

import (
	"fmt"
	"sort"
	"sync"
)

// Factory constructs a Client for one platform.
type Factory func(Config) (Client, error)

var (
	registryMu sync.RWMutex
	registry   = make(map[Platform]Factory)
)

// Register adds a platform factory. It is intended to be called from an
// adapter's init function. Registering the same platform twice panics, as
// that always indicates a programming error.
func Register(platform Platform, factory Factory) {
	if !platform.Valid() {
		panic(fmt.Sprintf("yatori: register invalid platform %q", platform))
	}
	if factory == nil {
		panic(fmt.Sprintf("yatori: register nil factory for %s", platform))
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := registry[platform]; exists {
		panic(fmt.Sprintf("yatori: platform %s already registered", platform))
	}
	registry[platform] = factory
}

// New builds a client for cfg.Platform.
func New(cfg Config) (Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	cfgo := cfg.Options.Resolve()
	cfg.Options = cfgo

	registryMu.RLock()
	factory, ok := registry[cfg.Platform]
	registryMu.RUnlock()
	if !ok {
		return nil, &Error{
			Platform: cfg.Platform,
			Op:       "new",
			Code:     "unregistered",
			Message:  "no adapter registered for platform",
		}
	}
	client, err := factory(cfg)
	if err != nil {
		return nil, Wrap(cfg.Platform, "new", err)
	}
	if client == nil {
		return nil, &Error{Platform: cfg.Platform, Op: "new", Code: "nil_client"}
	}
	return client, nil
}

// Registered returns the platforms that currently have an adapter.
func Registered() []Platform {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]Platform, 0, len(registry))
	for platform := range registry {
		out = append(out, platform)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
