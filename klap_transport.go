package tapo

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type KlapTransport struct {
	Username    string
	Password    string
	Host        string
	httpClient  *http.Client
	retryConfig *RetryConfig

	Cookies []*http.Cookie
	Session *KlapEncryptionSession
}

func NewKlapTransport(ctx context.Context, username, password, host string, options Options) (*KlapTransport, error) {
	var client = options.HttpClient
	if client == nil {
		client = DefaultHttpClient
	}

	if !strings.Contains(host, ":") {
		host = host + ":80"
	}

	tr := &KlapTransport{
		Username:    username,
		Password:    password,
		Host:        host,
		httpClient:  client,
		retryConfig: options.RetryConfig,
	}
	err := tr.handshake(ctx)
	if err != nil {
		return nil, err
	}
	return tr, nil
}

func (k *KlapTransport) generateAuthHashV2() []byte {
	emailHash := sha1.New()
	passwordHash := sha1.New()

	emailHash.Write([]byte(k.Username))
	emailHashBytes := emailHash.Sum(nil)

	passwordHash.Write([]byte(k.Password))
	passwordHashBytes := passwordHash.Sum(nil)

	mixedHashBytes := append(emailHashBytes, passwordHashBytes...)
	finalHashBytes := sha256.Sum256(mixedHashBytes)

	return finalHashBytes[:]
}

func (k *KlapTransport) generateSeedAuthHash(localSeed []byte, remoteSeed []byte, authHash []byte, handshakeStage int) []byte {
	var finalHashContentBytes []byte

	switch handshakeStage {
	case 1:
		finalHashContentBytes = append(localSeed, remoteSeed...)
	case 2:
		finalHashContentBytes = append(remoteSeed, localSeed...)
	}

	finalHashContentBytes = append(finalHashContentBytes, authHash...)

	finalHashBytes := sha256.Sum256(finalHashContentBytes)
	return finalHashBytes[:]
}

func (k *KlapTransport) handshake1(ctx context.Context) ([]byte, []byte, error) {
	localSeed := make([]byte, 16)
	_, err := rand.Read(localSeed)
	if err != nil {
		return nil, nil, fmt.Errorf("error while generating random string: %s", err)
	}

	u, err := url.Parse(fmt.Sprintf("http://%s/app/handshake1", k.Host))
	if err != nil {
		return nil, nil, err
	}

	bodyBytesReader := bytes.NewBuffer(localSeed)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bodyBytesReader)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating HTTP request: %s", err)
	}

	response, err := k.httpClient.Do(request)
	if err != nil {
		return nil, nil, fmt.Errorf("error making HTTP request: %s", err)
	}

	defer response.Body.Close()

	// Check request status
	if response.StatusCode != 200 {
		return nil, nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	k.Cookies = response.Cookies()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading response body: %s", err)
	}

	// Recover results from server
	remoteSeed := bodyBytes[0:16]

	return localSeed, remoteSeed, nil
}

func (k *KlapTransport) handshake2(ctx context.Context, localSeed, remoteSeed []byte) error {
	authHash := k.generateAuthHashV2()
	remoteSeedAuthHash := k.generateSeedAuthHash(localSeed, remoteSeed, authHash, 2)

	// Create URL for handshake2
	u, err := url.Parse(fmt.Sprintf("http://%s/app/handshake2", k.Host))
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(remoteSeedAuthHash))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %s", err)
	}

	for _, cookie := range k.Cookies {
		request.AddCookie(cookie)
	}
	response, err := k.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("error making HTTP request: %s", err)
	}

	defer response.Body.Close()
	httpResponseBody, err := io.ReadAll(response.Body)
	if response.StatusCode != 200 {
		return fmt.Errorf("handshake 2 failed with status code: %d, response: %s", response.StatusCode, httpResponseBody)
	}

	k.Session = NewKlapEncryptionSession(
		string(localSeed),
		string(remoteSeed),
		string(authHash))

	return nil
}

func (k *KlapTransport) handshake(ctx context.Context) error {
	// Perform first stage of handshake phase
	// The mission here is to get a remote seed and cookies
	localSeed, remoteSeed, err := k.handshake1(ctx)
	if err != nil {
		return err
	}

	time.Sleep(200 * time.Millisecond)

	// Perform second stage of handshake phase
	// The mission here is to get a KLAP encryption session
	err = k.handshake2(ctx, localSeed, remoteSeed)
	if err != nil {
		return err
	}

	time.Sleep(200 * time.Millisecond)
	return nil
}

func (k *KlapTransport) ExecuteRequest(ctx context.Context, request *RequestSpec) (response json.RawMessage, err error) {
	return ExecuteHttpRequest(ctx, k, request, k.retryConfig)
}

func (k *KlapTransport) executeHttpRequest(ctx context.Context, request *RequestSpec) ([]byte, int, error) {
	httpRequest, err := k.prepareRequest(ctx, request)
	if err != nil {
		return nil, -1, err
	}

	httpResponse, err := k.httpClient.Do(httpRequest)
	if err != nil {
		return nil, -1, err
	}
	defer httpResponse.Body.Close()

	httpResponseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, -1, err
	}

	if httpResponse.StatusCode != 200 {
		return httpResponseBody, httpResponse.StatusCode, errors.New(fmt.Sprintf("request exited with failed status: %d", httpResponse.StatusCode))
	}

	decryptedResponseBody := k.Session.decrypt(httpResponseBody)
	return decryptedResponseBody, httpResponse.StatusCode, nil
}

func (k *KlapTransport) prepareRequest(ctx context.Context, request *RequestSpec) (*http.Request, error) {
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	jsonRequest := string(jsonBytes)

	encryptedPayload, seq := k.Session.encrypt(jsonRequest)

	u, err := url.Parse(fmt.Sprintf("http://%s/app/request?seq=%d", k.Host, seq))
	if err != nil {
		return nil, err
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(encryptedPayload))
	if err != nil {
		return nil, err
	}

	httpRequest.Header.Set("Content-Type", "application/json")
	for _, cookie := range k.Cookies {
		httpRequest.AddCookie(cookie)
	}
	return httpRequest, nil
}
