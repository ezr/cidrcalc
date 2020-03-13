package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

func usage() {
	fmt.Printf("Usage: %s ADDRESS/MASK\n", os.Args[0])
	fmt.Printf("Example: %s 10.9.19.101/23\n", os.Args[0])
	os.Exit(1)
}

func getAddrBinary(ip net.IP) string {
	binaryRepresentation := ""
	length := len(ip)
	for i, octet := range ip {
		if i+1 == length {
			binaryRepresentation = binaryRepresentation + fmt.Sprintf("%08b", octet)
		} else {
			binaryRepresentation = binaryRepresentation + fmt.Sprintf("%08b.", octet)
		}
	}
	return binaryRepresentation
}

func longestIPLength(addrs ...net.IP) int {
	max := 0
	for _, addr := range addrs {
		if max < len(addr.String()) {
			max = len(addr.String())
		}
	}
	return max
}

func padding(addr net.IP, l int) string {
	return strings.Repeat(" ", l-len(addr.String()))
}

func main() {
	if len(os.Args) != 2 {
		usage()
	}
	isIPv4CIDR, err := regexp.MatchString("^\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\/\\d{1,2}$", os.Args[1])
	if err != nil {
		log.Fatal("error with IPv4 regex")
	}
	if !isIPv4CIDR {
		fmt.Println("error - argument must be an IPv4 address in CIDR notation")
		usage()
	}

	ip, ipnet, err := net.ParseCIDR(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	ip = ip.To4()

	maskAddr := net.IPv4(ipnet.Mask[0], ipnet.Mask[1], ipnet.Mask[2], ipnet.Mask[3]).To4()

	var maskInverse [4]byte
	for i, val := range ipnet.Mask {
		maskInverse[i] = val ^ 0xff
	}

	broadcastAddr := net.ParseIP("0.0.0.0").To4()
	for i, val := range ip {
		broadcastAddr[i] = val | maskInverse[i]
	}

	maxAddr := net.ParseIP("0.0.0.0").To4()
	copy(maxAddr, broadcastAddr)
	maxAddr[3] = maxAddr[3] - 1

	minAddr := ipnet.IP.To4()
	minAddr[3] = minAddr[3] + 1

	maskOnes, _ := ipnet.Mask.Size()

	if maskOnes == 31 {
		// /31 is a special case: no broadcast address and the first address is usable
		minAddr[3] = minAddr[3] - 1
		maxAddr[3] = minAddr[3] + 1
	}

	numHosts := (1 << (32 - maskOnes)) - 2

	l := longestIPLength(ip, maskAddr, minAddr, maxAddr, broadcastAddr)

	namePadding := 9
	// amount of padding to use depends on what
	// fields we are going to print
	if maskOnes > 31 {
		namePadding = 4
	} else if maskOnes > 30 {
		namePadding = 7
	}
	printAddress := func(name string, addr net.IP) {
		fmt.Printf("%s%s : %s  %s%s\n", name, strings.Repeat(" ", namePadding-len(name)), addr, padding(addr, l), getAddrBinary(addr))
	}

	printAddress("ip", ip)
	printAddress("mask", maskAddr)
	if maskOnes < 32 {
		printAddress("minAddr", minAddr)
		printAddress("maxAddr", maxAddr)
	}
	if maskOnes < 31 {
		printAddress("broadcast", broadcastAddr)
		fmt.Println("num hosts :", numHosts)
	}
}
