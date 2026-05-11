# Thrift Client Pool

一个基于对象池模式的 Go Thrift 客户端连接池实现，用于解决高并发场景下 Thrift 连接频繁建立导致的 TIME_WAIT 套接字耗尽问题。

## 问题背景

在使用 Thrift 进行 RPC 通信时，如果频繁创建和销毁连接，会导致大量 TIME_WAIT 状态的套接字积累，最终耗尽系统资源。使用连接池可以复用连接，大幅减少 TIME_WAIT 状态的套接字数量。

**性能对比示例**（1000 个请求）：
- 无连接池：产生约 1000 个 TIME_WAIT 状态的套接字
- 使用连接池：仅产生 2-3 个 TIME_WAIT 状态的套接字

## 特性

- ✅ 基于对象池模式实现，支持连接复用
- ✅ 支持多种 Thrift 协议（Binary、Compact、SimpleJSON、JSON）
- ✅ 支持多种传输层（普通 Socket、SSL、Buffered、Framed）
- ✅ 灵活的配置选项，支持自定义池大小和超时时间
- ✅ 完整的生命周期管理（创建、验证、销毁、回收）

## 安装

```bash
go get github.com/hongweikkx/thrift-client-pool
```

## 使用方式

### 1. 基本用法

```go
package main

import (
	"context"
	pool "github.com/hongweikkx/thrift-client-pool"
	"github.com/hongweikkx/thrift-client-pool/example/gen-go/tutorial"
)

func main() {
	// 创建连接池配置
	config := pool.NewPoolDefaultConfig()
	config.Address = "localhost:9000"
	config.Protocol = pool.POOL_THRIFT_PROTOCOL_BINARY
	config.ObjectPoolConfig.MaxTotal = 10      // 最大连接数
	config.ObjectPoolConfig.MaxIdle = 5        // 最大空闲连接数
	config.ObjectPoolConfig.MinIdle = 2        // 最小空闲连接数

	// 创建连接池
	p := pool.NewPool(config)
	defer p.Destroy(context.Background())

	// 借用连接
	client, err := p.BorrowObject(context.Background())
	if err != nil {
		panic(err)
	}
	defer p.ReturnObject(context.Background(), client)

	// 使用 Thrift 客户端
	calculatorClient := tutorial.NewCalculatorClient(client.Client)
	result, _ := calculatorClient.Add(context.Background(), 1, 1)
	println("1+1 =", result)
}
```

### 2. 配置选项

```go
config := pool.NewPoolDefaultConfig()

// Thrift 配置
config.Address = "127.0.0.1:9000"              // 服务器地址
config.Protocol = pool.POOL_THRIFT_PROTOCOL_COMPACT  // 协议类型
config.Secure = false                          // 是否使用 SSL
config.IsTransportBuffer = true                // 是否使用 Buffered 传输
config.BufferSize = 8192                       // Buffer 大小
config.IsTransportFrame = true                 // 是否使用 Framed 传输

// 对象池配置
config.ObjectPoolConfig.MaxTotal = 50          // 最大连接数
config.ObjectPoolConfig.MaxIdle = 20           // 最大空闲连接数
config.ObjectPoolConfig.MinIdle = 5            // 最小空闲连接数
```

## 协议支持

| 常量 | 值 | 说明 |
|------|-----|------|
| `POOL_THRIFT_PROTOCOL_BINARY` | 0 | 二进制协议（默认）|
| `POOL_THRIFT_PROTOCOL_COMPACT` | 1 | 紧凑协议 |
| `POOL_THRIFT_PROTOCOL_SIMPLEJSON` | 2 | 简单 JSON 协议 |
| `POOL_THRIFT_PROTOCOL_JSON` | 3 | JSON 协议 |

## 快速开始

### 1. 编译

```bash
cd example
go build -o thrift-example
```

### 2. 启动服务器

```bash
./thrift-example
```

### 3. 在另一个终端运行普通客户端

```bash
./thrift-example -mod client
```

### 4. 在另一个终端运行连接池客户端

```bash
./thrift-example -mod client-pool
```

### 5. 运行测试

```bash
go test -v ./...
```

## API 文档

### Pool

```go
// NewPool 创建一个新的连接池
func NewPool(config PoolConfig) *Pool

// NewPoolDefaultConfig 返回默认的连接池配置
func NewPoolDefaultConfig() PoolConfig

// BorrowObject 从池中借用一个连接对象
func (p *Pool) BorrowObject(ctx context.Context) (*ThriftClient, error)

// ReturnObject 将连接对象归还到池中
func (p *Pool) ReturnObject(ctx context.Context, obj *ThriftClient) error

// Destroy 销毁连接池，释放所有资源
func (p *Pool) Destroy(ctx context.Context)
```

### ThriftClient

```go
type ThriftClient struct {
	Client *thrift.TStandardClient // Thrift 标准客户端
}
```

## 何时使用此库

✅ **推荐使用**
- 大规模使用 Thrift 的微服务系统
- 高并发的 Thrift 客户端
- 需要减少 TIME_WAIT 套接字的场景

❌ **不推荐使用**
- 新项目（优先考虑 gRPC）
- 连接数很少的应用
- 已使用服务网格管理连接的系统

## 与其他方案的对比

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|---------|
| **无连接池** | 简单 | 套接字耗尽、性能低 | 小规模应用 |
| **此库（连接池）** | 简单易用、高效、开箱即用 | 需要手动配置 | Thrift 微服务 |
| **gRPC** | 现代、性能好、HTTP/2 | 需要重写服务 | 新项目 |
| **Service Mesh** | 透明、功能强大 | 复杂、资源消耗大 | 大型 K8s 环境 |

## 依赖

- Go 1.25 或更高版本
- Apache Thrift v0.20.0
- go-commons-pool/v2 v2.1.1

## 许可证

MIT

## 作者

原作者: hongweikkx

## 常见问题

### Q: 连接池的最优大小是多少？
A: 这取决于你的应用场景。一般建议 MaxTotal 设置为预期并发数的 1.5-2 倍，MinIdle 设置为 1-5，MaxIdle 设置为 MaxTotal 的 30-50%。

### Q: 如何监控连接池状态？
A: 可以通过 `commons-pool` 提供的 stats API 获取池的状态信息。详见 commons-pool 文档。

### Q: 连接池是线程安全的吗？
A: 是的。底层 commons-pool 已经处理了并发安全。

## 更新日志

### v1.0.0 (2026-05-09)
- ✨ 升级 Go 到 1.25
- ✨ 升级 Apache Thrift 到 v0.20.0
- 📝 完善文档和注释
- 🐛 改进错误处理

---

更多信息，请参考 `example/client.go` 中的示例代码。
