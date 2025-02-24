package tapo

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SslAesTransport struct {
	host        string
	username    string
	password    string
	pwdHash     string
	localNonce  string
	serverNonce string
	digestPwd   string
	stok        string
	seq         int
	encryption  *AES
	httpClient  *http.Client
	retryConfig *RetryConfig
}

var defaultHttpTransport = &http.Transport{
	TLSClientConfig: &tls.Config{
		CipherSuites: []uint16{
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		},
		InsecureSkipVerify: true,
	},
}

func NewSslAesTransport(ctx context.Context, host, user, password string, options Options) (*SslAesTransport, error) {
	client := options.HttpClient
	if client == nil {
		client = &http.Client{Transport: defaultHttpTransport}
	}

	if !strings.Contains(host, ":") {
		host = host + ":443"
	}

	transport, err := handshake(ctx, host, user, password, client)
	if err != nil {
		return nil, err
	}

	transport.retryConfig = options.RetryConfig

	return transport, nil
}

func handshake(ctx context.Context, host string, user string, password string, client *http.Client) (*SslAesTransport, error) {
	s := &SslAesTransport{
		host:       host,
		username:   user,
		password:   password,
		pwdHash:    sha256HashUpperCase([]byte(password)),
		httpClient: client,
	}
	ln, err := s.generateLocalNonce()
	if err != nil {
		return nil, err
	}
	s.localNonce = ln
	time.Sleep(200 * time.Millisecond)
	handshake1, err := s.handshake1(ctx)
	if err != nil {
		return nil, err
	}
	s.serverNonce = handshake1.Result.Data.Nonce
	s.digestPwd = s.generateDigestPassword()
	time.Sleep(200 * time.Millisecond)
	handshake2, err := s.handshake2(ctx)
	if err != nil {
		return nil, err
	}
	s.stok = handshake2.Result.Stok
	s.seq = handshake2.Result.StartSeq
	time.Sleep(200 * time.Millisecond)

	key := s.GenerateEncryptionToken("lsk")
	iv := s.GenerateEncryptionToken("ivb")

	encryption, err := NewAES(key, iv)
	if err != nil {
		return nil, err
	}

	s.encryption = encryption

	return s, nil
}

func (t *SslAesTransport) generateDigestPassword() string {
	digestPasswordHash := sha256HashUpperCase([]byte(t.pwdHash + t.localNonce + t.serverNonce))
	return digestPasswordHash + t.localNonce + t.serverNonce
}

func (t *SslAesTransport) GenerateEncryptionToken(tokenType string) []byte {
	hashedKey := sha256HashUpperCase([]byte(t.localNonce + t.pwdHash + t.serverNonce))
	finalPayload := []byte(tokenType + t.localNonce + t.serverNonce + hashedKey)
	finalHash := sha256Hash(finalPayload)
	return finalHash[:16]
}

func (t *SslAesTransport) GenerateTag(request string, seq int) string {
	pwdNonceHash := sha256HashUpperCase([]byte(t.pwdHash + t.localNonce))
	tag := sha256HashUpperCase([]byte(pwdNonceHash + request + strconv.Itoa(seq)))
	return tag
}

func (t *SslAesTransport) generateLocalNonce() (string, error) {
	localNonce := make([]byte, 8)
	_, err := rand.Read(localNonce)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(localNonce)), nil
}

func setHeaders(req *http.Request, host string) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Host", host+":443")
	req.Header.Set("Referer", "https://"+host)
	req.Header.Set("Requestbyapp", "true")
	req.Header.Set("User-Agent", "Tapo CameraClient Android")
}

func (t *SslAesTransport) handshake1(ctx context.Context) (*Handshake1Response, error) {
	requestBody := Handshake1Request{
		Method: "login",
		Params: Handshake1RequestParams{
			Cnonce:      t.localNonce,
			EncryptType: "3",
			Username:    "admin",
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Println("Error marshalling JSON:", err)
		return nil, err
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://"+t.host, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating request in handhsake1:", err)
		return nil, err
	}

	// Set request headers
	setHeaders(req, t.host)

	// Make the HTTP request
	resp, err := t.httpClient.Do(req)
	if err != nil {
		log.Println("Error making request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return nil, err
	}
	responseBody := Handshake1Response{}
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return nil, err
	}
	return &responseBody, nil
}

func (t *SslAesTransport) handshake2(ctx context.Context) (*Handshake2Response, error) {
	requestBody := Handshake2Request{
		Method: "login",
		Params: Handshake2RequestParams{
			Cnonce:       t.localNonce,
			EncryptType:  "3",
			DigestPasswd: t.digestPwd,
			Username:     "admin",
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Println("Error marshalling JSON:", err)
		return nil, err
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://"+t.host, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating request:", err)
		return nil, err
	}

	// Set request headers
	setHeaders(req, t.host)

	// Make the HTTP request
	resp, err := t.httpClient.Do(req)
	if err != nil {
		log.Println("Error making request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return nil, err
	}
	responseBody := Handshake2Response{}
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return nil, err
	}
	return &responseBody, nil
}

func (t *SslAesTransport) ExecuteRequest(ctx context.Context, request *RequestSpec) (json.RawMessage, error) {
	return ExecuteHttpRequest(ctx, t, request, t.retryConfig)
}

func (t *SslAesTransport) executeHttpRequest(ctx context.Context, rr *RequestSpec) ([]byte, int, error) {
	multiRequestBody, err := json.Marshal(rr)
	if err != nil {
		return nil, -1, err
	}
	encryptedParams, err := t.encryption.Encrypt(multiRequestBody)

	if err != nil {
		return nil, -1, err
	}

	apiRequest := SecurePassThroughRequest{
		Method: "securePassthrough",
		Params: struct {
			Request string `json:"request"`
		}{
			Request: encryptedParams,
		},
	}

	apiRequestBody, err := json.Marshal(apiRequest)
	if err != nil {
		log.Println("Error marshalling JSON:", err)
		return nil, -1, err
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://"+t.host+"/stok="+t.stok+"/ds", bytes.NewBuffer(apiRequestBody))
	if err != nil {
		log.Println("Error creating request:", err)
		return nil, -1, err
	}

	tag := t.GenerateTag(string(apiRequestBody), t.seq)
	setHeaders(req, t.host)

	req.Header.Set("Seq", strconv.Itoa(t.seq))
	req.Header.Set("Tapo_tag", tag)

	// Make the HTTP request
	resp, err := t.httpClient.Do(req)
	if err != nil {
		log.Println("Error making request:", err)
		return nil, -1, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return nil, -1, err
	}

	responseBody := SecurePassThroughResponse{}
	err = json.Unmarshal(body, &responseBody)
	if err != nil || responseBody.ErrorCode != 0 {
		var errorResponseWithCode ErrorResponseWithCode

		if err = json.Unmarshal(body, &errorResponseWithCode); err == nil && errorResponseWithCode.Result.Data.Code != 0 {
			return nil, -1, errors.New(fmt.Sprintf("request exited with error_code: %d, code: %+v", errorResponseWithCode.ErrorCode, errorResponseWithCode.Result.Data.Code))
		}
	}

	decryptedResponse, err := t.encryption.Decrypt(responseBody.Result.Response)
	if err != nil {
		return nil, -1, err
	}

	var errorResponse ErrorResponse

	if err = json.Unmarshal(decryptedResponse, &errorResponse); err == nil && errorResponse.ErrCode != 0 {
		return nil, -1, errors.New(fmt.Sprintf("request exited with err_code: %d, err_msg: %+v", errorResponse.ErrCode, errorResponse.ErrMsg))
	}

	var errorMethodResponse ErrorMethodResponse

	if err = json.Unmarshal(decryptedResponse, &errorMethodResponse); err == nil && len(errorMethodResponse.Result.Responses) > 0 && errorMethodResponse.Result.Responses[0].ErrorCode != 0 {
		return nil, -1, errors.New(fmt.Sprintf("request exited with error_code: %d, method: %+v", errorMethodResponse.Result.Responses[0].ErrorCode, errorMethodResponse.Result.Responses[0].Method))
	}

	t.seq++
	return decryptedResponse, resp.StatusCode, nil
}

type Handshake1Request struct {
	Method string                  `json:"method"`
	Params Handshake1RequestParams `json:"params"`
}

type Handshake1RequestParams struct {
	Cnonce      string `json:"cnonce"`
	EncryptType string `json:"encrypt_type"`
	Username    string `json:"username"`
}

type Handshake1Response struct {
	ErrorCode int `json:"error_code"`
	Result    struct {
		Data struct {
			Code          int      `json:"code"`
			EncryptType   []string `json:"encrypt_type"`
			Key           string   `json:"key"`
			Nonce         string   `json:"nonce"`
			DeviceConfirm string   `json:"device_confirm"`
		} `json:"data"`
	} `json:"result"`
}

type Handshake2Request struct {
	Method string                  `json:"method"`
	Params Handshake2RequestParams `json:"params"`
}

type Handshake2RequestParams struct {
	Cnonce       string `json:"cnonce"`
	EncryptType  string `json:"encrypt_type"`
	DigestPasswd string `json:"digest_passwd"`
	Username     string `json:"username"`
}

type Handshake2Response struct {
	ErrorCode int `json:"error_code"`
	Result    struct {
		Stok      string `json:"stok"`
		UserGroup string `json:"user_group"`
		StartSeq  int    `json:"start_seq"`
	} `json:"result"`
}

type SecurePassThroughRequest struct {
	Method string `json:"method"`
	Params struct {
		Request string `json:"request"`
	} `json:"params"`
}

type SecurePassThroughResponse struct {
	ErrorCode int `json:"error_code"`
	Seq       int `json:"seq"`
	Result    struct {
		Response string `json:"response"`
	} `json:"result"`
}

type ErrorResponse struct {
	ErrCode int    `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
}

type ErrorMethodResponse struct {
	Result struct {
		Responses []struct {
			Method    string `json:"method"`
			ErrorCode int    `json:"error_code"`
		} `json:"responses"`
	} `json:"result"`
	ErrorCode int `json:"error_code"`
}

type ErrorResponseWithCode struct {
	ErrorCode int `json:"error_code"`
	Result    struct {
		Data struct {
			Code int `json:"code"`
		} `json:"data"`
	} `json:"result"`
}
