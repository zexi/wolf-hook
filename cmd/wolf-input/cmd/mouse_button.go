package cmd

import (
	"github.com/spf13/cobra"
)

var mouseButtonCmd = &cobra.Command{
	Use:   "mouse-button",
	Short: "发送鼠标按键事件",
	RunE: func(cmd *cobra.Command, args []string) error {
		button, _ := cmd.Flags().GetInt("button")
		isPress, _ := cmd.Flags().GetBool("press")
		return getInput().SendMouseButton(sessionID, uint8(button), isPress)
	},
}

func init() {
	mouseButtonCmd.Flags().Int("button", 1, "鼠标按键 (1=左键, 2=右键, 3=中键)")
	mouseButtonCmd.Flags().Bool("press", true, "按键是否按下")
	mouseButtonCmd.MarkFlagRequired("button")
}
