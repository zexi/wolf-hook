package cmd

import (
	"github.com/spf13/cobra"
)

var textCmd = &cobra.Command{
	Use:   "text",
	Short: "发送文本输入事件",
	RunE: func(cmd *cobra.Command, args []string) error {
		text, _ := cmd.Flags().GetString("text")
		return getInput().SendUTF8Text(sessionID, text)
	},
}

func init() {
	textCmd.Flags().String("text", "", "要输入的文本")
	textCmd.MarkFlagRequired("text")
}
