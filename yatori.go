// Package yatori is the public entry point of yatori-go-core.
//
// A caller creates one Client per account:
//
//	client, err := yatori.New(yatori.Config{
//	    Platform: yatori.XueXiTong,
//	    Account:  yatori.Account{Username: "ACCOUNT", Password: "PASSWORD"},
//	})
//	if err != nil {
//	    return err
//	}
//	defer client.Close()
//
//	if err := client.Login(ctx); err != nil {
//	    return err
//	}
//
// The same methods are available on every platform. Operations a platform
// does not implement return an error wrapping ErrUnsupported; check
// client.Capabilities() to discover support up front.
package yatori

import (
	"context"

	"github.com/yatori-dev/yatori-go-core/v3/core"
	_ "github.com/yatori-dev/yatori-go-core/v3/internal/driver"
)

// New constructs a Client for cfg.Platform.
func New(cfg Config) (Client, error) {
	return core.New(cfg)
}

// Call dispatches a dynamic action to a client. Typed code should prefer
// the methods on Client; Call is aimed at configuration driven callers.
func Call(ctx context.Context, client Client, action Action, params map[string]any) (any, error) {
	return core.Call(ctx, client, action, params)
}

// Supports reports whether client implements every capability in want.
func Supports(client Client, want Capability) bool {
	return core.Supports(client, want)
}

// Registered returns the platforms that currently have an adapter.
func Registered() []Platform {
	return core.Registered()
}

// AllPlatforms returns every platform known to the SDK.
func AllPlatforms() []Platform {
	return core.AllPlatforms()
}

// ParsePlatform normalizes a platform identifier or alias.
func ParsePlatform(raw string) (Platform, error) {
	return core.ParsePlatform(raw)
}

// ParseAction normalizes a dynamic action name.
func ParseAction(raw string) (Action, error) {
	return core.ParseAction(raw)
}

// Flatten returns every leaf node in depth first order.
func Flatten(nodes []Node) []Node {
	return core.Flatten(nodes)
}
