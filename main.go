package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"yunion.io/x/pkg/errors"

	"github.com/zexi/wolf-hook/pkg/handlers"
	"github.com/zexi/wolf-hook/pkg/moonlight/client"
	"github.com/zexi/wolf-hook/pkg/util/procutils"

	"yunion.io/x/log"
)

var (
	addr                string
	port                int
	ulimitNofileHard    int
	ulimitNofileSoft    int
	autoStart           bool
	noExitWhenAppLaunch bool
)

func init() {
	flag.StringVar(&addr, "addr", "127.0.0.1", "HTTP server listen address")
	flag.IntVar(&port, "port", 8080, "HTTP server listen port")
	flag.IntVar(&ulimitNofileHard, "ulimit-nofile-hard", 10240, "ulimit nofile hard")
	flag.IntVar(&ulimitNofileSoft, "ulimit-nofile-soft", 10240, "ulimit nofile soft")
	flag.BoolVar(&autoStart, "auto-start", false, "auto start moonlight client after HTTP server starts")
	flag.BoolVar(&noExitWhenAppLaunch, "no-exit-when-app-launch", false, "do not exit when app launch (skip sway process monitoring)")
	flag.Parse()
}

// getConfigFromEnv 从环境变量获取 Moonlight 配置
func getConfigFromEnv() (string, int, string) {
	// 从环境变量读取服务器IP，默认为 220.196.214.104
	hostIP := os.Getenv("CLOUDPODS_HOST_EIP")
	if hostIP == "" {
		hostIP = os.Getenv("CLOUDPODS_HOST_ACCESS_IP")
		if hostIP == "" {
			hostIP = "220.196.214.104"
		}
	}

	// 从环境变量读取HTTP端口，默认为 20008
	httpPortStr := os.Getenv("WOLF_HTTP_PORT")
	httpPort := 20008
	if httpPortStr != "" {
		if port, err := strconv.Atoi(httpPortStr); err == nil {
			httpPort = port
		}
	}

	// 从环境变量读取客户端ID，默认为 go_client_001
	clientID := os.Getenv("WOLF_CLIENT_ID")
	if clientID == "" {
		clientID = "go_client_001"
	}

	return hostIP, httpPort, clientID
}

// checkServerConnectivity 检查服务器IP和HTTP端口是否可达
func checkServerConnectivity(hostIP string, httpPort int) error {
	log.Infof("检查服务器连通性: %s:%d", hostIP, httpPort)

	// 检查TCP连接
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", hostIP, httpPort), 5*time.Second)
	if err != nil {
		return errors.Wrapf(err, "无法连接到服务器 %s:%d", hostIP, httpPort)
	}
	defer conn.Close()

	log.Infof("服务器连通性检查通过: %s:%d", hostIP, httpPort)
	return nil
}

// startMoonlightClient 启动 Moonlight 客户端
func startMoonlightClient() error {
	// 从环境变量获取配置
	hostIP, httpPort, clientID := getConfigFromEnv()
	log.Infof("Moonlight 配置: 服务器IP=%s, HTTP端口=%d, 客户端ID=%s", hostIP, httpPort, clientID)

	// 检查服务器连通性
	if err := checkServerConnectivity(hostIP, httpPort); err != nil {
		return errors.Wrap(err, "服务器连通性检查失败")
	}
	log.Infof("等待5秒，确保服务器 pulseaudio 完全启动")
	time.Sleep(5 * time.Second)

	// 创建客户端
	log.Infof("============= 启动 Moonlight 客户端 ==========")
	moonlightClient := client.NewMoonlightClient(hostIP, httpPort, clientID)

	// 执行配对
	log.Infof("开始配对流程...")
	if err := moonlightClient.PairAndConnect(); err != nil {
		return errors.Wrap(err, "配对失败")
	}
	log.Infof("配对成功！")

	// 获取服务器信息
	serverInfo, err := moonlightClient.GetServerInfo()
	if err != nil {
		return errors.Wrap(err, "获取服务器信息失败")
	} else {
		log.Infof("服务器信息: 主机名=%s, 版本=%s, 状态=%s", serverInfo.Hostname, serverInfo.AppVersion, serverInfo.State)
		log.Infof("支持的显示模式数量: %d", len(serverInfo.SupportedDisplayMode.DisplayMode))
		for i, mode := range serverInfo.SupportedDisplayMode.DisplayMode {
			log.Infof("  模式%d: %sx%s@%sHz", i+1, mode.Width, mode.Height, mode.RefreshRate)
		}
	}

	// 获取应用程序列表
	appList, err := moonlightClient.GetAppList()
	if err != nil {
		return errors.Wrap(err, "获取应用程序列表失败")
	}

	log.Infof("应用程序列表: 共%d个应用", len(appList.Apps))
	for _, app := range appList.Apps {
		log.Infof("  应用: ID=%s, 标题=%s, HDR支持=%s", app.ID, app.AppTitle, app.IsHdrSupported)
	}

	// 启动第一个应用程序（如果有的话）
	if len(appList.Apps) > 0 {
		firstApp := appList.Apps[0]
		log.Infof("自动启动第一个应用: %s (ID: %s)", firstApp.AppTitle, firstApp.ID)

		launchResult, err := moonlightClient.LaunchApp(firstApp.ID, 1920, 1080, 60)
		if err != nil {
			return errors.Wrap(err, "启动应用程序失败")
		} else {
			log.Infof("启动成功！会话ID: %s", launchResult.SessionURL0)
		}
	} else {
		return errors.Errorf("没有找到可用的应用程序")
	}

	return nil
}

func setupRlimits(hard, soft uint64) error {
	l := &syscall.Rlimit{
		Max: hard,
		Cur: soft,
	}
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, l); err != nil {
		return errors.Wrap(err, "syscall.Setrlimit")
	}
	log.Infof("set ulimit nofile hard %d soft %d", l.Cur, l.Max)
	return nil
}

func main() {
	log.Infof("============= WOLF HOOK ==========")
	if err := setupRlimits(uint64(ulimitNofileHard), uint64(ulimitNofileSoft)); err != nil {
		log.Fatalf("setup ulimit nofile hard: %s", err)
	}

	go procutils.WaitZombieLoop(context.Background())

	srv := &http.Server{
		Handler:      getHandler(),
		Addr:         fmt.Sprintf("%s:%d", addr, port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Infof("Listening on %s:%d", addr, port)

	// 如果启用了自动启动，在后台启动 Moonlight 客户端
	if autoStart {
		go func() {
			// 等待一小段时间确保 HTTP 服务完全启动
			for {
				time.Sleep(1 * time.Second)
				if err := startMoonlightClient(); err != nil {
					log.Errorf("start moonlight client: %v", err)
				} else {
					break
				}
			}
		}()
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("listen and serve: %v", err)
	}
}

func getHandler() http.Handler {
	r := mux.NewRouter()
	r.Handle("/hook/start", handlers.NewStartController(noExitWhenAppLaunch)).Methods("POST")
	r.Handle("/hook/stop", handlers.NewStopController()).Methods("POST")
	r.Handle("/hook/status", handlers.NewGetStatusController()).Methods("GET")
	r.Handle("/hook/exec", handlers.NewExecController()).Methods("POST")
	r.Handle("/hook/write-hwdb", handlers.NewWriteHwdbController()).Methods("POST")
	r.Handle("/steam/owned-games", handlers.NewSteamOwnedGamesController()).Methods("GET")
	return r
}
