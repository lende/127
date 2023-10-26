package lib127

import (
	"encoding/binary"
	"net"
	"testing"
)

func TestBlockRange(t *testing.T) {
	tests := []struct{ block, minIP, maxIP string }{
		{"127.0.0.0/8", "127.0.0.1", "127.255.255.254"},
	}
	minIP, maxIP := make(net.IP, 4), make(net.IP, 4)
	for _, test := range tests {
		_, ipnet, err := net.ParseCIDR(test.block)
		if err != nil {
			t.Errorf("Unexpected error:\n\t%v", err)
		}
		minU32, maxU32 := ipSpan(ipnet)
		binary.BigEndian.PutUint32(minIP, minU32)
		binary.BigEndian.PutUint32(maxIP, maxU32)
		if test.minIP != minIP.String() || test.maxIP != maxIP.String() {
			t.Errorf("Host address range in block %v:\nwant:\n\t%v - %v,\nhave:\n\t%v - %v",
				test.block, test.minIP, test.maxIP, minIP.String(), maxIP.String())
		}
	}
}
