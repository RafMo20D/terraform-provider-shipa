package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// API endpoints
const (
	apiClusters    = "provisioner/clusters"
	apiPoolsConfig = "pools-config"
	apiPools       = "pools"
	apiApps        = "apps"
	apiUsers       = "users"
	apiPlans       = "plans"
	apiTeams       = "teams"
	apiRoles       = "roles"
)

func apiAppEnvs(appName string) string {
	return fmt.Sprintf("%s/%s/env", apiApps, appName)
}

func apiAppCname(appName string) string {
	return fmt.Sprintf("%s/%s/cname", apiApps, appName)
}

func apiAppDeploy(appName string) string {
	return fmt.Sprintf("%s/%s/deploy", apiApps, appName)
}

func apiRolePermissions(role string) string {
	return fmt.Sprintf("%s/%s/permissions", apiRoles, role)
}

func apiRoleUser(role string) string {
	return fmt.Sprintf("%s/%s/user", apiRoles, role)
}


type Client struct {
	HostURL    string
	HTTPClient *http.Client
	Token      string
}

func NewClient(host, token string) (*Client, error) {
	if host == "" {
		return nil, errors.New("host can not be empty")
	}

	if token == "" {
		return nil, errors.New("token can not be empty")
	}

	c := &Client{
		HostURL:    host,
		HTTPClient: &http.Client{Timeout: 500 * time.Second},
		Token:      token,
	}

	err := c.testAuthentication()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, int, error) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer " + c.Token)

	log.Printf("Headers: %+v\n", req.Header)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	return body, res.StatusCode, err
}

func (c *Client) get(out interface{}, urlPath ...string) error {
	req, err := c.newRequest("GET", nil, urlPath...)
	if err != nil {
		return err
	}

	body, statusCode, err := c.doRequest(req)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return ErrStatus(statusCode, body)
	}
	log.Println("JSON unmarshal:", string(body))
	return json.Unmarshal(body, out)
}

func (c *Client) newURLEncodedRequest(method string, params map[string]string, urlPath ...string) (*http.Request, error) {
	URL := strings.Join(append([]string{c.HostURL}, urlPath...), "/")
	log.Printf("> %s: %s\n", method, URL)

	data := url.Values{}
	for key, val := range params {
		data.Set(key, val)
	}

	//r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	return http.NewRequest(method, URL, strings.NewReader(data.Encode())) // URL-encoded payload
}

func (c *Client) newRequest(method string, payload interface{}, urlPath ...string) (*http.Request, error) {
	var body io.Reader
	URL := strings.Join(append([]string{c.HostURL}, urlPath...), "/")

	log.Printf("> %s: %s\n", method, URL)

	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(data)

		log.Printf("Payload: %s\n", string(data))
	}

	return http.NewRequest(method, URL, body)
}

func (c *Client) newRequestWithParams(method string, payload interface{}, urlPath []string, params map[string]string) (*http.Request, error) {
	var body io.Reader
	URL := strings.Join(append([]string{c.HostURL}, urlPath...), "/")

	paramValues := make([]string, 0)
	for key, val := range params {
		paramValues = append(paramValues, fmt.Sprintf("%s=%s", key, val))
	}
	paramsStr := strings.Join(paramValues, "&")

	if paramsStr != "" {
		URL = fmt.Sprintf("%s?%s", URL, paramsStr)
	}

	log.Printf("> %s: %s\n", method, URL)

	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(data)

		log.Printf("Payload: %s\n", string(data))
	}

	return http.NewRequest(method, URL, body)
}

func (c *Client) newRequestWithParamsList(method string, payload interface{}, urlPath []string, params []*QueryParam) (*http.Request, error) {
	var body io.Reader
	URL := strings.Join(append([]string{c.HostURL}, urlPath...), "/")

	paramValues := make([]string, 0)
	for _, p := range params {
		paramValues = append(paramValues, fmt.Sprintf("%s=%v", p.Key, p.Val))
	}
	paramsStr := strings.Join(paramValues, "&")

	if paramsStr != "" {
		URL = fmt.Sprintf("%s?%s", URL, paramsStr)
	}

	log.Printf("> %s: %s\n", method, URL)

	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(data)

		log.Printf("Payload: %s\n", string(data))
	}

	return http.NewRequest(method, URL, body)
}

func (c *Client) updateRequest(method string, payload interface{}, urlPath ...string) ([]byte, int, error) {
	req, err := c.newRequest(method, payload, urlPath...)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req)
}

func (c *Client) updateURLEncodedRequest(method string, params map[string]string, urlPath ...string) ([]byte, int, error) {
	req, err := c.newURLEncodedRequest(method, params, urlPath...)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.doRequest(req)
}

func (c *Client) post(payload interface{}, urlPath ...string) error {
	body, statusCode, err := c.updateRequest("POST", payload, urlPath...)
	if err != nil {
		return err
	}

	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		return ErrStatus(statusCode, body)
	}
	return nil
}

func (c *Client) postURLEncoded(params map[string]string, urlPath ...string) error {
	body, statusCode, err := c.updateURLEncodedRequest("POST", params, urlPath...)
	if err != nil {
		return err
	}

	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		return ErrStatus(statusCode, body)
	}
	return nil
}

func (c *Client) put(payload interface{}, urlPath ...string) error {
	body, statusCode, err := c.updateRequest("PUT", payload, urlPath...)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return ErrStatus(statusCode, body)
	}
	return nil
}

func (c *Client) delete(urlPath ...string) error {
	req, err := c.newRequest("DELETE", nil, urlPath...)
	if err != nil {
		return err
	}

	body, statusCode, err := c.doRequest(req)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return ErrStatus(statusCode, body)
	}
	return nil
}

type QueryParam struct {
	Key string
	Val interface{}
}

func (c *Client) deleteWithParams(params []*QueryParam, urlPath ...string) error {
	req, err := c.newRequestWithParamsList("DELETE", nil, urlPath, params)
	if err != nil {
		return err
	}

	body, statusCode, err := c.doRequest(req)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return ErrStatus(statusCode, body)
	}
	return nil
}

func (c *Client) deleteWithPayload(payload interface{}, params map[string]string, urlPath ...string) error {
	req, err := c.newRequestWithParams("DELETE", payload, urlPath, params)
	if err != nil {
		return err
	}

	body, statusCode, err := c.doRequest(req)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return ErrStatus(statusCode, body)
	}
	return nil
}

func ErrStatus(statusCode int, body []byte) error {
	return fmt.Errorf("status: %d, body: %s", statusCode, body)
}

func (c *Client) testAuthentication() error {
	_, err := c.ListPlans()
	return err
}
