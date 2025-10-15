# Lockstep-Core 游戏服务器

一个基于 WebTransport 的通用游戏服务器框架，采用依赖注入设计模式。

## 快速开始

### 安装依赖

```bash
go mod download
```

### 安装 Wire 工具

```bash
go install github.com/google/wire/cmd/wire@latest
```

### 重新生成依赖注入代码（修改依赖后）

```bash
wire
# 或者使用完整路径
~/go/bin/wire
```

### 编译

```bash
go build -o lockstep-server .
```

### 运行

```bash
./lockstep-server
```

服务器将在 `:4433` 端口启动。

## API 端点

### HTTP 端点

- `GET /rooms` - 获取所有房间列表
- `POST /rooms` - 创建新房间
  ```json
  {
    "room_id": "room1"
  }
  ```

### WebTransport 端点

- `/join/{roomID}` - 加入指定房间（升级到 WebTransport 连接）

## 扩展游戏逻辑

### 添加新的游戏逻辑

1. 在 `src/logic/` 目录下创建新的文件
2. 实现你的游戏逻辑（不依赖具体的网络实现）
3. 如果需要新的依赖，在 `src/logic/wire.go` 中添加 provider
4. 运行 `wire` 重新生成依赖注入代码

### 示例：添加自定义会话处理器

```go
// src/logic/custom_handler.go
package logic

import (
    "github.com/quic-go/webtransport-go"
)

type CustomSessionHandler struct {
    // 你的自定义字段
}

func NewCustomSessionHandler() *CustomSessionHandler {
    return &CustomSessionHandler{}
}

func (h *CustomSessionHandler) HandleSession(
    session *webtransport.Session, 
    room *Room, 
    playerID string,
) {
    // 你的自定义逻辑
}
```

然后在 `src/logic/wire.go` 中替换默认实现：

```go
var ProviderSet = wire.NewSet(
    NewRoomManager,
    NewCustomSessionHandler,  // 使用自定义处理器
    wire.Bind(new(RoomManagerInterface), new(*RoomManager)),
    wire.Bind(new(PlayerSessionHandler), new(*CustomSessionHandler)),
)
```

## 开发指南

### 目录职责

- **config/**: 只包含配置相关代码
- **logic/**: 只包含游戏逻辑，不包含网络代码
- **server/**: 只包含服务器和网络相关代码，不包含游戏规则
- **utils/**: 通用工具函数

### 设计原则

1. **单一职责**：每个模块只做一件事
2. **依赖倒置**：高层模块不依赖低层模块，都依赖抽象
3. **接口隔离**：使用接口定义依赖，方便测试和替换
4. **开闭原则**：对扩展开放，对修改封闭



## 项目特点

- ✅ **清晰的分层架构**：服务器层与游戏逻辑层完全解耦
- ✅ **依赖注入**：使用 Google Wire 实现编译时依赖注入
- ✅ **可扩展性**：游戏逻辑模块化，易于扩展
- ✅ **现代协议**：基于 WebTransport (HTTP/3) 实现低延迟通信

## 项目结构

```
lockstep-core/
├── main.go                 # 应用程序入口
├── wire.go                 # Wire 依赖注入配置（不参与编译）
├── wire_gen.go            # Wire 自动生成的代码（不要手动编辑）
├── go.mod                 # Go 模块依赖
├── data/                  # 运行时数据目录
│   └── key/              # TLS 证书存储
├── scripts/              # 工具脚本
└── src/
    ├── config/           # 配置管理模块
    │   ├── config.go    # 服务器配置
    │   └── wire.go      # Wire provider set
    ├── logic/            # 游戏逻辑层（与服务器无关）
    │   ├── interfaces.go        # 游戏逻辑接口定义
    │   ├── room.go             # 房间管理逻辑
    │   ├── room_manager.go     # 房间管理器
    │   ├── session_handler.go  # 玩家会话处理
    │   └── wire.go             # Wire provider set
    ├── server/           # 服务器层（与游戏逻辑无关）
    │   ├── webtransport_server.go  # WebTransport 服务器封装
    │   ├── handlers.go             # HTTP 请求处理器
    │   └── wire.go                 # Wire provider set
    ├── constants/        # 常量定义
    │   └── names.go
    └── utils/           # 工具函数
        ├── localdata.go  # 本地数据目录管理
        └── tls/         # TLS 证书生成工具
            └── generate.go
```

## 架构说明

### 分层设计

1. **配置层 (config/)**
   - 管理服务器配置
   - 处理 TLS 证书加载
   - 提供配置构建器模式

2. **游戏逻辑层 (logic/)**
   - `Room`: 游戏房间实体，管理玩家和消息广播
   - `RoomManager`: 房间管理器，负责房间的创建、获取、删除
   - `PlayerSessionHandler`: 处理玩家会话和消息接收
   - **完全独立于服务器实现**，可以轻松切换到其他网络协议

3. **服务器层 (server/)**
   - `WebTransportServer`: WebTransport 服务器封装
   - `HTTPHandlers`: HTTP 请求处理器
   - **只负责网络通信**，不包含任何游戏逻辑

### 依赖注入

项目使用 [Google Wire](https://github.com/google/wire) 进行编译时依赖注入：

- 每个模块都有自己的 `wire.go` 文件定义 Provider Set
- 根目录的 `wire.go` 组装所有依赖
- `wire_gen.go` 是自动生成的，包含实际的初始化代码

**优势：**
- 编译时检查依赖关系，避免运行时错误
- 零反射，性能优异
- 依赖关系清晰可见
- 易于测试和替换实现
