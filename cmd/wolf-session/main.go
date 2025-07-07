package main

import (
	"log"
	"os"
	"strconv"

	"github.com/zexi/wolf-hook/pkg/moonlight/client"
)

// getConfigFromEnv 从环境变量获取配置
func getConfigFromEnv() (string, int, string) {
	// 从环境变量读取服务器IP，默认为 220.196.214.104
	hostIP := os.Getenv("WOLF_HOST_IP")
	if hostIP == "" {
		hostIP = "220.196.214.104"
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

func main() {
	// 从环境变量获取配置
	hostIP, httpPort, clientID := getConfigFromEnv()

	log.Printf("使用配置: 服务器IP=%s, HTTP端口=%d, 客户端ID=%s", hostIP, httpPort, clientID)

	// 配置客户端
	moonlightClient := client.NewMoonlightClient(hostIP, httpPort, clientID)

	// 执行配对
	if err := moonlightClient.PairAndConnect(); err != nil {
		log.Fatalf("配对失败: %v", err)
	}

	log.Println("配对成功！")

	// 获取服务器信息
	serverInfo, err := moonlightClient.GetServerInfo()
	if err != nil {
		log.Printf("获取服务器信息失败: %v", err)
	} else {
		log.Printf("服务器信息: 主机名=%s, 版本=%s, 状态=%s", serverInfo.Hostname, serverInfo.AppVersion, serverInfo.State)
		log.Printf("支持的显示模式数量: %d", len(serverInfo.SupportedDisplayMode.DisplayMode))
		for i, mode := range serverInfo.SupportedDisplayMode.DisplayMode {
			log.Printf("  模式%d: %sx%s@%sHz", i+1, mode.Width, mode.Height, mode.RefreshRate)
		}
	}

	// 获取应用程序列表
	appList, err := moonlightClient.GetAppList()
	if err != nil {
		log.Printf("获取应用程序列表失败: %v", err)
	} else {
		log.Printf("应用程序列表: 共%d个应用", len(appList.Apps))
		for _, app := range appList.Apps {
			log.Printf("  应用: ID=%s, 标题=%s, HDR支持=%s", app.ID, app.AppTitle, app.IsHdrSupported)
		}
	}

	// 启动应用程序（示例）
	launchResult, err := moonlightClient.LaunchApp("1", 1920, 1080, 60)
	if err != nil {
		log.Printf("启动应用程序失败: %v", err)
	} else {
		log.Printf("启动结果: 会话ID=%s", launchResult.SessionURL0)
	}
}
