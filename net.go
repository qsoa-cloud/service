package service

import (
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

var netAddrRe = regexp.MustCompile(`^(?:([\w]+)://)?(.+$)`)

func splitNetAddr(addr string) (string, string) {
	addrParts := netAddrRe.FindStringSubmatch(addr)
	if len(addrParts) != 3 {
		log.Fatalf("Invalid address '%s'", addr)
	}
	if addrParts[1] == "" {
		addrParts[1] = "tcp"
	}

	return addrParts[1], addrParts[2]
}

func qListen(sNet, sAddr string) net.Listener {
	l, err := net.Listen(sNet, sAddr)
	if err != nil {
		log.Fatalf("Cannot listen %s://%s: %v", sNet, sAddr, err)
	}

	if strings.HasPrefix(sNet, "unix") {
		if err := os.Chmod(sAddr, os.ModeSocket|0660); err != nil {
			log.Fatalf("Cannot change socket %s permissions: %v", sAddr, err)
		}
	}

	return l
}
