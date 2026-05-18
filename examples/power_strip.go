package main

import (
	"context"
	"fmt"
	"log"
	"time"

	tapo "github.com/tess1o/tapo-go"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	strip, err := tapo.NewPowerStrip(ctx, "192.168.1.10", "tapo_email@gmail.com", "my_tapo_password", tapo.Options{
		RetryConfig: tapo.DefaultRetryConfig,
	})
	if err != nil {
		log.Fatalf("Error creating power strip: %s", err)
	}

	sockets, err := strip.GetSockets(ctx)
	if err != nil {
		log.Fatalf("Error getting sockets: %s", err)
	}

	for _, s := range sockets {
		fmt.Printf("Socket %d (%s):\n", s.Position, s.Nickname)
		fmt.Printf("  On:            %v (on for %ds)\n", s.DeviceOn, s.OnTime)
		fmt.Printf("  Power:         %dW (%dmW)\n", s.CurrentPower, s.PowerMw)
		fmt.Printf("  Current/Volt:  %dmA / %dmV\n", s.CurrentMa, s.VoltageMv)
		fmt.Printf("  Today:         %dWh in %d min\n", s.TodayEnergy, s.TodayRuntime)
		fmt.Printf("  This month:    %dWh in %d min\n", s.MonthEnergy, s.MonthRuntime)
		fmt.Printf("  Cumulative:    %dWh\n", s.EnergyWh)
		fmt.Println()
	}
}
