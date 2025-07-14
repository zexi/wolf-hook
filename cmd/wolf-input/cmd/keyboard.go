package cmd

import (
	"github.com/spf13/cobra"
)

var keyboardCmd = &cobra.Command{
	Use:   "keyboard",
	Short: "发送键盘按键事件",
	RunE: func(cmd *cobra.Command, args []string) error {
		keyCode, _ := cmd.Flags().GetInt("key-code")
		modifiers, _ := cmd.Flags().GetInt("modifiers")
		isPress, _ := cmd.Flags().GetBool("press")

		return getInput().SendKeyboardKey(sessionID, uint16(keyCode), uint8(modifiers), isPress)
	},
}

func init() {
	keyboardCmd.Flags().Int("key-code", 0, "键盘按键代码")
	keyboardCmd.Flags().Int("modifiers", 0, "键盘修饰键")
	keyboardCmd.Flags().Bool("press", true, "按键是否按下")
	keyboardCmd.MarkFlagRequired("key-code")
}
