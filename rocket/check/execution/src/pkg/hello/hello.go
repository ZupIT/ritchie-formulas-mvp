package hello

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ZupIT/ritchie-cli/pkg/prompt"
)

const (
	host = "https://dennis.devdennis.zup.io"
)

type Inputs struct {
	Username    string
	Password    string
	ExecutionID string
	Context     string
}

type loginRequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type loginResponse struct {
	Token string `json:"token,omitempty"`
	TTL   int64  `json:"ttl,omitempty"`
}

type context struct {
	Name string `json:"name,omitempty"`
}

type contexts []context

type formula struct {
	Command string `json:"command,omitempty"`
	Inputs  inputs `json:"inputs,omitempty"`
}

type formulas []formula

type input struct {
	Name    string `json:"name,omitempty"`
	Label   string `json:"label,omitempty"`
	Type    string `json:"type,omitempty"`
	Items   items  `json:"items,omitempty"`
	Default string `json:"default,omitempty"`
	Value   string `json:"value,omitempty"`
}

type items []string

type inputs []input

type commandRequest struct {
	ID      string `json:"id,omitempty"`
	Command string `json:"command,omitempty"`
	Inputs  inputs `json:"inputs,omitempty"`
}

type executionResponse struct {
	Status  string  `json:"status,omitempty"`
	Content content `json:"content,omitempty"`
}

type content struct {
	ID         string   `json:"id,omitempty"`
	StatusCode int      `json:"statusCode,omitempty"`
	User       string   `json:"user,omitempty"`
	StartTime  ExecTime `json:"startTime,omitempty"`
	EndTime    ExecTime `json:"endTime,omitempty"`
	FormulaErr string   `json:"formulaErr,omitempty"`
	FormulaOut string   `json:"formulaOutput,omitempty"`
}

type ExecTime time.Time

func (t ExecTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(t).Unix(), 10)), nil
}

func (t *ExecTime) UnmarshalJSON(s []byte) (err error) {
	r := string(s)
	q, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return err
	}
	*(*time.Time)(t) = time.Unix(q, 0)
	return nil
}

func (t ExecTime) Unix() int64 {
	return time.Time(t).Unix()
}

func (t ExecTime) Time() time.Time {
	return time.Time(t).UTC()
}

func (t ExecTime) String() string {
	return t.Time().String()
}

func (t ExecTime) Sub(o ExecTime) time.Duration {
	return time.Time(t).Sub(time.Time(o))
}

func (in Inputs) Run() {
	// login
	loginResp, err := in.login()
	if err != nil {
		prompt.Error(err.Error())
		os.Exit(1)
	}

	execResp, err := in.execution(loginResp.Token, in.ExecutionID, in.Context)
	if err != nil {
		prompt.Error(err.Error())
		os.Exit(1)
	} else if execResp.Status == "Ready" {
		cont := execResp.Content
		execTime := cont.EndTime.Sub(cont.StartTime)
		prompt.Info(fmt.Sprintf("Execution ID: %s", in.ExecutionID))
		prompt.Info(fmt.Sprintf("Execution time: %s", execTime.String()))
		fmt.Println("-----")
		fmt.Println("stdout:")
		prompt.Info(execResp.Content.FormulaOut)
		fmt.Println("stderr:")
		prompt.Info(execResp.Content.FormulaErr)
	} else {
		prompt.Info("Execution not found or it's being processed")
	}

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

	loginURL := fmt.Sprintf("%s/login", host)
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
		prompt.Success("done")
		return loginResp, err
	case 401:
		return loginResp, errors.New("login failed! Verify your credentials")
	default:
		return loginResp, errors.New("login failed")
	}

}

func (in Inputs) execution(token, ID, ctx string) (executionResponse, error) {
	execResp := executionResponse{}

	execURL := fmt.Sprintf("%s/executions/%s", host, ID)
	req, err := http.NewRequest(http.MethodGet, execURL, nil)
	if err != nil {
		return execResp, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("x-org", "zup")
	req.Header.Set("x-authorization", token)
	req.Header.Set("x-ctx", ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return execResp, fmt.Errorf("error getting execution: %w", err)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return execResp, fmt.Errorf("error reading response: %w", err)
	}

	switch resp.StatusCode {
	case 200:
		if err = json.Unmarshal(b, &execResp); err != nil {
			return execResp, fmt.Errorf("error decoding response: %w", err)
		}
		return execResp, nil
	case 401, 403:
		return execResp, errors.New("authorization failed! Verify your credentials")
	case 404:
		return execResp, nil
	default:
		return execResp, errors.New("error getting execution")
	}
}
