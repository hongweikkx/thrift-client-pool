# thrift-client-pool
thrift client pool

# usage
参考: [example/client.go](https://github.com/hongweikkx/thrift-client-pool/blob/03af42bcd3c9d38fc8d04318c09a1c248186b9e9/example/client.go#L100)

# example
1. 编译: `cd example && go build`
2. 启动server `./example`
3. 启动普通client  `./example -mod client`  or  启动池化client `./example -mod client-pool`
4. test `go test .`

# 效果
默认1000个请求; 
不使用连接池TIME_WAIT状态数增加了1000个, 而使用连接池TIME_WAIT状态数增加了2个
