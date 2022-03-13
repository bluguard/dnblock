package client

import "net"

type Client interface {
	ResolveV4(name string) (net.IP, error)
	ResolveV6(name string) (net.IP, error)
}

type ReversableClient interface {
	Client
	ReverseResolve(ip string)
}
