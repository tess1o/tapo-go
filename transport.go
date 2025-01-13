package tapo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

type httpTransport interface {
	executeHttpRequest(ctx context.Context, request *RequestSpec) ([]byte, int, error)
}

func ExecuteHttpRequest(ctx context.Context, transport httpTransport, req *RequestSpec, retryConfig *RetryConfig) (json.RawMessage, error) {
	var responseBody json.RawMessage
	var statusCode int
	var err error

	var retryCount int
	if retryConfig == nil {
		retryCount = 1
	} else {
		retryCount = retryConfig.RetryCount
	}

	for retries := 0; retries <= retryCount; retries++ {
		select {
		case <-ctx.Done():
			return nil, errors.New("context canceled")
		default:
			responseBody, statusCode, err = transport.executeHttpRequest(ctx, req)
			if err == nil && statusCode == 200 {
				return responseBody, nil
			}
			log.Printf("Request failed (attempt %d): status code: %d, error: %v", retries, statusCode, err)

			if retries == retryCount || (retryConfig != nil && retryConfig.Retry403ErrorsOnly && statusCode != 403) {
				return nil, fmt.Errorf("request exited with failed status: %d, error: %v", statusCode, err)
			}

			select {
			case <-ctx.Done():
				return nil, errors.New("context canceled")
			default:
			}
		}
	}
	return nil, fmt.Errorf("request failed after retries")
}
