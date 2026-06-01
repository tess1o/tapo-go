package tapo

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// PowerStrip represents a Tapo P304M (or similar) smart Wi-Fi power strip.
// It uses KLAP transport, the same as SmartPlug (P110/P115).
type PowerStrip struct {
	*Device
}

// NewPowerStrip creates a PowerStrip for the given host.
func NewPowerStrip(ctx context.Context, host, email, password string, options Options) (*PowerStrip, error) {
	transport, err := NewKlapTransport(ctx, email, password, host, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create KLAP transport for power strip %s: %w", host, err)
	}
	return &PowerStrip{Device: NewDevice(transport, options)}, nil
}

// PowerStripSocket holds the combined state and energy data for one socket.
type PowerStripSocket struct {
	DeviceID     string `json:"device_id"`
	Position     int    `json:"position"` // 1-based socket position on the strip
	Nickname     string `json:"nickname"` // human-readable name, decoded from base64
	DeviceOn     bool   `json:"device_on"`
	OnTime       int    `json:"on_time"`       // seconds the socket has been continuously on
	TodayRuntime int    `json:"today_runtime"` // minutes on today
	MonthRuntime int    `json:"month_runtime"` // minutes on this month
	TodayEnergy  int    `json:"today_energy"`  // Wh consumed today
	MonthEnergy  int    `json:"month_energy"`  // Wh consumed this month
	CurrentPower int    `json:"current_power"` // current power in W
	CurrentMa    int    `json:"current_ma"`    // current in mA
	VoltageMv    int    `json:"voltage_mv"`    // voltage in mV
	PowerMw      int    `json:"power_mw"`      // current power in mW (higher precision than CurrentPower)
	EnergyWh     int    `json:"energy_wh"`     // cumulative lifetime energy in Wh
}

// GetSockets returns the current state and energy data for all sockets on the strip.
// It makes one call to get_child_device_list, then one control_child call per socket.
func (p *PowerStrip) GetSockets(ctx context.Context) ([]PowerStripSocket, error) {
	var listResp powerStripChildListResponse
	if err := p.ExecuteMethod(ctx, "get_child_device_list", nil, &listResp); err != nil {
		return nil, fmt.Errorf("get_child_device_list failed: %w", err)
	}

	sockets := make([]PowerStripSocket, 0, len(listResp.Result.ChildDeviceList))

	for _, child := range listResp.Result.ChildDeviceList {
		s := PowerStripSocket{
			DeviceID: child.DeviceID,
			Position: child.Position,
			Nickname: decodeBase64Nickname(child.Nickname),
			DeviceOn: child.DeviceOn,
			OnTime:   child.OnTime,
		}

		energy, err := p.getSocketEnergy(ctx, child.DeviceID)
		if err != nil {
			return nil, fmt.Errorf("energy query failed for socket %d (%s): %w", child.Position, s.Nickname, err)
		}

		s.TodayRuntime = energy.todayRuntime
		s.MonthRuntime = energy.monthRuntime
		s.TodayEnergy = energy.todayEnergy
		s.MonthEnergy = energy.monthEnergy
		s.CurrentPower = energy.currentPower
		s.CurrentMa = energy.currentMa
		s.VoltageMv = energy.voltageMv
		s.PowerMw = energy.powerMw
		s.EnergyWh = energy.energyWh

		sockets = append(sockets, s)
	}

	return sockets, nil
}

// --- internal types ---

type powerStripChildListResponse struct {
	Result struct {
		ChildDeviceList []struct {
			DeviceID string `json:"device_id"`
			Position int    `json:"position"`
			Nickname string `json:"nickname"` // base64-encoded
			DeviceOn bool   `json:"device_on"`
			OnTime   int    `json:"on_time"`
		} `json:"child_device_list"`
	} `json:"result"`
	ErrorCode int `json:"error_code"`
}

type controlChildResponse struct {
	Result struct {
		ResponseData struct {
			Result struct {
				Responses []struct {
					Method    string          `json:"method"`
					Result    json.RawMessage `json:"result"`
					ErrorCode int             `json:"error_code"`
				} `json:"responses"`
			} `json:"result"`
		} `json:"responseData"`
	} `json:"result"`
	ErrorCode int `json:"error_code"`
}

type socketEnergyData struct {
	todayRuntime int
	monthRuntime int
	todayEnergy  int
	monthEnergy  int
	currentPower int
	currentMa    int
	voltageMv    int
	powerMw      int
	energyWh     int
}

func (p *PowerStrip) getSocketEnergy(ctx context.Context, deviceID string) (*socketEnergyData, error) {
	innerParams, err := json.Marshal(map[string]interface{}{
		"requests": []map[string]interface{}{
			{"method": "get_energy_usage"},
			{"method": "get_current_power"},
			{"method": "get_emeter_data"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inner requests: %w", err)
	}

	outerParams, err := json.Marshal(map[string]interface{}{
		"device_id": deviceID,
		"requestData": map[string]interface{}{
			"method": "multipleRequest",
			"params": json.RawMessage(innerParams),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal control_child params: %w", err)
	}

	var resp controlChildResponse
	if err := p.ExecuteMethod(ctx, "control_child", outerParams, &resp); err != nil {
		return nil, err
	}

	data := &socketEnergyData{}
	for _, r := range resp.Result.ResponseData.Result.Responses {
		if r.ErrorCode != 0 {
			continue
		}
		switch r.Method {
		case "get_energy_usage":
			var eu struct {
				TodayRuntime int `json:"today_runtime"`
				MonthRuntime int `json:"month_runtime"`
				TodayEnergy  int `json:"today_energy"`
				MonthEnergy  int `json:"month_energy"`
			}
			if err := json.Unmarshal(r.Result, &eu); err == nil {
				data.todayRuntime = eu.TodayRuntime
				data.monthRuntime = eu.MonthRuntime
				data.todayEnergy = eu.TodayEnergy
				data.monthEnergy = eu.MonthEnergy
			}
		case "get_current_power":
			var cp struct {
				CurrentPower int `json:"current_power"`
			}
			if err := json.Unmarshal(r.Result, &cp); err == nil {
				data.currentPower = cp.CurrentPower
			}
		case "get_emeter_data":
			var ed struct {
				CurrentMa int `json:"current_ma"`
				VoltageMv int `json:"voltage_mv"`
				PowerMw   int `json:"power_mw"`
				EnergyWh  int `json:"energy_wh"`
			}
			if err := json.Unmarshal(r.Result, &ed); err == nil {
				data.currentMa = ed.CurrentMa
				data.voltageMv = ed.VoltageMv
				data.powerMw = ed.PowerMw
				data.energyWh = ed.EnergyWh
			}
		}
	}

	return data, nil
}

func decodeBase64Nickname(encoded string) string {
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return encoded // return raw value if decode fails
	}
	return string(b)
}
