package marketo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	// DefaultTimeout is http client timeout and 60 seconds
	DefaultTimeout = 60
	identityBase   = "/identity"
	identityPath   = "/oauth/token"
)

// RecordResult holds Marketo record-level result
type RecordResult struct {
	ID      int    `json:"id"`
	Status  string `json:"status"`
	Reasons []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"reasons,omitempty"`
}

// Response is the common Marketo response which covers most of the Marketo response format
type Response struct {
	RequestID     string `json:"requestId"`
	Success       bool   `json:"success"`
	NextPageToken string `json:"nextPageToken,omitempty"`
	MoreResult    bool   `json:"moreResult,omitempty"`
	Errors        []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"errors,omitempty"`
	Result   json.RawMessage `json:"result,omitempty"`
	Warnings []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"warning,omitempty"`
}

// AuthToken holds data from Auth request
type AuthToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// Client Marketo http Client
type Client struct {
	authClient       *http.Client
	restClient       *http.Client
	restRoundTripper *restRoundTripper
	endpoint         string
	identityEndpoint string
	authLock         sync.Mutex
	auth             *AuthToken
	tokenExpiresAt   time.Time
	debug            bool
}

// authRoundTripper wrapper for authentication query params
type authRoundTripper struct {
	delegate     http.RoundTripper
	clientID     string
	clientSecret string
}

func (rt *authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.delegate == nil {
		rt.delegate = http.DefaultTransport
	}
	values := req.URL.Query()
	values.Add("client_id", rt.clientID)
	values.Add("client_secret", rt.clientSecret)
	values.Add("grant_type", "client_credentials")
	req.URL.RawQuery = values.Encode()
	return rt.delegate.RoundTrip(req)
}

// restRoundTripper wrapper for adding bearer token
type restRoundTripper struct {
	delegate http.RoundTripper
	token    string
}

func (rt *restRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.delegate == nil {
		rt.delegate = http.DefaultTransport
	}
	req.Header.Add("Authorization", "Bearer "+rt.token)
	return rt.delegate.RoundTrip(req)
}

// ClientConfig stores client configuration
type ClientConfig struct {
	// ID: Marketo client ID
	ID string
	// Secret: Marketo client secret
	Secret string
	// Endpoint: https://xxx-xxx-xxx.mktorest.com
	Endpoint string
	// Timeout, optional: default http timeout is 60 seconds
	Timeout uint
	// Debug, optional: a flag to show logging output
	Debug bool
}

// NewClient returns a new Marketo Client
func NewClient(config ClientConfig) (*Client, error) {
	// create two roundtrippers
	aRT := authRoundTripper{
		clientID:     config.ID,
		clientSecret: config.Secret,
	}
	rRT := restRoundTripper{}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	// Add credentials to the request
	c := &Client{
		authClient: &http.Client{
			Timeout:   time.Second * time.Duration(timeout),
			Transport: &aRT,
		},
		restClient: &http.Client{
			Timeout:   time.Second * time.Duration(timeout),
			Transport: &rRT,
		},
		restRoundTripper: &rRT,
		endpoint:         config.Endpoint,
		identityEndpoint: config.Endpoint + identityBase + identityPath,
		debug:            config.Debug,
	}

	if _, err := c.RefreshToken(); err != nil {
		return nil, err
	}
	return c, nil
}

// RefreshToken refreshes the auth token.
// This is purely for testing purpose and not intended to be used.
func (c *Client) RefreshToken() (auth AuthToken, err error) {
	if c.debug {
		log.Printf("[marketo/RefreshToken] start")
		defer func() {
			log.Print("[marketo/RefreshToken] DONE")
		}()
	}
	// Make request for token
	resp, err := c.authClient.Get(c.identityEndpoint)
	if err != nil {
		return auth, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return auth, err
		}
		return auth, fmt.Errorf("Authentication error: %d %s", resp.StatusCode, body)
	}

	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		return auth, err
	}
	if c.debug {
		log.Printf("[marketo/RefreshToken] New token: %v", auth)
	}
	c.authLock.Lock()
	defer c.authLock.Unlock()
	c.auth = &auth
	c.restRoundTripper.token = auth.AccessToken
	c.tokenExpiresAt = time.Now().Add(time.Duration(auth.ExpiresIn) * time.Second)
	return auth, nil
}

func (c *Client) url(paths ...string) string {
	return fmt.Sprintf("%s/%s", c.endpoint, strings.Join(paths, "/"))
}

func (c *Client) do(req *http.Request) (response *Response, err error) {
	var body []byte
	if c.debug {
		log.Printf("[marketo/do] URL: %s", req.URL)
		defer func() {
			log.Printf("[marketo/do] DONE: body %s", string(body))
		}()
	}
	resp, err := c.restClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Unexpected status code[%d] with body[%s]", resp.StatusCode, string(body))
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("No body! Check URL: %s", req.URL)
	}

	if err = json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) doWithRetry(req *http.Request) (response *Response, err error) {
	// check if token has been expired or not
	if c.tokenExpiresAt.Before(time.Now()) {
		if c.debug {
			log.Printf("[marketo/doWithRetry] token expired at: %s", c.tokenExpiresAt.String())
		}
		c.RefreshToken()
	}

	response, err = c.do(req)
	if err != nil {
		return nil, err
	}

	// check just in case we received 601 or 602
	retry, err := c.checkToken(response)
	if err != nil {
		return nil, err
	}
	if retry {
		response, err = c.do(req)
	}

	return response, err
}

func (c *Client) doRequest(req *http.Request) (response *http.Response, err error) {
	// check if token has been expired or not
	if c.tokenExpiresAt.Before(time.Now()) {
		if c.debug {
			log.Printf("[marketo/doWithRetry] token expired at: %s", c.tokenExpiresAt.String())
		}
		c.RefreshToken()
	}

	response, err = c.restClient.Do(req)
	if err != nil {
		return nil, err
	}

	// check just in case we received 601 or 602
	// retry, err := c.checkToken(response)
	// if err != nil {
	// 	return nil, err
	// }
	// if retry {
	// 	response, err = c.do(req)
	// }

	return response, err
}

func (c *Client) checkToken(response *Response) (retry bool, err error) {
	if len(response.Errors) > 0 && (response.Errors[0].Code == "601" || response.Errors[0].Code == "602") {
		retry = true
		if c.debug {
			log.Printf("[marketo/checkToken] Expired/invalid token: %s", response.Errors[0].Code)
		}
		_, err = c.RefreshToken()
	}
	return retry, err
}

// Send HTTP GET to resource url
func (c *Client) Get(resource string) (response *Response, err error) {
	if c.debug {
		log.Printf("[marketo/Get] %s", resource)
		defer func() {
			log.Print("[marketo/Get] DONE")
		}()
	}
	req, err := http.NewRequest("GET", c.endpoint+resource, nil)
	if err != nil {
		return nil, err
	}
	return c.doWithRetry(req)
}

// Send HTTP POST to resource url with given data
func (c *Client) Post(resource string, data []byte) (response *Response, err error) {
	if c.debug {
		log.Printf("[marketo/Post] %s, %s", resource, string(data))
		defer func() {
			log.Print("[marketo/Post] DONE")
		}()
	}
	req, err := http.NewRequest("POST", c.endpoint+resource, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doWithRetry(req)
}

// Send HTTP DELETE to resource url with given data
func (c *Client) Delete(resource string, data []byte) (response *Response, err error) {
	if c.debug {
		log.Printf("[marketo/Delete] %s, %s", resource, string(data))
		defer func() {
			log.Print("[marketo/Delete] DONE")
		}()
	}
	req, err := http.NewRequest("DELETE", c.endpoint+resource, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doWithRetry(req)
}

// TokenInfo holds authentication token and time at which expires.
type TokenInfo struct {
	// Token is the currently active token.
	Token string
	// Expires shows what time the token expires
	Expires time.Time
}

// GetTokenInfo returns current TokenInfo stored in Client
func (c *Client) GetTokenInfo() TokenInfo {
	return TokenInfo{c.auth.AccessToken, c.tokenExpiresAt}
}
