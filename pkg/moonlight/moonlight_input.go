package moonlight

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"yunion.io/x/log"
)

// InputType 定义输入类型
type InputType uint32

const (
	MouseMoveRel       InputType = 0x00000007
	MouseMoveAbs       InputType = 0x00000005
	MouseButtonPress   InputType = 0x00000008
	MouseButtonRelease InputType = 0x00000009
	KeyPress           InputType = 0x00000003
	KeyRelease         InputType = 0x00000004
	MouseScroll        InputType = 0x0000000A
	MouseHScroll       InputType = 0x55000001
	Touch              InputType = 0x55000002
	Pen                InputType = 0x55000003
	ControllerMulti    InputType = 0x0000000C
	ControllerArrival  InputType = 0x55000004
	ControllerTouch    InputType = 0x55000005
	ControllerMotion   InputType = 0x55000006
	ControllerBattery  InputType = 0x55000007
	Haptics            InputType = 0x0000000D
	UTF8Text           InputType = 0x00000017
)

// TouchEventType 定义触摸事件类型
type TouchEventType uint8

const (
	TouchDown       TouchEventType = 0x07
	TouchUp         TouchEventType = 0x08
	TouchMove       TouchEventType = 0x09
	TouchHover      TouchEventType = 0x0A
	TouchHoverLeave TouchEventType = 0x0B
	TouchCancel     TouchEventType = 0x0C
	TouchCancelAll  TouchEventType = 0x0D
	TouchButtonOnly TouchEventType = 0x0E
)

// ControllerType 定义控制器类型
type ControllerType uint8

const (
	ControllerUnknown  ControllerType = 0x00
	ControllerXbox     ControllerType = 0x01
	ControllerPS       ControllerType = 0x02
	ControllerNintendo ControllerType = 0x03
	ControllerAuto     ControllerType = 0xFF
)

// ControllerCapabilities 定义控制器能力
type ControllerCapabilities uint16

const (
	CapAnalogTriggers ControllerCapabilities = 0x01
	CapRumble         ControllerCapabilities = 0x02
	CapTriggerRumble  ControllerCapabilities = 0x04
	CapTouchpad       ControllerCapabilities = 0x08
	CapAccelerometer  ControllerCapabilities = 0x10
	CapGyro           ControllerCapabilities = 0x20
	CapBattery        ControllerCapabilities = 0x40
	CapRGBLED         ControllerCapabilities = 0x80
)

// ControllerButton 定义控制器按钮
type ControllerButton uint32

const (
	ButtonDPadUp      ControllerButton = 0x0001
	ButtonDPadDown    ControllerButton = 0x0002
	ButtonDPadLeft    ControllerButton = 0x0004
	ButtonDPadRight   ControllerButton = 0x0008
	ButtonStart       ControllerButton = 0x0010
	ButtonBack        ControllerButton = 0x0020
	ButtonHome        ControllerButton = 0x0400
	ButtonLeftStick   ControllerButton = 0x0040
	ButtonRightStick  ControllerButton = 0x0080
	ButtonLeftButton  ControllerButton = 0x0100
	ButtonRightButton ControllerButton = 0x0200
	ButtonA           ControllerButton = 0x1000
	ButtonB           ControllerButton = 0x2000
	ButtonX           ControllerButton = 0x4000
	ButtonY           ControllerButton = 0x8000
)

// MoonlightInput 定义输入处理结构体
type MoonlightInput struct {
	packetType uint16
	socketPath string
	baseURL    string
	client     *http.Client
}

// NewMoonlightInput 创建新的输入处理器
func NewMoonlightInput(socketPath string) *MoonlightInput {
	if socketPath == "" {
		socketPath = "/tmp/wolf.sock"
	}

	// 创建自定义的 Transport
	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}

	return &MoonlightInput{
		packetType: 0x0206, // INPUT_DATA
		socketPath: socketPath,
		baseURL:    "http://localhost", // 使用 localhost 作为 host，实际连接会通过 unix socket
		client: &http.Client{
			Transport: transport,
		},
	}
}

// createHeader 创建输入数据包的头部
func (m *MoonlightInput) createHeader(inputType InputType, dataSize int) []byte {
	buf := new(bytes.Buffer)

	// packet_type (2 bytes) + packet_len (2 bytes) + data_size (4 bytes)
	binary.Write(buf, binary.BigEndian, m.packetType)
	binary.Write(buf, binary.BigEndian, uint16(8+dataSize))
	binary.Write(buf, binary.BigEndian, uint32(dataSize))

	// input_type (4 bytes, little endian)
	binary.Write(buf, binary.LittleEndian, uint32(inputType))

	return buf.Bytes()
}

// MouseMoveRel 构造相对鼠标移动数据包
func (m *MoonlightInput) MouseMoveRel(deltaX, deltaY int16) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, deltaX)
	binary.Write(buf, binary.BigEndian, deltaY)

	data := buf.Bytes()
	header := m.createHeader(MouseMoveRel, len(data))
	return append(header, data...)
}

// MouseMoveAbs 构造绝对鼠标移动数据包
func (m *MoonlightInput) MouseMoveAbs(x, y, width, height int16) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, x)
	binary.Write(buf, binary.BigEndian, y)
	binary.Write(buf, binary.BigEndian, width)
	binary.Write(buf, binary.BigEndian, height)

	data := buf.Bytes()
	header := m.createHeader(MouseMoveAbs, len(data))
	return append(header, data...)
}

// MouseButton 构造鼠标按键数据包
func (m *MoonlightInput) MouseButton(button uint8, isPress bool) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, button)

	data := buf.Bytes()
	inputType := MouseButtonPress
	if !isPress {
		inputType = MouseButtonRelease
	}
	header := m.createHeader(inputType, len(data))
	return append(header, data...)
}

// KeyboardKey 构造键盘按键数据包
func (m *MoonlightInput) KeyboardKey(keyCode uint16, modifiers uint16, isPress bool) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint8(0))  // flags
	binary.Write(buf, binary.BigEndian, keyCode)   // key_code
	binary.Write(buf, binary.BigEndian, modifiers) // modifiers
	binary.Write(buf, binary.BigEndian, uint16(0)) // zero1

	data := buf.Bytes()
	inputType := KeyPress
	if !isPress {
		inputType = KeyRelease
	}
	header := m.createHeader(inputType, len(data))
	return append(header, data...)
}

// MouseScroll 构造鼠标滚轮数据包
func (m *MoonlightInput) MouseScroll(scrollAmount int16) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, scrollAmount)
	binary.Write(buf, binary.BigEndian, scrollAmount)
	binary.Write(buf, binary.BigEndian, int16(0))

	data := buf.Bytes()
	header := m.createHeader(MouseScroll, len(data))
	return append(header, data...)
}

// MouseHScroll 构造鼠标水平滚轮数据包
func (m *MoonlightInput) MouseHScroll(scrollAmount int16) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, scrollAmount)

	data := buf.Bytes()
	header := m.createHeader(MouseHScroll, len(data))
	return append(header, data...)
}

// UTF8Text 构造 UTF8 文本数据包
func (m *MoonlightInput) UTF8Text(text string) []byte {
	// 确保文本不超过32字节
	if len(text) > 32 {
		text = text[:32]
	}

	buf := new(bytes.Buffer)
	textBytes := []byte(text)
	padding := make([]byte, 32-len(textBytes))
	binary.Write(buf, binary.BigEndian, append(textBytes, padding...))

	data := buf.Bytes()
	header := m.createHeader(UTF8Text, len(data))
	return append(header, data...)
}

// SendInput 发送输入数据到指定的会话
func (m *MoonlightInput) SendInput(sessionID string, inputData []byte) error {
	url := fmt.Sprintf("%s/api/v1/sessions/input", m.baseURL)

	data := map[string]string{
		"session_id":       sessionID,
		"input_packet_hex": fmt.Sprintf("%x", inputData),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal input data failed: %v", err)
	}
	log.Infof("=== Send input data: %s", string(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request failed: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	return nil
}

// SendMouseMoveRel 发送相对鼠标移动
func (m *MoonlightInput) SendMouseMoveRel(sessionID string, deltaX, deltaY int16) error {
	data := m.MouseMoveRel(deltaX, deltaY)
	return m.SendInput(sessionID, data)
}

// SendMouseMoveAbs 发送绝对鼠标移动
func (m *MoonlightInput) SendMouseMoveAbs(sessionID string, x, y, width, height int16) error {
	data := m.MouseMoveAbs(x, y, width, height)
	return m.SendInput(sessionID, data)
}

// SendMouseButton 发送鼠标按键
func (m *MoonlightInput) SendMouseButton(sessionID string, button uint8, isPress bool) error {
	data := m.MouseButton(button, isPress)
	return m.SendInput(sessionID, data)
}

// SendKeyboardKey 发送键盘按键
func (m *MoonlightInput) SendKeyboardKey(sessionID string, keyCode uint16, modifiers uint16, isPress bool) error {
	data := m.KeyboardKey(keyCode, modifiers, isPress)
	return m.SendInput(sessionID, data)
}

// SendMouseScroll 发送鼠标滚轮
func (m *MoonlightInput) SendMouseScroll(sessionID string, scrollAmount int16) error {
	data := m.MouseScroll(scrollAmount)
	return m.SendInput(sessionID, data)
}

// SendMouseHScroll 发送鼠标水平滚轮
func (m *MoonlightInput) SendMouseHScroll(sessionID string, scrollAmount int16) error {
	data := m.MouseHScroll(scrollAmount)
	return m.SendInput(sessionID, data)
}

// SendUTF8Text 发送 UTF8 文本
func (m *MoonlightInput) SendUTF8Text(sessionID string, text string) error {
	data := m.UTF8Text(text)
	return m.SendInput(sessionID, data)
}
