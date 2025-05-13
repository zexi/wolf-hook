package cmd

import (
	"github.com/zexi/wolf-hook/pkg/moonlight"
)

var (
	// input 是共享的 MoonlightInput 实例
	input *moonlight.MoonlightInput
)

// initInput 初始化输入实例
func initInput() {
	if input == nil {
		input = moonlight.NewMoonlightInput(socketPath)
	}
}

// getInput 获取输入实例
func getInput() *moonlight.MoonlightInput {
	if input == nil {
		initInput()
	}
	return input
}
