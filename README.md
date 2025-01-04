# tapo-go

TP-Link TAPO API implemented in Go. Currently, P series is supported (P110, P115) and H200 hub (and its child devices).
Tested with H200 hub and T315 temperature + humidity sensor.

API is not stable, can be changed before release 1.0.0 is released.

## Usage

Smart Plugs (P110, P115)

```go
p115, err := tapo.NewSmartPlug("192.169.1.10", "tapo_email@gmail.com", "my_tapo_password", tapo.Options{})
if err != nil {
    log.Fatalf("Error creating smart plug: %s", err)
}
energyUsage, err := p115.GetEnergyUsage()
if err != nil {
    log.Fatalf("Error getting energy usage: %s", err)
}
log.Printf("Energy usage: %+v\n", energyUsage.Result)
```

Hub (H200) and its devices:

```go
hub, err := tapo.NewHub("192.168.1.15", "tapo_email@gmail.com", "my_tapo_password", tapo.Options{})
if err != nil {
    log.Fatalf("Error creating hub: %s", err)
}
devices, err := hub.GetChildDevices()
if err != nil {
    log.Fatalf("Error getting child devices: %s", err)
}
devicesJsonResponse, err := devices.MarshalJSON()
if err != nil {
    log.Fatalf("Error marshalling devices: %s", err)
}
log.Printf("Devices: %+v\n", string(devicesJsonResponse))
```

## Todo

- Add more methods to P11X and H200 devices
- Add structs to reflect the API responses, instead of json.RawMessage
- Add more error handling for AES transport (used in H200)

## Credits

Credits to https://github.com/python-kasa/python-kasa since I took AES transport algorithm (used in H200) from that
repository.