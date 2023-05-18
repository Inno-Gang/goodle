package moodle

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

func (mc *Client) callWsFuncs(
	requests []*wsFuncRequest,
) ([]*wsFuncResponse, error) {
	reqUrl := mc.baseUrl.JoinPath("/webservice/rest/server.php")

	q := reqUrl.Query()
	q.Set("moodlewsrestformat", "json")
	q.Set("wsfunction", "tool_mobile_call_external_functions")
	reqUrl.RawQuery = q.Encode()

	form := url.Values{}
	for i, req := range requests {
		form.Set(fmt.Sprintf("requests[%d][function]", i), req.functionName)
		if len(req.arguments) > 0 {
			form.Set(fmt.Sprintf("requests[%d][arguments]", i), req.arguments)
		}
		form.Set(fmt.Sprintf("requests[%d][settingfilter]", i), "1")
		form.Set(fmt.Sprintf("requests[%d][settingfileurl]", i), "1")
	}
	form.Set("moodlewssettinglang", "en_us")
	form.Set("wsfunction", "tool_mobile_call_external_functions")
	form.Set("wstoken", mc.token)
	body := strings.NewReader(form.Encode())

	resp, err := mc.client.Post(
		reqUrl.String(),
		"application/x-www-form-urlencoded",
		body,
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rawResult wsFuncResponseRaw
	err = json.NewDecoder(resp.Body).Decode(&rawResult)
	if err != nil {
		return nil, err
	}
	responses := make([]*wsFuncResponse, len(rawResult.Responses))
	for i, result := range rawResult.Responses {
		var respErr error
		if result.Error {
			respErr = errors.New("unknown error")
		}
		responses[i] = &wsFuncResponse{
			data:  result.Data,
			error: respErr,
		}
	}
	return responses, nil
}

func (mc *Client) callWsFunc(name string, args string) (string, error) {
	responses, err := mc.callWsFuncs([]*wsFuncRequest{{
		functionName: name,
		arguments:    args,
	}})
	if err != nil {
		return "", err
	}
	resp := responses[0]
	return resp.data, resp.error
}

type wsFuncRequest struct {
	functionName string
	arguments    string
}

type wsFuncResponse struct {
	data  string
	error error
}

type wsFuncResponseRaw struct {
	Responses []struct {
		Data  string `json:"data"`
		Error bool   `json:"error"`
	} `json:"responses"`
}
