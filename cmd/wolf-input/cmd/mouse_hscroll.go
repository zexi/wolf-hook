package cmd

import (
	"github.com/spf13/cobra"
)

var mouseHScrollCmd = &cobra.Command{
	Use:   "mouse-hscroll",
	Short: "发送鼠标水平滚轮事件",
	RunE: func(cmd *cobra.Command, args []string) error {
		scrollAmount, _ := cmd.Flags().GetInt("scroll")
		return getInput().SendMouseHScroll(sessionID, int16(scrollAmount))
	},
}

func init() {
	mouseHScrollCmd.Flags().Int("scroll", 0, "水平滚轮滚动量")
	mouseHScrollCmd.MarkFlagRequired("scroll")
}
