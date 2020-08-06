package formula

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ZupIT/ritchie-cli/pkg/prompt"
)

const (
	host      = "https://dennis.devdennis.zup.io"
	loginPath = "%s/login"
)

type Inputs struct {
	Username string
	Password string
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
	TTL   int64  `json:"ttl"`
}

func (in Inputs) Run() {
	loginResp, err := in.login()
	if err != nil {
		prompt.Error(err.Error())
		os.Exit(1)
	}

	prompt.Info(fmt.Sprintf("%v", loginResp))

}

func (in Inputs) Formulas() {

}

func (in Inputs) login() (loginResponse, error) {
	prompt.Info("Authenticating...")

	loginResp := loginResponse{}
	loginReq := loginRequest{
		in.Username,
		in.Password,
	}
	b, err := json.Marshal(&loginReq)
	if err != nil {
		return loginResp, fmt.Errorf("error encoding credential: %w", err)
	}

	loginURL := fmt.Sprintf(loginPath, host)
	req, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewBuffer(b))
	if err != nil {
		return loginResp, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-org", "zup")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return loginResp, fmt.Errorf("error performing login: %w", err)
	}

	defer resp.Body.Close()

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return loginResp, fmt.Errorf("error reading response: %w", err)
	}

	switch resp.StatusCode {
	case 200:
		if err = json.Unmarshal(b, &loginResp); err != nil {
			return loginResp, fmt.Errorf("error decoding response: %w", err)
		}
		prompt.Info("done")
		return loginResp, err
	case 401:
		return loginResp, errors.New("login failed! Verify your credentials")
	default:
		return loginResp, errors.New("login failed")
	}

}
