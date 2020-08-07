package formula

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"

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

func (in Inputs) Run() {
	loginResp, err := in.login()
	if err != nil {
		prompt.Error(err.Error())
		os.Exit(1)
	}

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

	formList := make([]string, len(formulasResp.Formulas))
	for i, f := range formulasResp.Formulas {
		formList[i] = f.Command
	}
	formName, err := list.List("Select a formula", formList)
	if err != nil {
		prompt.Error(err.Error())
		os.Exit(1)
	}

	var form formula
	for _, f := range formulasResp.Formulas {
		if f.Command == formName {
			form = f
			break
		}
	}

	cmdID, err := in.sendCommand(form, loginResp.Token, ctx)
	if err != nil {
		prompt.Error(err.Error())
		os.Exit(1)
	}

	prompt.Info(cmdID)

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
		prompt.Info("done")
		return formulasResp, err
	} else {
		return formulasResp, errors.New("error obtaining formulas")
	}
}

func (in Inputs) sendCommand(form formula, token, ctx string) (string, error) {
	list := prompt.NewSurveyList()
	text := prompt.NewSurveyText()
	boolean := prompt.NewSurveyBool()
	password := prompt.NewSurveyPassword()

	inputs := make([]input, len(form.Inputs))
	for i, in := range form.Inputs {
		var err error
		var inputVal string
		var valBool bool
		switch iType := in.Type; iType {
		case "text":
			if in.Items != nil {
				inputVal, err = list.List(in.Label, in.Items)
			} else {
				validate := in.Default == ""
				inputVal, err = text.Text(in.Label, validate)
				if inputVal == "" {
					inputVal = in.Default
				}
			}
		case "bool":
			valBool, err = boolean.Bool(in.Label, in.Items)
			inputVal = strconv.FormatBool(valBool)
		case "password":
			inputVal, err = password.Password(in.Label)
		}

		if err != nil {
			return "", fmt.Errorf("error reading inputs: %w", err)
		}

		inputs[i] = input{
			Name:  in.Name,
			Type:  in.Type,
			Value: inputVal,
		}
	}

	prompt.Info("Sending command...")

	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("error generatind UUID: %w", err)
	}
	cmdReq := commandRequest{
		ID:      id.String(),
		Command: form.Command,
		Inputs:  inputs,
	}

	b, err := json.Marshal(&cmdReq)
	if err != nil {
		return "", fmt.Errorf("error encoding command: %w", err)
	}

	cmdURL := fmt.Sprintf("%s/commands", host)
	req, err := http.NewRequest(http.MethodPost, cmdURL, bytes.NewBuffer(b))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-org", "zup")
	req.Header.Set("x-ctx", ctx)
	req.Header.Set("x-authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending command: %w", err)
	}

	defer resp.Body.Close()

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	switch resp.StatusCode {
	case 201:
		prompt.Info("done")
		return cmdReq.ID, err
	case 401:
		return "", errors.New("command failed! Verify your credentials")
	default:
		return "", errors.New("command failed")
	}
}
