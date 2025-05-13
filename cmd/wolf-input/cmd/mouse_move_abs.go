package cmd

import (
	"github.com/spf13/cobra"
)

var mouseMoveAbsCmd = &cobra.Command{
	Use:   "mouse-move-abs",
	Short: "发送鼠标绝对移动事件",
	RunE: func(cmd *cobra.Command, args []string) error {
		x, _ := cmd.Flags().GetInt("x")
		y, _ := cmd.Flags().GetInt("y")
		width, _ := cmd.Flags().GetInt("width")
		height, _ := cmd.Flags().GetInt("height")
		return getInput().SendMouseMoveAbs(sessionID, int16(x), int16(y), int16(width), int16(height))
	},
}

func init() {
	mouseMoveAbsCmd.Flags().Int("x", 0, "X 坐标")
	mouseMoveAbsCmd.Flags().Int("y", 0, "Y 坐标")
	mouseMoveAbsCmd.Flags().Int("width", 1920, "屏幕宽度")
	mouseMoveAbsCmd.Flags().Int("height", 1080, "屏幕高度")
	mouseMoveAbsCmd.MarkFlagRequired("x")
	mouseMoveAbsCmd.MarkFlagRequired("y")
}
