package cmd

import (
	"github.com/spf13/cobra"
)

var mouseScrollCmd = &cobra.Command{
	Use:   "mouse-scroll",
	Short: "发送鼠标滚轮事件",
	RunE: func(cmd *cobra.Command, args []string) error {
		scrollAmount, _ := cmd.Flags().GetInt("scroll")
		return getInput().SendMouseScroll(sessionID, int16(scrollAmount))
	},
}

func init() {
	mouseScrollCmd.Flags().Int("scroll", 0, "滚轮滚动量")
	mouseScrollCmd.MarkFlagRequired("scroll")
}
