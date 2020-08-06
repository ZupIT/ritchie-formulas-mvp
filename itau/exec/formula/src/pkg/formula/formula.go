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
	host = "http://0.0.0.0:8882"
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

type formulasResponse struct {
	Contexts contexts `json:"contexts"`
	Formulas formulas `json:"formulas"`
}

type context struct {
	Name string `json:"name"`
}

type contexts []context

type formula struct {
	Command string `json:"command"`
	Inputs  inputs `json:"inputs"`
}

type formulas []formula

type input struct {
	Name    string `json:"name"`
	Label   string `json:"label"`
	Type    string `json:"type"`
	Items   items  `json:"items"`
	Default string `json:"default"`
}

type items []string

type inputs []input

func (in Inputs) Run() {
	loginResp, err := in.login()
	if err != nil {
		prompt.Error(err.Error())
		os.Exit(1)
	}

	formulasResp, err := in.formulas(loginResp)
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

	prompt.Info(fmt.Sprintf("%s", ctx))

	formList := make([]string, len(formulasResp.Formulas))
	for i, f := range formulasResp.Formulas {
		formList[i] = f.Command
	}
	form, err := list.List("Select a formula", formList)
	if err != nil {
		prompt.Error(err.Error())
		os.Exit(1)
	}

	prompt.Info(fmt.Sprintf("%s", form))

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
		prompt.Info("done")
		return loginResp, err
	case 401:
		return loginResp, errors.New("login failed! Verify your credentials")
	default:
		return loginResp, errors.New("login failed")
	}

}

func (in Inputs) formulas(loginResp loginResponse) (formulasResponse, error) {
	prompt.Info("Obtaining formulas...")

	formulasResp := formulasResponse{}

	formulasURL := fmt.Sprintf("%s/formulas", host)
	req, err := http.NewRequest(http.MethodGet, formulasURL, nil)
	if err != nil {
		return formulasResp, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("x-org", "zup")
	req.Header.Set("x-authorization", loginResp.Token)

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
		prompt.Info("done")
		return formulasResp, err
	} else {
		return formulasResp, errors.New("error obtaining formulas")
	}
}
