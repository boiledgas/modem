package main

import (
	"log"

	"encoding/hex"
	"modem"
	"time"
)

func main() {
	m := modem.Huawei{}
	m.Name = "COM7"
	m.Timeout = time.Second * 5
	if err := m.Connect(); err != nil {
		log.Printf("connect: %v", err)
	}
	if err := m.Reset(); err != nil {
		log.Printf("atz: %v", err)
	}
	if err := m.Echo(true); err != nil {
		log.Printf("echo: %v", err)
	}
	if err := m.ReportRssi(true); err != nil {
		log.Printf("curc: %v", err)
	}
	if result, err := m.Information(); err != nil {
		log.Printf("command: %v", err)
	} else {
		log.Printf("result: %v", result)
	}
	if err := m.MessageFormat(0); err != nil {
		log.Printf("%v", err)
	}
	bytes, _ := hex.DecodeString("0011000B919761167499F20000FF0BE8329BFD06DDDF723619")
	if err := m.PduMessage(bytes); err != nil {
		log.Printf("%v", err)
	}

	time.Sleep(time.Second * 10)
	log.Printf("disconnect")
	if err := m.Disconnect(); err != nil {
		log.Printf("disconnect: %v", err)
	}
}
