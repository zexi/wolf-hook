package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/zexi/wolf-hook/pkg/moonlight/client"
)

// 全局配置
var (
	hostIP     string
	httpPort   int
	clientID   string
	autoPair   bool
	noAutoPair bool
)

// 创建并配置客户端
func createClient() *client.MoonlightClient {
	log.Printf("使用配置: 服务器IP=%s, HTTP端口=%d, 客户端ID=%s", hostIP, httpPort, clientID)
	return client.NewMoonlightClient(hostIP, httpPort, clientID)
}

// 确保客户端已配对
func ensurePaired(moonlightClient *client.MoonlightClient) error {
	if noAutoPair {
		log.Println("跳过自动配对，请先手动执行 'pair' 命令")
		return fmt.Errorf("需要先配对，请运行: wolf-session pair")
	}

	// 这里可以添加检查配对状态的逻辑
	// 目前简单起见，总是尝试配对
	log.Println("正在与服务器配对...")
	if err := moonlightClient.PairAndConnect(); err != nil {
		return fmt.Errorf("配对失败: %v", err)
	}
	log.Println("配对成功！")
	return nil
}

// 配对命令
func cmdPair(cmd *cobra.Command, args []string) {
	moonlightClient := createClient()

	// 执行配对
	if err := moonlightClient.PairAndConnect(); err != nil {
		log.Fatalf("配对失败: %v", err)
	}

	log.Println("配对成功！")
}

// 获取服务器信息命令
func cmdServerInfo(cmd *cobra.Command, args []string) {
	moonlightClient := createClient()

	// 确保已配对
	if err := ensurePaired(moonlightClient); err != nil {
		log.Fatalf("配对失败: %v", err)
	}

	// 获取服务器信息
	serverInfo, err := moonlightClient.GetServerInfo()
	if err != nil {
		log.Fatalf("获取服务器信息失败: %v", err)
	}

	log.Printf("服务器信息: 主机名=%s, 版本=%s, 状态=%s", serverInfo.Hostname, serverInfo.AppVersion, serverInfo.State)
	log.Printf("支持的显示模式数量: %d", len(serverInfo.SupportedDisplayMode.DisplayMode))
	for i, mode := range serverInfo.SupportedDisplayMode.DisplayMode {
		log.Printf("  模式%d: %sx%s@%sHz", i+1, mode.Width, mode.Height, mode.RefreshRate)
	}
}

// 获取应用程序列表命令
func cmdAppList(cmd *cobra.Command, args []string) {
	moonlightClient := createClient()

	// 确保已配对
	if err := ensurePaired(moonlightClient); err != nil {
		log.Fatalf("配对失败: %v", err)
	}

	// 获取应用程序列表
	appList, err := moonlightClient.GetAppList()
	if err != nil {
		log.Fatalf("获取应用程序列表失败: %v", err)
	}

	log.Printf("应用程序列表: 共%d个应用", len(appList.Apps))
	for _, app := range appList.Apps {
		log.Printf("  应用: ID=%s, 标题=%s, HDR支持=%s", app.ID, app.AppTitle, app.IsHdrSupported)
	}
}

// 启动应用程序命令
func cmdLaunch(cmd *cobra.Command, args []string) {
	// 从命令标志获取参数
	appID, _ := cmd.Flags().GetString("app-id")
	width, _ := cmd.Flags().GetInt("width")
	height, _ := cmd.Flags().GetInt("height")
	fps, _ := cmd.Flags().GetInt("fps")

	moonlightClient := createClient()

	// 确保已配对
	if err := ensurePaired(moonlightClient); err != nil {
		log.Fatalf("配对失败: %v", err)
	}

	// 启动应用程序
	launchResult, err := moonlightClient.LaunchApp(appID, width, height, fps)
	if err != nil {
		log.Fatalf("启动应用程序失败: %v", err)
	}

	log.Printf("启动结果: 会话ID=%s", launchResult.SessionURL0)
}

func main() {
	// 创建根命令
	rootCmd := &cobra.Command{
		Use:   "wolf-session",
		Short: "Wolf Session - Moonlight 客户端",
		Long: `Wolf Session 是一个 Moonlight 客户端工具，用于与 Moonlight 服务器进行交互。

支持的功能包括：
- 与服务器配对
- 获取服务器信息
- 获取应用程序列表
- 启动应用程序

使用说明：
1. 默认情况下，所有命令都会自动进行配对
2. 使用 --no-auto-pair 标志可以禁用自动配对，需要先手动执行 'pair' 命令
3. 配对成功后，可以正常使用其他功能`,
	}

	// 添加全局标志
	rootCmd.PersistentFlags().StringVar(&hostIP, "host", "220.196.214.104", "服务器IP地址")
	rootCmd.PersistentFlags().IntVar(&httpPort, "port", 20008, "HTTP端口")
	rootCmd.PersistentFlags().StringVar(&clientID, "client-id", "go_client_001", "客户端ID")
	rootCmd.PersistentFlags().BoolVar(&noAutoPair, "no-auto-pair", false, "禁用自动配对")

	// 配对命令
	pairCmd := &cobra.Command{
		Use:   "pair",
		Short: "与服务器配对",
		Long:  `与指定的 Moonlight 服务器进行配对操作`,
		Run:   cmdPair,
	}
	rootCmd.AddCommand(pairCmd)

	// 服务器信息命令
	serverInfoCmd := &cobra.Command{
		Use:   "server-info",
		Short: "获取服务器信息",
		Long:  `获取 Moonlight 服务器的详细信息，包括主机名、版本、支持的显示模式等`,
		Run:   cmdServerInfo,
	}
	rootCmd.AddCommand(serverInfoCmd)

	// 应用程序列表命令
	appListCmd := &cobra.Command{
		Use:   "app-list",
		Short: "获取应用程序列表",
		Long:  `获取服务器上可用的应用程序列表`,
		Run:   cmdAppList,
	}
	rootCmd.AddCommand(appListCmd)

	// 启动应用程序命令
	launchCmd := &cobra.Command{
		Use:   "launch",
		Short: "启动应用程序",
		Long:  `启动指定的应用程序，可以配置显示分辨率和帧率`,
		Run:   cmdLaunch,
	}
	launchCmd.Flags().String("app-id", "1", "应用程序ID")
	launchCmd.Flags().Int("width", 1920, "显示宽度")
	launchCmd.Flags().Int("height", 1080, "显示高度")
	launchCmd.Flags().Int("fps", 60, "帧率")
	rootCmd.AddCommand(launchCmd)

	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
