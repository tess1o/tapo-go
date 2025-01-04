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

func (h *Hub) GetChildDevices() (json.RawMessage, error) {
	params := json.RawMessage("{\"requests\":[{\"method\":\"getChildDeviceList\",\"params\":{\"childControl\":{\"start_index\":0}}}]}")
	var response json.RawMessage
	err := h.ExecuteMethod("multipleRequest", params, &response)
	return response, err
}
