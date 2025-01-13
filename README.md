# tapo-go

TP-Link TAPO API implemented in Go. Currently, P series is supported (P110, P115) and H200 hub (and its child devices).
Tested with H200 hub and T315 temperature + humidity sensor.

API is not stable, can be changed before release 1.0.0 is released.

## Usage

Smart Plugs (P110, P115)

```go
package main

import (
	"context"
	"github.com/tess1o/tapo-go"
	"log"
	"time"
)

func main() {
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

```

Hub (H200) and its devices:

```go
package main

import (
	"context"
	"github.com/tess1o/tapo-go"
	"log"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	hub, err := tapo.NewHub(ctx, "192.168.1.2", "tapo_email@gmail.com", "my_tapo_password", tapo.Options{})
	if err != nil {
		log.Printf("Error creating hub: %s", err)
		return
	}

	hubInfo, err := hub.GetDeviceInfo(ctx)
	if err != nil {
		log.Printf("Error getting hub device info: %s", err)
	} else {
		log.Printf("Device info: %+v\n", hubInfo)
	}
	
	t := tapo.NewTSeriesDevices(hub)
	seriesDevices, err := t.GetTSeriesDevices(ctx)
	if err != nil {
		log.Printf("Error getting TSeries devices: %s", err)
		return
	}
	log.Printf("T Series devices: %+v\n", seriesDevices)
}

```

## Todo

- Add more methods to P11X and H200 devices

## Credits

Credits to https://github.com/python-kasa/python-kasa since I took AES transport algorithm (used in H200) from that
repository.