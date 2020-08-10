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
	Username         string
	Password         string
	Provider         string
	ProviderUsername string
	ProviderSecret   string
}

type loginRequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type loginResponse struct {
	Token string `json:"token,omitempty"`
	TTL   int64  `json:"ttl,omitempty"`
}

type formulasResponse struct {
	Contexts contexts `json:"contexts,omitempty"`
	Formulas formulas `json:"formulas,omitempty"`
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

	text := prompt.NewSurveyText()
	pass := prompt.NewSurveyPassword()

	var username string
	var secret string
	if in.Provider == "github" {
		username, _ = text.Text("Username", true)
		secret, _ = pass.Password("Token")

	} else {
		username, _ = text.Text("AccessKeyID", true)
		secret, _ = pass.Password("SecretAccessKey")
	}

	in.ProviderUsername = username
	in.ProviderSecret = secret

	// login
	loginResp, err := in.login()
	if err != nil {
		prompt.Error(err.Error())
		os.Exit(1)
	}

	// formulas e context
	formulasResp, err := in.formulas(loginResp.Token)
	if err != nil {
		prompt.Error(err.Error())
		os.Exit(1)
	}

	list := prompt.NewSurveyList()

	ctxList := make([]string, len(formulasResp.Contexts))
	for i, c := range formulasResp.Contexts {
		ctxList[i] = c.Name
	}
	ctx, err := list.List("Select a context", ctxList)
	if err != nil {
		prompt.Error(err.Error())
		os.Exit(1)
	}

	// set credential
	err = in.setCredential(loginResp.Token, ctx)
	if err != nil {
		prompt.Error(err.Error())
		os.Exit(1)
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

func (in Inputs) formulas(token string) (formulasResponse, error) {
	prompt.Info("Obtaining formulas...")

	formulasResp := formulasResponse{}

	formulasURL := fmt.Sprintf("%s/formulas", host)
	req, err := http.NewRequest(http.MethodGet, formulasURL, nil)
	if err != nil {
		return formulasResp, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("x-org", "zup")
	req.Header.Set("x-authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return formulasResp, fmt.Errorf("error obtaining formulas: %w", err)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return formulasResp, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode == 200 {
		if err = json.Unmarshal(b, &formulasResp); err != nil {
			return formulasResp, fmt.Errorf("error decoding response: %w", err)
		}
		prompt.Success("done")
		return formulasResp, err
	} else {
		return formulasResp, errors.New("error obtaining formulas")
	}
}

func (in Inputs) setCredential(token, ctx string) error {
	prompt.Info("Authenticating...")

	type CredentialAWS struct {
		AccessKeyId     string `json:"accesskeyid"`
		SecretAccessKey string `json:"secretaccesskey"`
	}

	type CredentialGithub struct {
		Username string `json:"username"`
		Token    string `json:"token"`
	}

	type credAWSRequest struct {
		Service    string        `json:"service,omitempty"`
		Credential CredentialAWS `json:"credential"`
	}

	type credGithubRequest struct {
		Service    string           `json:"service,omitempty"`
		Credential CredentialGithub `json:"credential"`
	}

	var b []byte
	var err error
	if in.Provider == "github" {
		req := credGithubRequest{
			Service: in.Provider,
			Credential: CredentialGithub{
				Username: in.ProviderUsername,
				Token:    in.ProviderSecret,
			},
		}
		b, err = json.Marshal(&req)
		if err != nil {
			return fmt.Errorf("error encoding credential: %w", err)
		}

	} else {
		req := credAWSRequest{
			Service: in.Provider,
			Credential: CredentialAWS{
				AccessKeyId:     in.ProviderUsername,
				SecretAccessKey: in.ProviderSecret,
			},
		}
		b, err = json.Marshal(&req)
		if err != nil {
			return fmt.Errorf("error encoding credential: %w", err)
		}
	}

	credURL := fmt.Sprintf("%s/credentials", host)
	req, err := http.NewRequest(http.MethodPost, credURL, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-org", "zup")
	req.Header.Set("x-authorization", token)
	req.Header.Set("x-ctx", ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error setting credential: %w", err)
	}

	defer resp.Body.Close()

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	switch resp.StatusCode {
	case 201:
		prompt.Success("done")
		return nil
	case 401:
		return errors.New("set credential failed! Verify your credentials")
	case 403:
		return errors.New("set credential failed! You have not access for the resource")
	default:
		return errors.New(resp.Status + string(b))
	}

}
