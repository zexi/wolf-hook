# Moonlight 客户端包

这个包提供了与 Moonlight/Wolf 服务器进行配对和通信的客户端功能。

## 功能特性

- **配对功能**: 完整的 5 阶段配对流程
- **服务器信息**: 获取服务器状态和配置信息
- **应用列表**: 获取可用的应用程序列表
- **应用启动**: 启动指定的应用程序
- **环境变量配置**: 支持通过环境变量配置连接参数

## 快速开始

### 基本使用

```go
package main

import (
    "log"
    "github.com/zexi/wolf-hook/pkg/moonlight/client"
)

func main() {
    // 创建客户端
    moonlightClient := client.NewMoonlightClient("192.168.1.100", 20008, "my_client_001")
    
    // 执行配对
    if err := moonlightClient.PairAndConnect(); err != nil {
        log.Fatalf("配对失败: %v", err)
    }
    
    // 获取服务器信息
    serverInfo, err := moonlightClient.GetServerInfo()
    if err != nil {
        log.Printf("获取服务器信息失败: %v", err)
    } else {
        log.Printf("服务器: %s, 版本: %s", serverInfo.Hostname, serverInfo.AppVersion)
    }
    
    // 获取应用列表
    appList, err := moonlightClient.GetAppList()
    if err != nil {
        log.Printf("获取应用列表失败: %v", err)
    } else {
        for _, app := range appList.Apps {
            log.Printf("应用: %s (ID: %s)", app.AppTitle, app.ID)
        }
    }
    
    // 启动应用
    launchResult, err := moonlightClient.LaunchApp("1", 1920, 1080, 60)
    if err != nil {
        log.Printf("启动应用失败: %v", err)
    } else {
        log.Printf("启动成功，会话ID: %s", launchResult.SessionURL0)
    }
}
```

### 环境变量配置

```bash
# 设置环境变量
export MOONLIGHT_HOST_IP=192.168.1.100
export WOLF_HTTP_PORT=20008
export WOLF_CLIENT_ID=my_client_001

# 运行程序
./wolf-session
```

## API 参考

### 结构体

#### `MoonlightClient`
主要的客户端结构体，提供所有功能接口。

#### `ServerInfo`
服务器信息结构体，包含：
- `Hostname`: 服务器主机名
- `AppVersion`: 应用版本
- `State`: 服务器状态
- `SupportedDisplayMode`: 支持的显示模式

#### `App`
应用程序结构体，包含：
- `ID`: 应用ID
- `AppTitle`: 应用标题
- `IsHdrSupported`: 是否支持HDR

#### `AppList`
应用程序列表结构体，包含：
- `Apps`: 应用列表数组

#### `LaunchResponse`
启动应用响应结构体，包含：
- `SessionURL0`: 会话URL
- `GameSession`: 游戏会话ID

### 方法

#### `NewMoonlightClient(hostIP string, httpPort int, clientID string) *MoonlightClient`
创建新的客户端实例。

**参数:**
- `hostIP`: 服务器IP地址
- `httpPort`: HTTP端口（HTTPS端口自动计算为HTTP端口-5）
- `clientID`: 客户端唯一标识符

#### `PairAndConnect() error`
执行完整的配对流程并建立连接。

#### `GetServerInfo() (*ServerInfo, error)`
获取服务器信息。

#### `GetAppList() (*AppList, error)`
获取应用程序列表。

#### `LaunchApp(appID string, width, height, refreshRate int) (*LaunchResponse, error)`
启动指定的应用程序。

**参数:**
- `appID`: 应用ID
- `width`: 显示宽度
- `height`: 显示高度
- `refreshRate`: 刷新率

## 配对流程

客户端使用 5 阶段配对流程：

1. **阶段1**: 发送客户端证书和 salt
2. **阶段2**: 发送客户端挑战
3. **阶段3**: 发送服务器挑战响应
4. **阶段4**: 发送客户端配对密钥
5. **阶段5**: HTTPS 验证

配对成功后，客户端可以使用 HTTPS 接口与服务器通信。

## 错误处理

所有方法都返回 `error` 类型，建议进行适当的错误处理：

```go
if err := client.PairAndConnect(); err != nil {
    log.Printf("配对失败: %v", err)
    return
}
```

## 注意事项

1. **证书认证**: HTTPS 接口需要客户端证书认证
2. **端口计算**: HTTPS 端口 = HTTP 端口 - 5
3. **超时设置**: 默认请求超时为 30 秒
4. **TLS 验证**: 默认跳过 TLS 证书验证（`InsecureSkipVerify: true`）

## 示例项目

完整的使用示例请参考 `cmd/wolf-session/main.go`。 