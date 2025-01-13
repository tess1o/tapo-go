package main

import (
	"context"
	"github.com/tess1o/tapo-go"
	"log"
	"time"
)

func main() {
	smartPlugExample()
	hubExample()
	tDevicesExample()
}

func smartPlugExample() {
	ctx := context.Background()
	p115, err := tapo.NewSmartPlug(ctx, "192.168.1.10", "tapo_email@gmail.com", "my_tapo_password", tapo.Options{
		RetryConfig: tapo.DefaultRetryConfig,
	})
	if err != nil {
		log.Printf("Error creating smart plug: %s", err)
		return
	}
	energyUsage, err := p115.GetEnergyUsage(ctx)
	if err != nil {
		log.Printf("Error getting energy usage: %s", err)
		return
	}
	log.Printf("Energy usage: %+v\n", energyUsage.Result)
}

func hubExample() {
	ctx := context.Background()
	hub, err := tapo.NewHub(ctx, "192.168.1.11", "tapo_email@gmail.com", "my_tapo_password", tapo.Options{
		RetryConfig: tapo.DefaultRetryConfig,
	})
	if err != nil {
		log.Fatalf("Error creating hub: %s", err)
	}
	info, err := hub.GetDeviceInfo(ctx)
	if err != nil {
		log.Printf("Error getting device info: %s", err)
	} else {
		log.Printf("Device info: %+v\n", info)
	}
}

func tDevicesExample() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	hub, err := tapo.NewHub(ctx, "192.168.1.2", "tapo_email@gmail.com", "my_tapo_password", tapo.Options{})
	if err != nil {
		log.Printf("Error creating hub: %s", err)
		return
	}
	t := tapo.NewTSeriesDevices(hub)
	seriesDevices, err := t.GetTSeriesDevices(ctx)
	if err != nil {
		log.Printf("Error getting TSeries devices: %s", err)
		return
	}
	log.Printf("T Series devices: %+v\n", seriesDevices)
}
