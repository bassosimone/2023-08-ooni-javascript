package main

import (
	"context"

	"github.com/apex/log"
	"github.com/bassosimone/2023-08-ooni-javascript/pkg/geolocatev2"
	"github.com/ooni/probe-engine/pkg/netxlite"
)

func main() {
	// FIXME: all the stuff below looks a bit clumsy to me
	netx := &netxlite.Netx{
		Underlying: &netxlite.DefaultTProxy{},
	}
	reso := netx.NewStdlibResolver(log.Log)
	client := geolocatev2.NewSTUNIPLookupClient("stun.l.google.com:19302", log.Log, reso, netx.Underlying)
	addr, err := client.LookupProbeIPv4(context.Background())
	log.Infof("%+v %+v", addr, err)
	addr, err = client.LookupProbeIPv6(context.Background())
	log.Infof("%+v %+v", addr, err)
}
