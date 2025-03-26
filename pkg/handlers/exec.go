package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"yunion.io/x/log"
)

type execController struct{}

func NewExecController() http.Handler {
	return new(execController)
}

type ExecParams struct {
	Cmd  string   `json:"cmd"`  // 要执行的命令
	Args []string `json:"args"` // 命令参数
	User string   `json:"user"` // 执行命令的用户
}

type ExecResponse struct {
	Output string `json:"output"`          // 命令输出
	Error  string `json:"error,omitempty"` // 错误信息，如果有的话
}

func (e *execController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Infof("Exec request received: %s", r.URL.Path)

	// 解析请求参数
	params := new(ExecParams)
	if err := json.NewDecoder(r.Body).Decode(params); err != nil {
		log.Errorf("解析请求参数失败: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 执行命令
	output, err := e.runCommand(params.Cmd, params.Args, params.User)

	// 准备响应
	response := ExecResponse{
		Output: string(output),
	}

	if err != nil {
		log.Errorf("执行命令失败: %v, output: %s", err, output)
		response.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		log.Infof("命令执行成功: %s", output)
		w.WriteHeader(http.StatusOK)
	}

	// 返回响应
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Errorf("编码响应失败: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (e *execController) runCommand(command string, args []string, user string) (string, error) {
	// 设置命令的环境变量
	env := os.Environ()
	env = append(env, fmt.Sprintf("USER=%s", user))

	// 执行命令
	cmd := exec.Command(command, args...)
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	return string(output), err
}
