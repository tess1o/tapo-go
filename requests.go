package tapo

import "encoding/json"

type RequestSpec struct {
	Method          string          `json:"method"`
	RequestTimeMils int64           `json:"requestTimeMils"`
	TerminalUUID    string          `json:"terminal_uuid,omitempty"`
	Params          json.RawMessage `json:"params,omitempty"`
}

type DeviceInfoResponse struct {
	Result struct {
		DeviceId           string `json:"device_id"`
		FwVer              string `json:"fw_ver"`
		HwVer              string `json:"hw_ver"`
		Type               string `json:"type"`
		Model              string `json:"model"`
		Mac                string `json:"mac"`
		HwId               string `json:"hw_id"`
		FwId               string `json:"fw_id"`
		OemId              string `json:"oem_id"`
		Ip                 string `json:"ip"`
		TimeDiff           int    `json:"time_diff"`
		Ssid               string `json:"ssid"`
		Rssi               int    `json:"rssi"`
		SignalLevel        int    `json:"signal_level"`
		AutoOffStatus      string `json:"auto_off_status"`
		AutoOffRemainTime  int    `json:"auto_off_remain_time"`
		Longitude          int    `json:"longitude"`
		Latitude           int    `json:"latitude"`
		Lang               string `json:"lang"`
		Avatar             string `json:"avatar"`
		Region             string `json:"region"`
		Specs              string `json:"specs"`
		Nickname           string `json:"nickname"`
		HasSetLocationInfo bool   `json:"has_set_location_info"`
		DeviceOn           bool   `json:"device_on"`
		OnTime             int    `json:"on_time"`
		DefaultStates      struct {
			Type  string `json:"type"`
			State struct {
			} `json:"state"`
		} `json:"default_states"`
		OverheatStatus        string `json:"overheat_status"`
		PowerProtectionStatus string `json:"power_protection_status"`
		OvercurrentStatus     string `json:"overcurrent_status"`
		ChargingStatus        string `json:"charging_status"`
	} `json:"result"`
	ErrorCode int `json:"error_code"`
}

type EnergyUsageResponse struct {
	Result struct {
		TodayRuntime      int    `json:"today_runtime"`
		MonthRuntime      int    `json:"month_runtime"`
		TodayEnergy       int    `json:"today_energy"`
		MonthEnergy       int    `json:"month_energy"`
		LocalTime         string `json:"local_time"`
		ElectricityCharge []int  `json:"electricity_charge"`
		CurrentPower      int    `json:"current_power"`
	} `json:"result"`
	ErrorCode int `json:"error_code"`
}

type CurrentPower struct {
	Result struct {
		CurrentPower int `json:"current_power"`
	} `json:"result"`
	ErrorCode int `json:"error_code"`
}

type EmeterData struct {
	Result struct {
		CurrentMa int `json:"current_ma"`
		VoltageMv int `json:"voltage_mv"`
		PowerMw   int `json:"power_mw"`
		EnergyWh  int `json:"energy_wh"`
	} `json:"result"`
	ErrorCode int `json:"error_code"`
}

type SetDeviceParameterResponse struct {
	ErrorCode int `json:"error_code"`
}
