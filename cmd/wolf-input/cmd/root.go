package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	sessionID  string
	socketPath string
	rootCmd    = &cobra.Command{
		Use:   "wolf-input",
		Short: "Wolf Input 是一个用于发送输入事件的命令行工具",
		Long:  `Wolf Input 是一个用于发送输入事件的命令行工具，支持鼠标移动、点击、键盘输入等功能。`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// 在命令执行前初始化输入实例
			initInput()
			initSessionID()
		},
	}
)

func init() {
	// 添加全局可选参数 session-id
	rootCmd.PersistentFlags().StringVar(&sessionID, "session", "", "会话ID (可选，如果不指定将自动获取第一个可用会话)")

	// 添加 socket 路径参数
	rootCmd.PersistentFlags().StringVar(&socketPath, "socket", "/tmp/wolf.sock", "Unix socket 路径")

	// 添加各个子命令
	rootCmd.AddCommand(mouseMoveRelCmd)
	rootCmd.AddCommand(mouseMoveAbsCmd)
	rootCmd.AddCommand(mouseButtonCmd)
	rootCmd.AddCommand(keyboardCmd)
	rootCmd.AddCommand(mouseScrollCmd)
	rootCmd.AddCommand(mouseHScrollCmd)
	rootCmd.AddCommand(textCmd)
}

func initSessionID() {
	if sessionID == "" {
		var err error
		sessionID, err = getInput().GetFirstSessionID()
		if err != nil {
			log.Fatalf("自动获取会话ID失败: %v", err)
		}
	}
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
