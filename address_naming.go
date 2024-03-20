package main

import (
	"net"
	"net/netip"
	"strings"
)

func EncodeIPToWords(ip net.IP) string {
	digits := ip.To4()
	var digitStrings []string
	for _, digit := range digits {
		digitStrings = append(digitStrings, wordList[digit])
	}
	println(strings.Join(digitStrings, " "))
	println(ip.String())
	return strings.Join(digitStrings, " ")
}


func DecodeIPFromWords(str string) net.IP {
	digitWords := strings.Split(str, " ")
	if len(digitWords) != 4 {
		return nil
	}
	actualDigits := []byte{}
	for _, digitWord := range digitWords {
		for j, word := range wordList {
			if digitWord == word {
				actualDigits = append(actualDigits, byte(j))
			}
		}
	}
	ip, ok := netip.AddrFromSlice(actualDigits)
	if !ok {
		return net.IP{}
	}
	return net.IP(ip.AsSlice())
}