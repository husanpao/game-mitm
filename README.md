# game-mitm

## 项目概述
`game-mitm` 是一个用 Go 语言实现的中间人代理服务器项目，主要用于拦截、修改和转发网络请求。

## 功能简介
### 协议支持
- **HTTP/HTTPS 代理**
    - 处理普通 HTTP 请求及 HTTPS 隧道请求，读取客户端请求，可修改请求体和请求头后转发到目标服务器。
    - 自动管理 CA 证书，保障 HTTPS 连接安全。
- **WebSocket (WS) / WebSocket Secure (WSS) 支持**
    - 能检测 WebSocket 升级请求，对 WS 和 WSS 连接进行相应处理。

### 中间人监听与修改
- 允许用户注册 `Handle` 函数，对特定 URL 的请求和响应进行拦截与修改。



## 使用方法
1. 克隆项目：
   ```sh
   git clone https://github.com/husanpao/game-mitm.git
   cd game-mitm
   ```
2. 运行项目：
   ```sh
   go run cmd/main.go
   ```
3. 注册处理函数示例：
   ```go
   proxy := gamemitm.NewProxy()
   proxy.OnRequest("example.com").Do(func(body []byte, ctx *gamemitm.ProxyCtx) []byte {
       // 处理请求体
       fmt.Println("Handling request for example.com")
       return body
   })
   ```

## 贡献指南
1. 安装 Go 1.20 或更高版本。
2. 克隆项目，创建新分支：
   ```sh
   git checkout -b feature/your-feature-name
   ```
3. 进行代码修改和测试，提交并推送：
   ```sh
   git add .
   git commit -m "Add your commit message here"
   git push origin feature/your-feature-name
   ```
4. 创建 Pull Request 等待审核合并。

