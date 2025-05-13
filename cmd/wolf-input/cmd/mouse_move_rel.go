package cmd

import (
	"github.com/spf13/cobra"
)

var mouseMoveRelCmd = &cobra.Command{
	Use:   "mouse-move-rel",
	Short: "发送鼠标相对移动事件",
	RunE: func(cmd *cobra.Command, args []string) error {
		deltaX, _ := cmd.Flags().GetInt("delta-x")
		deltaY, _ := cmd.Flags().GetInt("delta-y")
		return getInput().SendMouseMoveRel(sessionID, int16(deltaX), int16(deltaY))
	},
}

func init() {
	mouseMoveRelCmd.Flags().Int("delta-x", 0, "X 轴偏移量")
	mouseMoveRelCmd.Flags().Int("delta-y", 0, "Y 轴偏移量")
	mouseMoveRelCmd.MarkFlagRequired("delta-x")
	mouseMoveRelCmd.MarkFlagRequired("delta-y")
}
