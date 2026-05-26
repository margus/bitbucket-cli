package actions

import (
	"fmt"
	"strconv"

	"github.com/margus/bb-cli/internal/api"
	"github.com/margus/bb-cli/internal/util"
)

type Env struct{}

func (Env) Default() string { return "environments" }
func (Env) Commands() map[string]string {
	return map[string]string{
		"environments":    "list, l",
		"variables":       "variables, v",
		"createVariable":  "create-variable, c",
		"updateVariable":  "update-variable, u",
	}
}
func (Env) RequireGit() bool { return true }

func (e Env) Dispatch(method string, args []string) error {
	switch method {
	case "environments":
		return e.environments()
	case "variables":
		if len(args) < 1 {
			return fmt.Errorf("variables requires <env-uuid>")
		}
		return e.variables(args[0])
	case "createVariable":
		if len(args) < 3 {
			return fmt.Errorf("create-variable requires <env-uuid> <key> <value> [secured]")
		}
		secured := false
		if len(args) >= 4 {
			secured = parseBool(args[3])
		}
		return e.createVariable(args[0], args[1], args[2], secured)
	case "updateVariable":
		if len(args) < 4 {
			return fmt.Errorf("update-variable requires <env-uuid> <var-uuid> <key> <value> [secured]")
		}
		secured := false
		if len(args) >= 5 {
			secured = parseBool(args[4])
		}
		return e.updateVariable(args[0], args[1], args[2], args[3], secured)
	default:
		return fmt.Errorf("unknown method: %s", method)
	}
}

type envListResp struct {
	Values []struct {
		UUID string `json:"uuid"`
		Name string `json:"name"`
	} `json:"values"`
}

func (Env) environments() error {
	c := api.New()
	var r envListResp
	if err := c.JSON("GET", "/environments", nil, true, &r); err != nil {
		return err
	}
	for _, env := range r.Values {
		util.O(map[string]string{"uuid": env.UUID, "name": env.Name}, "yellow")
	}
	return nil
}

type envVar struct {
	UUID    string `json:"uuid"`
	Key     string `json:"key"`
	Value   string `json:"value,omitempty"`
	Secured bool   `json:"secured"`
}

func (Env) variables(envUUID string) error {
	c := api.New()
	var r struct {
		Values []envVar `json:"values"`
	}
	if err := c.JSON("GET", "/deployments_config/environments/"+envUUID+"/variables", nil, true, &r); err != nil {
		return err
	}
	for _, v := range r.Values {
		util.O(map[string]string{
			"uuid":    v.UUID,
			"key":     v.Key,
			"value":   v.Value,
			"secured": yesNo(v.Secured),
		}, "yellow")
	}
	return nil
}

func (Env) createVariable(envUUID, key, value string, secured bool) error {
	c := api.New()
	var resp envVar
	payload := map[string]any{"key": key, "value": value, "secured": secured}
	if err := c.JSON("POST", "/deployments_config/environments/"+envUUID+"/variables", payload, true, &resp); err != nil {
		return err
	}
	printVariableResponse(resp)
	return nil
}

func (Env) updateVariable(envUUID, varUUID, key, value string, secured bool) error {
	c := api.New()
	var resp envVar
	payload := map[string]any{"key": key, "value": value, "secured": secured}
	if err := c.JSON("PUT", "/deployments_config/environments/"+envUUID+"/variables/"+varUUID, payload, true, &resp); err != nil {
		return err
	}
	printVariableResponse(resp)
	return nil
}

func printVariableResponse(v envVar) {
	util.O(map[string]string{
		"uuid":    v.UUID,
		"key":     v.Key,
		"value":   v.Value,
		"secured": yesNo(v.Secured),
	}, "yellow")
}

func yesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func parseBool(s string) bool {
	v, _ := strconv.ParseBool(s)
	return v
}
