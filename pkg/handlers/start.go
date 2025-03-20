package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"yunion.io/x/log"
	"yunion.io/x/pkg/errors"
)

const (
	GOW_STARTUP_APP_SH = "/entrypoint.sh"
	HOOK_ENV_FILE      = "/opt/bin/hook-env.sh"
)

type startController struct{}

func NewStartController() http.Handler {
	return new(startController)
}

type StartParams struct {
	Envs map[string]string `json:"envs"`
}

func (s startController) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	log.Infof("=======start request recevied")
	params := new(StartParams)
	if err := json.NewDecoder(request.Body).Decode(params); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	go func() {
		if err := s.launchApp(params); err != nil {
			log.Errorf("launch app failed: %v", err)
		}
	}()
	log.Printf("======get start params: %+v", params)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("OK"))
	return
}

func (s startController) launchApp(params *StartParams) error {
	cmd := exec.Command(GOW_STARTUP_APP_SH)
	cmd.Env = os.Environ()
	for k, v := range params.Envs {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	var envContent string
	for _, e := range cmd.Env {
		pair := strings.SplitN(e, "=", 2)
		envContent += fmt.Sprintf("export %s='%s'\n", pair[0], pair[1])
	}
	if err := os.WriteFile(HOOK_ENV_FILE, []byte(envContent), 0644); err != nil {
		log.Errorf("write env file failed: %v", err)
	} else {
		log.Infof("env content: \n%s", envContent)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Infof("launch app as subprocess: %v with env: %v", cmd.Args, cmd.Environ())
	if err := cmd.Run(); err != nil {
		log.Errorf("start app failed: %v", err)
		return errors.Wrap(err, "start app failed")
	}
	return nil
}

/*func (s startController) launchExecApp(params *StartParams) error {
	cmd := exec.Command(GOW_STARTUP_APP_SH)
	cmd.Env = os.Environ()
	for k, v := range params.Envs {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	var envContent string
	for _, e := range cmd.Env {
		pair := strings.SplitN(e, "=", 2)
		envContent += fmt.Sprintf("export %s='%s'\n", pair[0], pair[1])
	}
	if err := os.WriteFile(HOOK_ENV_FILE, []byte(envContent), 0644); err != nil {
		log.Errorf("write env file failed: %v", err)
	} else {
		log.Infof("env content: \n%s", envContent)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Infof("launch app: %v with env: %v", cmd.Args, cmd.Environ())
	if err := syscall.Exec(GOW_STARTUP_APP_SH, []string{GOW_STARTUP_APP_SH}, cmd.Environ()); err != nil {
		log.Errorf("start exec app failed: %v", err)
	}
	return nil
}*/
