package net

import (
	"context"
	"net"
	"time"
)

// ListenConfig extends the standard net.ListeConfig with a central holder
// for the Context bound to the listeners
type ListenConfig struct {
	net.ListenConfig

	// Context used when registering the listeners
	Context context.Context
}

// NewListenConfig assists creating a ListenConfig due to the two-layer definition
// making difficult static declaration when `net` is shadowed
func NewListenConfig(ctx context.Context, keepalive time.Duration) *ListenConfig {
	if ctx == nil {
		ctx = context.Background()
	}

	return &ListenConfig{
		ListenConfig: net.ListenConfig{
			KeepAlive: keepalive,
		},
		Context: ctx,
	}
}

// Listen acts like the standard net.Listen but using the context.Context,
// KeepAlive, and optional Control function from our ListenConfig struct
func (lc ListenConfig) Listen(network, addr string) (net.Listener, error) {
	ctx := lc.Context
	if ctx == nil {
		ctx = context.Background()
	}

	return lc.ListenConfig.Listen(ctx, network, addr)
}

// ListenPacket acts like the standard net.ListenPacket but using the context.Context,
// KeepAlive, and optional Control function from our ListenConfig struct
func (lc ListenConfig) ListenPacket(network, addr string) (net.PacketConn, error) {
	ctx := lc.Context
	if ctx == nil {
		ctx = context.Background()
	}

	return lc.ListenConfig.ListenPacket(ctx, network, addr)
}

// ListenAll acts like Listen but on a list of addresses
func (lc ListenConfig) ListenAll(network string, addrs []string) ([]net.Listener, error) {
	out := make([]net.Listener, 0, len(addrs))

	for _, addr := range addrs {
		lsn, err := lc.Listen(network, addr)
		if err != nil {
			for _, lsn := range out {
				_ = lsn.Close()
			}
			return nil, err
		}
		out = append(out, lsn)
	}

	return out, nil
}

// ListenAllPacket acts like ListenPacket but on a list of addresses
func (lc ListenConfig) ListenAllPacket(network string, addrs []string) ([]net.PacketConn, error) {
	out := make([]net.PacketConn, 0, len(addrs))

	for _, addr := range addrs {
		lsn, err := lc.ListenPacket(network, addr)
		if err != nil {
			for _, lsn := range out {
				_ = lsn.Close()
			}
			return nil, err
		}
		out = append(out, lsn)
	}

	return out, nil
}
