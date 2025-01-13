package tapo

import (
	"context"
	"encoding/json"
)

type SmartPlug struct {
	*Device
}

func NewSmartPlug(ctx context.Context, host, email, password string, options Options) (*SmartPlug, error) {
	tr, err := NewKlapTransport(ctx, email, password, host, options)
	if err != nil {
		return nil, err
	}
	tapo := NewDevice(tr, options)
	return &SmartPlug{tapo}, nil
}

func (t *SmartPlug) TurnOn(ctx context.Context) (*SetDeviceParameterResponse, error) {
	var response *SetDeviceParameterResponse
	params := json.RawMessage("{\"device_on\":true}")
	err := t.ExecuteMethod(ctx, "set_device_info", params, &response)
	return response, err
}

func (t *SmartPlug) TurnOff(ctx context.Context) (*SetDeviceParameterResponse, error) {
	var response *SetDeviceParameterResponse
	params := json.RawMessage("{\"device_on\":false}")
	err := t.ExecuteMethod(ctx, "set_device_info", params, &response)
	return response, err
}

func (t *SmartPlug) GetEnergyUsage(ctx context.Context) (*EnergyUsageResponse, error) {
	var response *EnergyUsageResponse
	err := t.ExecuteMethod(ctx, "get_energy_usage", nil, &response)
	return response, err
}

func (t *SmartPlug) GetCurrentPower(ctx context.Context) (*CurrentPower, error) {
	var response *CurrentPower
	err := t.ExecuteMethod(ctx, "get_current_power", nil, &response)
	return response, err
}

func (t *SmartPlug) GetEmeterData(ctx context.Context) (*EmeterData, error) {
	var response *EmeterData
	err := t.ExecuteMethod(ctx, "get_emeter_data", nil, &response)
	return response, err
}

func (t *SmartPlug) DeviceInfo(ctx context.Context) (*DeviceInfoResponse, error) {
	var response *DeviceInfoResponse
	err := t.ExecuteMethod(ctx, "get_device_info", nil, &response)
	return response, err
}
