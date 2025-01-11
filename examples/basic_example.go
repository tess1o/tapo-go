package main

import (
	"github.com/tess1o/tapo-go"
	"log"
)

func main() {
	smartPlugExample()
	hubExample()
	tDevicesExample()
}

func smartPlugExample() {
	p115, err := tapo.NewSmartPlug("192.168.1.10", "tapo_email@gmail.com", "my_tapo_password", tapo.Options{})
	if err != nil {
		log.Fatalf("Error creating smart plug: %s", err)
	}
	energyUsage, err := p115.GetEnergyUsage()
	if err != nil {
		log.Fatalf("Error getting energy usage: %s", err)
	}
	log.Printf("Energy usage: %+v\n", energyUsage.Result)
}

func hubExample() {
	hub, err := tapo.NewHub("192.168.1.15", "tapo_email@gmail.com", "my_tapo_password", tapo.Options{})
	if err != nil {
		log.Fatalf("Error creating hub: %s", err)
	}
	devices, err := hub.GetChildDevices()
	if err != nil {
		log.Fatalf("Error getting child devices: %s", err)
	}
	log.Printf("Devices: %+v\n", devices)
}

func tDevicesExample() {
	hub, err := tapo.NewHub("192.168.1.15", "tapo_email@gmail.com", "my_tapo_password", tapo.Options{})
	if err != nil {
		log.Fatalf("Error creating hub: %s", err)
	}
	t := tapo.NewTSeriesDevices(hub)
	seriesDevices, err := t.GetTSeriesDevices()
	if err != nil {
		log.Fatalf("Error getting TSeries devices: %s", err)
	}
	log.Printf("T Series devices: %+v\n", seriesDevices)
}
