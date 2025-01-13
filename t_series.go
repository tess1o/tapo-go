package tapo

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
)

type TSeries struct {
	hub *Hub
}

func NewTSeriesDevices(hub *Hub) *TSeries {
	return &TSeries{hub}
}

func (t *TSeries) GetTSeriesDevices(ctx context.Context) ([]TSeriesResponse, error) {
	devices, err := t.hub.GetChildDevices(ctx)
	if err != nil {
		return nil, err
	}
	if devices.ErrorCode != 0 {
		return nil, errors.New("error getting devices, error code: " + strconv.Itoa(devices.ErrorCode))
	}
	deviceList := make([]TSeriesResponse, 0)
	for _, d := range devices.Result.Responses {
		var childDeviceList []byte
		childDeviceList, err = d.Result.ChildDeviceList.MarshalJSON()
		if err != nil {
			return nil, err
		}
		var response []TSeriesResponse
		err = json.Unmarshal(childDeviceList, &response)
		if err != nil {
			return nil, err
		}
		deviceList = append(deviceList, response...)
	}
	return deviceList, err
}

type TSeriesResponse struct {
	ParentDeviceId           string  `json:"parent_device_id"`
	HwVer                    string  `json:"hw_ver"`
	FwVer                    string  `json:"fw_ver"`
	DeviceId                 string  `json:"device_id"`
	Mac                      string  `json:"mac"`
	Type                     string  `json:"type"`
	Model                    string  `json:"model"`
	HwId                     string  `json:"hw_id"`
	OemId                    string  `json:"oem_id"`
	Specs                    string  `json:"specs"`
	Category                 string  `json:"category"`
	BindCount                int     `json:"bind_count"`
	StatusFollowEdge         bool    `json:"status_follow_edge"`
	Status                   string  `json:"status"`
	LastOnboardingTimestamp  int     `json:"lastOnboardingTimestamp"`
	Rssi                     int     `json:"rssi"`
	SignalLevel              int     `json:"signal_level"`
	JammingRssi              int     `json:"jamming_rssi"`
	JammingSignalLevel       int     `json:"jamming_signal_level"`
	AtLowBattery             bool    `json:"at_low_battery"`
	TempUnit                 string  `json:"temp_unit"`
	CurrentTemp              float64 `json:"current_temp"`
	CurrentHumidity          int     `json:"current_humidity"`
	CurrentTempException     float64 `json:"current_temp_exception"`
	CurrentHumidityException int     `json:"current_humidity_exception"`
	Nickname                 string  `json:"nickname"`
	Avatar                   string  `json:"avatar"`
	ReportInterval           int     `json:"report_interval"`
	Region                   string  `json:"region"`
}
