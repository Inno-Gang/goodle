package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"

	"github.com/inno-gang/goodle/moodle"
)

type IuCredentials struct {
	Email    string
	Password string
}

type IuAuthenticator struct {
	baseUrl string
}

const iuMoodleBaseUrl = "https://moodle.innopolis.university"

func NewIuAuthenticator() *IuAuthenticator {
	return &IuAuthenticator{baseUrl: iuMoodleBaseUrl}
}

const moodleLoginPageUrl = "https://moodle.innopolis.university/admin/tool/mobile/launch.php?service=moodle_mobile_app&passport=12.34567890&urlscheme=moodlemobile"

var loginRedirectLinkRegex = regexp.MustCompile(`href="(https://moodle\.innopolis\.university/auth/oauth2/login\.php\?[^"]+)"`)
var innoAuthLinkPathRegex = regexp.MustCompile(`action="(/adfs/oauth2/authorize[^"]+)"`)

func (ia *IuAuthenticator) Authenticate(
	ctx context.Context,
	credentials IuCredentials,
) (*moodle.Client, error) {
	jar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		Jar: jar,
	}

	moodleLoginPage, err := getHtml(ctx, httpClient, moodleLoginPageUrl)
	if err != nil {
		return nil, err
	}

	authRedirectUrl, err := applyRegexOnHtml(loginRedirectLinkRegex, moodleLoginPage)
	if err != nil {
		return nil, err
	}

	innoAuthPage, err := getHtml(ctx, httpClient, authRedirectUrl)
	if err != nil {
		return nil, err
	}

	innoAuthPath, err := applyRegexOnHtml(innoAuthLinkPathRegex, innoAuthPage)
	if err != nil {
		return nil, err
	}

	token, err := loginInnoObtainWsToken(
		ctx,
		httpClient,
		innoAuthPath,
		credentials.Email,
		credentials.Password,
	)
	if err != nil {
		return nil, err
	}

	return moodle.NewClient(
		ia.baseUrl,
		token,
		httpClient,
	)
}

func getHtml(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func applyRegexOnHtml(regex *regexp.Regexp, htmlContent string) (string, error) {
	match := regex.FindStringSubmatch(htmlContent)
	if match == nil {
		return "", errors.New("no match")
	}
	return html.UnescapeString(match[1]), nil
}

const innoAuthBaseUrl = "https://sso.university.innopolis.ru"

func loginInnoObtainWsToken(
	ctx context.Context,
	client *http.Client,
	authPath string,
	username string,
	password string,
) (string, error) {
	innoAuthLink := innoAuthBaseUrl + authPath
	formData := getFormUrlEncodedForAuth(username, password)

	req, err := http.NewRequestWithContext(ctx, "POST", innoAuthLink, strings.NewReader(formData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	submitTokenPageHtml := string(data)

	tokenEncoded, err := submitTokenPage(ctx, client, submitTokenPageHtml)
	if err != nil {
		return "", err
	}

	token, err := base64.StdEncoding.DecodeString(tokenEncoded)
	if err != nil {
		return "", err
	}

	parts := strings.Split(string(token), ":::")
	if len(parts) < 2 {
		return "", errors.New("expected >=2 parts in the decoded token")
	}

	return parts[1], nil
}

func getFormUrlEncodedForAuth(username string, password string) string {
	form := url.Values{}
	form.Set("UserName", username)
	form.Set("Password", password)
	form.Set("Kmsi", "true")
	form.Set("AuthMethod", "FormsAuthentication")
	return form.Encode()
}

var moodleMobileTokenRegex = regexp.MustCompile(`moodlemobile://token=(\S+)\s*$`)

func submitTokenPage(
	ctx context.Context,
	client *http.Client,
	pageHtml string,
) (string, error) {
	actionUrl, code, state, err := parsePartsFromSubmitTokenPage(pageHtml)
	if err != nil {
		return "", err
	}
	form := url.Values{}
	form.Set("code", code)
	form.Set("state", state)

	var token string
	oldCheckRedirect := client.CheckRedirect
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		reqUrl := req.URL.String()
		if match := moodleMobileTokenRegex.FindStringSubmatch(reqUrl); match != nil {
			token = match[1]
		}
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, "POST", actionUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	client.CheckRedirect = oldCheckRedirect
	if err == nil {
		defer resp.Body.Close()
		err = errors.New("failed to find token in redirect url")
	}

	if token == "" {
		return "", err
	}

	return token, nil
}

var submitTokenFormActionRegex = regexp.MustCompile(`action="([^"]+)"`)
var submitTokenFormCodeFieldRegex = regexp.MustCompile(`name="code"\s+value="([^"]+)"`)
var submitTokenFormStateFieldRegex = regexp.MustCompile(`name="state"\s+value="([^"]+)"`)

func parsePartsFromSubmitTokenPage(pageHtml string) (
	action string,
	code string,
	state string,
	err error,
) {
	action, err = applyRegexOnHtml(submitTokenFormActionRegex, pageHtml)
	if err != nil {
		err = fmt.Errorf("failed to find action pattern in token submit page: %w", err)
		return
	}
	code, err = applyRegexOnHtml(submitTokenFormCodeFieldRegex, pageHtml)
	if err != nil {
		err = fmt.Errorf("failed to find code field pattern in token submit page: %w", err)
		return
	}
	state, err = applyRegexOnHtml(submitTokenFormStateFieldRegex, pageHtml)
	if err != nil {
		err = fmt.Errorf("failed to find state field pattern in token submit page: %w", err)
		return
	}
	return
}
