package tapo

import (
	"context"
	"encoding/json"
)

type Hub struct {
	*Device
}

func NewHub(ctx context.Context, host, email, password string, options Options) (*Hub, error) {
	tr, err := NewSslAesTransport(ctx, host, email, password, options)
	if err != nil {
		return nil, err
	}
	tapo := NewDevice(tr, options)
	return &Hub{tapo}, nil
}

func (h *Hub) GetChildDevices(ctx context.Context) (ChildDeviceListResponse, error) {
	params := json.RawMessage("{\"requests\":[{\"method\":\"getChildDeviceList\",\"params\":{\"childControl\":{\"start_index\":0}}}]}")
	var response ChildDeviceListResponse
	err := h.ExecuteMethod(ctx, "multipleRequest", params, &response)
	return response, err
}

func (h *Hub) GetDeviceInfo(ctx context.Context) (HubDeviceInfoResponse, error) {
	params := json.RawMessage("{\"requests\":[{\"method\":\"getDeviceInfo\",\"params\":{\"device_info\": {\"name\": [\"basic_info\"]}}}]}")
	var response HubDeviceInfoResponse
	err := h.ExecuteMethod(ctx, "multipleRequest", params, &response)
	return response, err
}

type HubDeviceInfoResponse struct {
	Result struct {
		Responses []struct {
			Method string `json:"method"`
			Result struct {
				DeviceInfo struct {
					BasicInfo struct {
						DeviceType           string `json:"device_type"`
						DeviceModel          string `json:"device_model"`
						DeviceName           string `json:"device_name"`
						DeviceInfo           string `json:"device_info"`
						HwVersion            string `json:"hw_version"`
						SwVersion            string `json:"sw_version"`
						DeviceAlias          string `json:"device_alias"`
						Mac                  string `json:"mac"`
						DevId                string `json:"dev_id"`
						OemId                string `json:"oem_id"`
						HwId                 string `json:"hw_id"`
						Status               string `json:"status"`
						BindStatus           bool   `json:"bind_status"`
						ChildNum             int    `json:"child_num"`
						Avatar               string `json:"avatar"`
						Latitude             int    `json:"latitude"`
						Longitude            int    `json:"longitude"`
						HasSetLocationInfo   int    `json:"has_set_location_info"`
						NeedSyncSha1Password int    `json:"need_sync_sha1_password"`
						ProductName          string `json:"product_name"`
						Region               string `json:"region"`
						LocalIp              string `json:"local_ip"`
					} `json:"basic_info"`
					Info struct {
						DeviceType           string `json:"device_type"`
						DeviceModel          string `json:"device_model"`
						DeviceName           string `json:"device_name"`
						DeviceInfo           string `json:"device_info"`
						HwVersion            string `json:"hw_version"`
						SwVersion            string `json:"sw_version"`
						DeviceAlias          string `json:"device_alias"`
						Mac                  string `json:"mac"`
						DevId                string `json:"dev_id"`
						OemId                string `json:"oem_id"`
						HwId                 string `json:"hw_id"`
						Status               string `json:"status"`
						BindStatus           bool   `json:"bind_status"`
						ChildNum             int    `json:"child_num"`
						Avatar               string `json:"avatar"`
						Latitude             int    `json:"latitude"`
						Longitude            int    `json:"longitude"`
						HasSetLocationInfo   int    `json:"has_set_location_info"`
						NeedSyncSha1Password int    `json:"need_sync_sha1_password"`
						ProductName          string `json:"product_name"`
						Region               string `json:"region"`
						LocalIp              string `json:"local_ip"`
					} `json:"info"`
				} `json:"device_info"`
			} `json:"result"`
			ErrorCode int `json:"error_code"`
		} `json:"responses"`
	} `json:"result"`
	ErrorCode int `json:"error_code"`
}

type ChildDeviceListResponse struct {
	Result struct {
		Responses []struct {
			Method string `json:"method"`
			Result struct {
				StartIndex      int             `json:"start_index"`
				ChildDeviceList json.RawMessage `json:"child_device_list"`
				Sum             int             `json:"sum"`
			} `json:"result"`
			ErrorCode int `json:"error_code"`
		} `json:"responses"`
	} `json:"result"`
	ErrorCode int `json:"error_code"`
}
