package pool

import (
	"context"
	"fmt"
	"testing"

	commonPool "github.com/jolestar/go-commons-pool/v2"
)

// MockThriftClientFactory 是一个用于测试的mock工厂
type MockThriftClientFactory struct {
	MakeObjectCount    int
	DestroyObjectCount int
	ValidateObjectCount int
	FailOnMake         bool
	FailOnDestroy      bool
}

func (m *MockThriftClientFactory) MakeObject(ctx context.Context) (*commonPool.PooledObject, error) {
	m.MakeObjectCount++
	if m.FailOnMake {
		return nil, fmt.Errorf("mock: failed to create object")
	}
	client := &ThriftClient{}
	return commonPool.NewPooledObject(client), nil
}

func (m *MockThriftClientFactory) DestroyObject(ctx context.Context, object *commonPool.PooledObject) error {
	m.DestroyObjectCount++
	if m.FailOnDestroy {
		return fmt.Errorf("mock: failed to destroy object")
	}
	return nil
}

func (m *MockThriftClientFactory) ValidateObject(ctx context.Context, object *commonPool.PooledObject) bool {
	m.ValidateObjectCount++
	return true
}

func (m *MockThriftClientFactory) ActivateObject(ctx context.Context, object *commonPool.PooledObject) error {
	return nil
}

func (m *MockThriftClientFactory) PassivateObject(ctx context.Context, object *commonPool.PooledObject) error {
	return nil
}

// TestNewPoolDefaultConfig 测试默认配置
func TestNewPoolDefaultConfig(t *testing.T) {
	config := NewPoolDefaultConfig()
	
	if config.Address != "localhost:9000" {
		t.Errorf("expected address localhost:9000, got %s", config.Address)
	}
	if config.Protocol != POOL_THRIFT_PROTOCOL_BINARY {
		t.Errorf("expected protocol BINARY, got %d", config.Protocol)
	}
	if config.Secure != false {
		t.Errorf("expected Secure to be false")
	}
	if config.IsTransportBuffer != false {
		t.Errorf("expected IsTransportBuffer to be false")
	}
	if config.IsTransportFrame != false {
		t.Errorf("expected IsTransportFrame to be false")
	}
	if config.TConfiguration == nil {
		t.Errorf("expected TConfiguration to be initialized")
	}
	if config.ObjectPoolConfig == nil {
		t.Errorf("expected ObjectPoolConfig to be initialized")
	}
}

// TestProtocolConstants 测试协议常量
func TestProtocolConstants(t *testing.T) {
	tests := []struct {
		name     string
		protocol int
		expected int
	}{
		{"BINARY", POOL_THRIFT_PROTOCOL_BINARY, 0},
		{"COMPACT", POOL_THRIFT_PROTOCOL_COMPACT, 1},
		{"SIMPLEJSON", POOL_THRIFT_PROTOCOL_SIMPLEJSON, 2},
		{"JSON", POOL_THRIFT_PROTOCOL_JSON, 3},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.protocol != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, tt.protocol)
			}
		})
	}
}

// TestPoolCreation 测试池的创建
func TestPoolCreation(t *testing.T) {
	config := NewPoolDefaultConfig()
	pool := NewPool(config)
	
	if pool == nil {
		t.Errorf("expected pool to be created, got nil")
	}
	if pool.Config.Address != config.Address {
		t.Errorf("expected address %s, got %s", config.Address, pool.Config.Address)
	}
	if pool.ObjectPool == nil {
		t.Errorf("expected ObjectPool to be initialized")
	}
}

// TestBorrowAndReturnObject 测试借用和归还对象
func TestBorrowAndReturnObject(t *testing.T) {
	config := NewPoolDefaultConfig()
	config.ObjectPoolConfig.MaxIdle = 5
	config.ObjectPoolConfig.MaxTotal = 10
	
	// 创建一个使用mock工厂的对象池
	mockFactory := &MockThriftClientFactory{}
	objPool := commonPool.NewObjectPool(context.Background(), mockFactory, config.ObjectPoolConfig)
	pool := &Pool{
		Config:     config,
		ObjectPool: objPool,
	}
	
	ctx := context.Background()
	
	// 借用一个对象
	client, err := pool.BorrowObject(ctx)
	if err != nil {
		t.Errorf("failed to borrow object: %v", err)
	}
	if client == nil {
		t.Errorf("expected client to be non-nil")
	}
	if mockFactory.MakeObjectCount != 1 {
		t.Errorf("expected MakeObject to be called once, got %d calls", mockFactory.MakeObjectCount)
	}
	
	// 归还对象
	err = pool.ReturnObject(ctx, client)
	if err != nil {
		t.Errorf("failed to return object: %v", err)
	}
}

// TestBorrowMultipleObjects 测试借用多个对象
func TestBorrowMultipleObjects(t *testing.T) {
	config := NewPoolDefaultConfig()
	config.ObjectPoolConfig.MaxIdle = 10
	config.ObjectPoolConfig.MaxTotal = 20
	
	mockFactory := &MockThriftClientFactory{}
	objPool := commonPool.NewObjectPool(context.Background(), mockFactory, config.ObjectPoolConfig)
	pool := &Pool{
		Config:     config,
		ObjectPool: objPool,
	}
	
	ctx := context.Background()
	
	// 借用多个对象
	clients := make([]*ThriftClient, 5)
	for i := 0; i < 5; i++ {
		client, err := pool.BorrowObject(ctx)
		if err != nil {
			t.Errorf("failed to borrow object %d: %v", i, err)
		}
		if client == nil {
			t.Errorf("expected client %d to be non-nil", i)
		}
		clients[i] = client
	}
	
	if mockFactory.MakeObjectCount != 5 {
		t.Errorf("expected MakeObject to be called 5 times, got %d calls", mockFactory.MakeObjectCount)
	}
	
	// 归还所有对象
	for i, client := range clients {
		err := pool.ReturnObject(ctx, client)
		if err != nil {
			t.Errorf("failed to return object %d: %v", i, err)
		}
	}
}

// TestBorrowObjectReuse 测试对象复用
func TestBorrowObjectReuse(t *testing.T) {
	config := NewPoolDefaultConfig()
	config.ObjectPoolConfig.MaxIdle = 10
	config.ObjectPoolConfig.MaxTotal = 20
	
	mockFactory := &MockThriftClientFactory{}
	objPool := commonPool.NewObjectPool(context.Background(), mockFactory, config.ObjectPoolConfig)
	pool := &Pool{
		Config:     config,
		ObjectPool: objPool,
	}
	
	ctx := context.Background()
	
	// 第一次借用
	client1, err := pool.BorrowObject(ctx)
	if err != nil {
		t.Errorf("failed to borrow first object: %v", err)
	}
	
	// 归还第一个对象
	err = pool.ReturnObject(ctx, client1)
	if err != nil {
		t.Errorf("failed to return first object: %v", err)
	}
	
	// 第二次借用（应该复用第一个对象）
	client2, err := pool.BorrowObject(ctx)
	if err != nil {
		t.Errorf("failed to borrow second object: %v", err)
	}
	
	if mockFactory.MakeObjectCount != 1 {
		t.Errorf("expected MakeObject to be called only once (reuse), got %d calls", mockFactory.MakeObjectCount)
	}
	
	// 归还第二个对象
	err = pool.ReturnObject(ctx, client2)
	if err != nil {
		t.Errorf("failed to return second object: %v", err)
	}
}

// TestPoolDestroy 测试池的销毁
func TestPoolDestroy(t *testing.T) {
	config := NewPoolDefaultConfig()
	config.ObjectPoolConfig.MaxIdle = 5
	config.ObjectPoolConfig.MaxTotal = 10
	
	mockFactory := &MockThriftClientFactory{}
	objPool := commonPool.NewObjectPool(context.Background(), mockFactory, config.ObjectPoolConfig)
	pool := &Pool{
		Config:     config,
		ObjectPool: objPool,
	}
	
	ctx := context.Background()
	
	// 借用和归还一个对象
	client, err := pool.BorrowObject(ctx)
	if err != nil {
		t.Errorf("failed to borrow object: %v", err)
	}
	
	err = pool.ReturnObject(ctx, client)
	if err != nil {
		t.Errorf("failed to return object: %v", err)
	}
	
	// 销毁池
	pool.Destroy(ctx)
	
	// 不应该能再借用对象（池已关闭）
	_, err = pool.BorrowObject(ctx)
	if err == nil {
		t.Errorf("expected error after pool destruction, got nil")
	}
}

// TestConfigCustomization 测试配置自定义
func TestConfigCustomization(t *testing.T) {
	config := NewPoolDefaultConfig()
	
	// 自定义配置
	config.Address = "192.168.1.1:9090"
	config.Protocol = POOL_THRIFT_PROTOCOL_COMPACT
	config.Secure = true
	config.IsTransportBuffer = true
	config.BufferSize = 4096
	config.IsTransportFrame = true
	config.ObjectPoolConfig.MaxIdle = 20
	config.ObjectPoolConfig.MaxTotal = 50
	
	// 验证配置
	if config.Address != "192.168.1.1:9090" {
		t.Errorf("expected address 192.168.1.1:9090, got %s", config.Address)
	}
	if config.Protocol != POOL_THRIFT_PROTOCOL_COMPACT {
		t.Errorf("expected protocol COMPACT, got %d", config.Protocol)
	}
	if config.Secure != true {
		t.Errorf("expected Secure to be true")
	}
	if config.IsTransportBuffer != true {
		t.Errorf("expected IsTransportBuffer to be true")
	}
	if config.BufferSize != 4096 {
		t.Errorf("expected BufferSize to be 4096, got %d", config.BufferSize)
	}
	if config.IsTransportFrame != true {
		t.Errorf("expected IsTransportFrame to be true")
	}
	if config.ObjectPoolConfig.MaxIdle != 20 {
		t.Errorf("expected MaxIdle to be 20, got %d", config.ObjectPoolConfig.MaxIdle)
	}
	if config.ObjectPoolConfig.MaxTotal != 50 {
		t.Errorf("expected MaxTotal to be 50, got %d", config.ObjectPoolConfig.MaxTotal)
	}
}

// TestThriftClientStruct 测试ThriftClient结构体
func TestThriftClientStruct(t *testing.T) {
	client := &ThriftClient{}
	
	if client.transPort != nil {
		t.Errorf("expected transPort to be nil initially")
	}
	if client.Client != nil {
		t.Errorf("expected Client to be nil initially")
	}
}

// TestMultiplePoolInstances 测试多个池实例独立性
func TestMultiplePoolInstances(t *testing.T) {
	config1 := NewPoolDefaultConfig()
	config1.Address = "host1:9000"
	config1.Protocol = POOL_THRIFT_PROTOCOL_BINARY
	
	config2 := NewPoolDefaultConfig()
	config2.Address = "host2:9000"
	config2.Protocol = POOL_THRIFT_PROTOCOL_COMPACT
	
	pool1 := NewPool(config1)
	pool2 := NewPool(config2)
	
	if pool1.Config.Address == pool2.Config.Address {
		t.Errorf("pools should have different addresses")
	}
	if pool1.Config.Protocol == pool2.Config.Protocol {
		t.Errorf("pools should have different protocols")
	}
}

// TestContextCancellation 测试context取消对BorrowObject的影响
func TestContextCancellation(t *testing.T) {
	config := NewPoolDefaultConfig()
	config.ObjectPoolConfig.MaxIdle = 1
	config.ObjectPoolConfig.MaxTotal = 1
	config.ObjectPoolConfig.BlockWhenExhausted = true
	
	mockFactory := &MockThriftClientFactory{}
	objPool := commonPool.NewObjectPool(context.Background(), mockFactory, config.ObjectPoolConfig)
	pool := &Pool{
		Config:     config,
		ObjectPool: objPool,
	}
	
	ctx := context.Background()
	
	// 借用唯一的对象
	client1, err := pool.BorrowObject(ctx)
	if err != nil {
		t.Errorf("failed to borrow first object: %v", err)
	}
	
	// 创建一个已取消的context
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()
	
	// 尝试在已取消的context上借用对象（应该立即返回错误因为context已取消）
	_, err = pool.BorrowObject(cancelCtx)
	if err == nil {
		t.Errorf("expected error with cancelled context, got nil")
	}
	
	// 归还第一个对象
	pool.ReturnObject(ctx, client1)
}

// TestBorrowObjectTypeAssertion 测试BorrowObject返回正确的类型
func TestBorrowObjectTypeAssertion(t *testing.T) {
	config := NewPoolDefaultConfig()
	config.ObjectPoolConfig.MaxIdle = 5
	config.ObjectPoolConfig.MaxTotal = 10
	
	mockFactory := &MockThriftClientFactory{}
	objPool := commonPool.NewObjectPool(context.Background(), mockFactory, config.ObjectPoolConfig)
	pool := &Pool{
		Config:     config,
		ObjectPool: objPool,
	}
	
	ctx := context.Background()
	client, err := pool.BorrowObject(ctx)
	if err != nil {
		t.Errorf("failed to borrow object: %v", err)
	}
	
	// 验证返回的是正确的类型
	if client == nil {
		t.Errorf("expected client to be non-nil")
	}
	
	_, ok := interface{}(client).(*ThriftClient)
	if !ok {
		t.Errorf("expected client to be of type *ThriftClient")
	}
	
	pool.ReturnObject(ctx, client)
}
