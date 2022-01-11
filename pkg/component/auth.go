package component

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/TencentBlueKing/gopkg/conv"
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/parnurzeal/gorequest"

	"iam/pkg/logging"
)

// AuthResponse is the struct of iam backend response
type AuthResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// Error will check if the response with error
func (r *AuthResponse) Error() error {
	if r.Code == 0 {
		return nil
	}

	return fmt.Errorf("response error[code=`%d`,  message=`%s`]", r.Code, r.Message)
}

// String will return the detail text of the response
func (r *AuthResponse) String() string {
	return fmt.Sprintf("response[code=`%d`, message=`%s`, data=`%v`]", r.Code, r.Message, r.Data)
}

// AuthClient is the interface of auth client
type AuthClient interface {
	Verify(bkAppCode, bkAppSecret string) (bool, error)
}

type authClient struct {
	Host string

	appCode   string
	appSecret string
}

// NewAuthClient will create a auth client
func NewAuthClient(host string, appCode string, appSecret string) AuthClient {
	host = strings.TrimRight(host, "/")
	return &authClient{
		Host:      host,
		appCode:   appCode,
		appSecret: appSecret,
	}
}

func (c *authClient) call(
	method Method,
	path string,
	data interface{},
	timeout int64,
) (map[string]interface{}, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("component", "authClient.call")

	callTimeout := time.Duration(timeout) * time.Second
	if timeout == 0 {
		callTimeout = defaultTimeout
	}

	url := fmt.Sprintf("%s%s", c.Host, path)
	result := AuthResponse{}
	start := time.Now()
	callbackFunc := NewMetricCallback("Auth", start)

	request := gorequest.New()
	switch method {
	case POST:
		request = request.Post(url)
	case GET:
		request = request.Get(url)
	}
	request = request.Timeout(callTimeout).Type("json")

	// set headers
	request.Header.Set("X-BK-APP-CODE", c.appCode)
	request.Header.Set("X-BK-APP-SECRET", c.appSecret)

	// do request
	resp, respBody, errs := request.
		Send(data).
		EndStruct(&result, callbackFunc)

	// NOTE: it's a sensitive api, so, no log request detail!
	// logFailHTTPRequest(start, request, resp, respBody, errs, &result)
	logger := logging.GetComponentLogger()

	var err error
	if len(errs) != 0 {
		// 敏感信息泄漏 ip+端口号, 替换为 *.*.*.*
		errsMessage := fmt.Sprintf("gorequest errorx=`%s`", errs)
		errsMessage = ipRegex.ReplaceAllString(errsMessage, replaceToIP)
		err = errors.New(errsMessage)

		err = errorWrapf(err, "errsCount=`%d`", len(errs))
		logger.Errorf("call auth api %s fail, err=%s", path, err.Error())
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("gorequest statusCode is %d not 200, respBody=%s",
			resp.StatusCode, conv.BytesToString(respBody))
		logger.Errorf("call auth api %s fail , err=%s", path, err.Error())
		return nil, errorWrapf(err, "status=%d", resp.StatusCode)
	}
	if result.Code != 0 {
		err = errors.New(result.Message)
		err = errorWrapf(err, "result.Code=%d", result.Code)
		logger.Errorf("call auth api %s ok but code in response is not 0, respBody=%s, err=%s",
			path, conv.BytesToString(respBody), err.Error())
		return nil, err
	}
	fmt.Println("result.Data", result.Data)

	return result.Data, nil
}

// Verify will check bkAppCode, bkAppSecret is valid
func (c *authClient) Verify(bkAppCode, bkAppSecret string) (bool, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("component", "authClient.Verify")

	path := fmt.Sprintf("/api/v1/apps/%s/access-keys/verify", bkAppCode)

	data, err := c.call(POST, path, map[string]interface{}{
		"bk_app_secret": bkAppSecret,
	}, 5)
	if err != nil {
		err = errorWrapf(err, "verify app_code=`%s` fail", bkAppCode)

		return false, err
	}
	matchI, ok := data["is_match"]
	if !ok {
		return false, errors.New("no is_match in response body")
	}

	match, ok := matchI.(bool)
	if !ok {
		return false, errors.New("is_match is not a valid bool")
	}
	return match, nil
}
