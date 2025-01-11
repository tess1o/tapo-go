package tapo

import (
	"encoding/json"
)

type Hub struct {
	*Device
}

func NewHub(host, email, password string, options Options) (*Hub, error) {
	tr, err := NewSslAesTransport(host, email, password, nil)
	if err != nil {
		return nil, err
	}
	tapo := NewDevice(tr, options)
	return &Hub{tapo}, nil
}

func (h *Hub) GetChildDevices() (ChildDeviceListResponse, error) {
	params := json.RawMessage("{\"requests\":[{\"method\":\"getChildDeviceList\",\"params\":{\"childControl\":{\"start_index\":0}}}]}")
	var response ChildDeviceListResponse
	err := h.ExecuteMethod("multipleRequest", params, &response)
	return response, err
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
