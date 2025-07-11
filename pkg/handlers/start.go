package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"yunion.io/x/log"
	"yunion.io/x/pkg/errors"
)

const (
	GOW_STARTUP_APP_SH = "/entrypoint.sh"
	HOOK_ENV_FILE      = "/opt/bin/hook-env.sh"
)

type startController struct {
	noExitWhenAppLaunch bool
}

func NewStartController(noExitWhenAppLaunch bool) http.Handler {
	return &startController{
		noExitWhenAppLaunch: noExitWhenAppLaunch,
	}
}

type StartParams struct {
	Envs map[string]string `json:"envs"`
}

func (s startController) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	params := new(StartParams)
	if err := json.NewDecoder(request.Body).Decode(params); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	go func() {
		SetState(STATE_RUNNING)
		if err := s.launchApp(params); err != nil {
			log.Errorf("launch app failed: %v", err)
			SetState(STATE_ERROR)
		}
	}()
	log.Printf("======get start params: %+v", params)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("OK"))
	return
}

// setupUdevControl 创建并设置 udev control 文件和目录
func (s startController) setupUdevControl() error {
	udevDir := "/run/udev"
	udevDataDir := "/run/udev/data"
	controlFile := "/run/udev/control"

	// 确保 udev 目录存在
	if err := os.MkdirAll(udevDir, 0777); err != nil {
		log.Errorf("创建目录 %s 失败: %v", udevDir, err)
		return errors.Wrap(err, "创建 udev 目录失败")
	}

	// 创建 data 目录
	if err := os.MkdirAll(udevDataDir, 0777); err != nil {
		log.Errorf("创建目录 %s 失败: %v", udevDataDir, err)
		return errors.Wrap(err, "创建 udev data 目录失败")
	}

	// 创建 control 文件
	if _, err := os.Create(controlFile); err != nil {
		log.Errorf("创建文件 %s 失败: %v", controlFile, err)
		return errors.Wrap(err, "创建 control 文件失败")
	}

	// 设置文件权限为 777
	if err := os.Chmod(controlFile, 0777); err != nil {
		log.Errorf("设置文件 %s 权限失败: %v", controlFile, err)
		return errors.Wrap(err, "设置 control 文件权限失败")
	}
	return nil
}

func (s startController) setupSteamDir() error {
	// 获取 HOME 环境变量
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return errors.Errorf("HOME 环境变量未设置")
	}

	// Steam 目录路径
	steamDir := fmt.Sprintf("%s/.steam", homeDir)
	debianInstallDir := fmt.Sprintf("%s/.steam/debian-installation", homeDir)

	// 设置 .steam 目录权限
	if _, err := os.Stat(steamDir); err == nil {
		// 目录存在，设置权限
		if err := os.Chown(steamDir, 1000, 1000); err != nil {
			log.Errorf("设置目录 %s 权限失败: %v", steamDir, err)
			return errors.Wrapf(err, "设置目录 %s 权限失败", steamDir)
		}
		log.Infof("已设置目录 %s 权限为 1000:1000", steamDir)
	} else if os.IsNotExist(err) {
		log.Infof("目录 %s 不存在，跳过权限设置", steamDir)
	} else {
		log.Errorf("检查目录 %s 时发生错误: %v", steamDir, err)
		return errors.Wrapf(err, "检查目录 %s 失败", steamDir)
	}

	// 设置 .steam/debian-installation 目录权限
	if _, err := os.Stat(debianInstallDir); err == nil {
		// 目录存在，设置权限
		if err := os.Chown(debianInstallDir, 1000, 1000); err != nil {
			log.Errorf("设置目录 %s 权限失败: %v", debianInstallDir, err)
			return errors.Wrapf(err, "设置目录 %s 权限失败", debianInstallDir)
		}
		log.Infof("已设置目录 %s 权限为 1000:1000", debianInstallDir)
	} else if os.IsNotExist(err) {
		log.Infof("目录 %s 不存在，跳过权限设置", debianInstallDir)
	} else {
		log.Errorf("检查目录 %s 时发生错误: %v", debianInstallDir, err)
		return errors.Wrapf(err, "检查目录 %s 失败", debianInstallDir)
	}

	return nil
}

func (s startController) launchApp(params *StartParams) error {
	// 设置 udev control 文件
	if err := s.setupUdevControl(); err != nil {
		return errors.Wrap(err, "设置 udev control 文件失败")
	}

	// 设置 Steam 目录权限
	if err := s.setupSteamDir(); err != nil {
		return errors.Wrapf(err, "设置 Steam 目录权限失败")
	}

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

	// 启动 goroutine 检查 sway 进程（如果未设置 no-exit-when-app-launch 标志）
	if !s.noExitWhenAppLaunch {
		go func() {
			for {
				// 检查是否有 sway 进程
				if !isSwayRunning() {
					log.Infof("未检测到 sway 进程，5 秒后退出程序")
					time.Sleep(2 * time.Second)
					log.Infof("退出程序")
					os.Exit(134)
				}
				// 每隔一段时间检查一次
				time.Sleep(3 * time.Second)
			}
		}()
	} else {
		log.Infof("已设置 no-exit-when-app-launch 标志，跳过 sway 进程检测")
	}

	return nil
}

// isSwayRunning 检查系统中是否有 sway 进程
func isSwayRunning() bool {
	cmd := exec.Command("pgrep", "sway")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
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
