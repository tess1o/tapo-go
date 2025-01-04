package tapo

import "encoding/json"

type SmartPlug struct {
	*Device
}

func NewSmartPlug(host, email, password string, options Options) (*SmartPlug, error) {
	tr, err := NewKlapTransport(email, password, host, nil)
	if err != nil {
		return nil, err
	}
	tapo := NewDevice(tr, options)
	return &SmartPlug{tapo}, nil
}

func (t *SmartPlug) TurnOn() (*SetDeviceParameterResponse, error) {
	var response *SetDeviceParameterResponse
	params := json.RawMessage("{\"device_on\":true}")
	err := t.ExecuteMethod("set_device_info", params, &response)
	return response, err
}

func (t *SmartPlug) TurnOff() (*SetDeviceParameterResponse, error) {
	var response *SetDeviceParameterResponse
	params := json.RawMessage("{\"device_on\":false}")
	err := t.ExecuteMethod("set_device_info", params, &response)
	return response, err
}

func (t *SmartPlug) GetEnergyUsage() (*EnergyUsageResponse, error) {
	var response *EnergyUsageResponse
	err := t.ExecuteMethod("get_energy_usage", nil, &response)
	return response, err
}

func (t *SmartPlug) GetCurrentPower() (*CurrentPower, error) {
	var response *CurrentPower
	err := t.ExecuteMethod("get_current_power", nil, &response)
	return response, err
}

func (t *SmartPlug) GetEmeterData() (*EmeterData, error) {
	var response *EmeterData
	err := t.ExecuteMethod("get_emeter_data", nil, &response)
	return response, err
}

func (t *SmartPlug) DeviceInfo() (*DeviceInfoResponse, error) {
	var response *DeviceInfoResponse
	err := t.ExecuteMethod("get_device_info", nil, &response)
	return response, err
}
