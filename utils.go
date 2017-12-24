package iptree

import (
	"bytes"
	"net"
)

func sameIPLen(x, y net.IPNet) bool {
	return len(x.IP) == len(y.IP)
}

//Don't use net.IP.Equal(ip), it does v4-to-v6 that we don't need
func sameIP(x, y net.IP) bool {
	return bytes.Equal(x, y)
}

func compareIP(x, y net.IP) int {
	return bytes.Compare(x, y)
}

func compareMask(x, y net.IPMask) int {
	return bytes.Compare(x, y)
}
