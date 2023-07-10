package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RequestClient struct {
	ctx          context.Context
	nodeId       string
	baseUrl      string
	accessToken  string
	refreshToken string
	username     string
	password     string
}

func (client *RequestClient) buildUrl(resource string, params map[string]interface{}) string {
	url := fmt.Sprintf("%s%s", client.baseUrl, resource)
	return url
}

func (client *RequestClient) RequestSync(
	method string,
	path string,
	params map[string]interface{},
	data map[string]interface{},
	withToken bool) map[string]interface{} {
	url := client.buildUrl(path, params)
	reqBodyBytes, _ := json.Marshal(data)
	req, _ := http.NewRequest(method, url, bytes.NewReader(reqBodyBytes))

	if method != http.MethodGet {
		req.Header.Set("content-type", "application/json;charset=utf-8")
	}

	if withToken {
		req.Header.Set("x-token", client.accessToken)
	}

	sender := http.Client{}
	resp, err := sender.Do(req)
	if err != nil {
		fmt.Println("HTTP call failed", err)
		return nil
	}

	respBytes, _ := io.ReadAll(resp.Body)
	newError(fmt.Sprintf("response %s %s from %s", http.MethodPost, url, respBytes)).AtDebug().WriteToLog()
	var result map[string]interface{}
	_ = json.Unmarshal(respBytes, &result)
	return result
}

func (client *RequestClient) Login() {
	data := make(map[string]interface{})
	data["username"] = client.username
	data["password"] = client.password
	var result = client.RequestSync(http.MethodPost, "/auth/login", nil, data, false)
	if result == nil {
		return
	}
	client.accessToken = result["accessToken"].(string)
	client.refreshToken = result["refreshToken"].(string)
	newError("hydra authenticated successfully.").AtInfo().WriteToLog()
}

func (client *RequestClient) RefreshToken() {
	data := make(map[string]interface{})
	data["refreshToken"] = client.refreshToken
	var result = client.RequestSync(http.MethodPost, "/auth/refreshToken", nil, data, true)
	if result == nil {
		return
	}
	client.accessToken = result["accessToken"].(string)
	client.refreshToken = result["refreshToken"].(string)
	newError("hydra token refreshed successfully.").AtInfo().WriteToLog()
}

func (client *RequestClient) UpdateTraffic(hash string, sent uint64, recv uint64) map[string]interface{} {
	data := make(map[string]interface{})
	data["hash"] = hash
	data["sent"] = sent
	data["recv"] = recv
	var result = client.RequestSync(http.MethodPost, "/api/subscriptions/updateTraffic", nil, data, true)
	if result == nil {
		return nil
	}
	newError(result).AtDebug().WriteToLog()
	return result
}

func (client *RequestClient) PullSubscriptions() map[string]interface{} {
	var result = client.RequestSync(http.MethodGet, fmt.Sprintf("/api/nodes/%s/subscriptions", client.nodeId), nil, nil, true)
	if result == nil {
		return nil
	}
	newError(result).AtDebug().WriteToLog()
	return result
}

func NewRequestClient(ctx context.Context, baseUrl string, nodeId string, username string, password string) (*RequestClient, error) {
	client := &RequestClient{
		ctx:      ctx,
		nodeId:   nodeId,
		baseUrl:  baseUrl,
		username: username,
		password: password,
	}
	return client, nil
}
