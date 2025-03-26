package handlers

import (
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"syscall"

	"yunion.io/x/log"
	"yunion.io/x/pkg/errors"
)

type stopController struct{}

func NewStopController() http.Handler {
	return new(stopController)
}

// killOtherProcesses 终止除当前进程外的所有进程
func (s *stopController) killOtherProcesses() error {
	currentPid := os.Getpid()

	// 读取 /proc 目录获取所有进程
	dirs, err := ioutil.ReadDir("/proc")
	if err != nil {
		log.Errorf("读取 /proc 目录失败: %v", err)
		return errors.Wrap(err, "读取 /proc 目录失败")
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		// 尝试将目录名转换为 pid
		pid, err := strconv.Atoi(dir.Name())
		if err != nil {
			continue // 不是进程目录
		}

		// 跳过当前进程
		if pid == currentPid {
			continue
		}

		// 检查进程是否存在并发送 SIGTERM 信号
		process, err := os.FindProcess(pid)
		if err != nil {
			continue
		}

		// 发送 SIGTERM 信号
		if err := process.Signal(syscall.SIGTERM); err != nil {
			log.Warningf("无法发送 SIGTERM 到进程 %d: %v", pid, err)
			// 如果 SIGTERM 失败，尝试 SIGKILL
			if err := process.Signal(syscall.SIGKILL); err != nil {
				log.Errorf("无法发送 SIGKILL 到进程 %d: %v", pid, err)
			}
		}
	}
	return nil
}

func (s *stopController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Infof("Stop request: %s", r.URL.Path)
	SetState(STATE_STOPPED)
	w.WriteHeader(http.StatusAccepted)
	go func() {
		log.Infof("正在终止其他进程...")
		if err := s.killOtherProcesses(); err != nil {
			log.Errorf("终止其他进程时发生错误: %v", err)
		}
	}()
}
