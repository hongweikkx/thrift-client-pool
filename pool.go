package pool

import (
	"context"
	"errors"
	"github.com/apache/thrift/lib/go/thrift"
	commonPool "github.com/jolestar/go-commons-pool/v2"
)

type PoolConfig struct {
	// thrift conf
	Address  string
	Protocol int
	Secure bool
	IsTransportBuffer bool
	BufferSize int
	IsTransportFrame bool
	*thrift.TConfiguration
	*commonPool.ObjectPoolConfig
}

const (
	POOL_THRIFT_PROTOCOL_BINARY = iota
	POOL_THRIFT_PROTOCOL_COMPACT
	POOL_THRIFT_PROTOCOL_SIMPLEJSON
	POOL_THRIFT_PROTOCOL_JSON
)

func NewPoolDefaultConfig() PoolConfig{
	return PoolConfig{
		Address:          "localhost:9000",
		Protocol:         POOL_THRIFT_PROTOCOL_BINARY,
		Secure:           false,
		IsTransportBuffer: false,
		BufferSize: 0,
		IsTransportFrame: false,
		TConfiguration:   &thrift.TConfiguration{},
		ObjectPoolConfig: commonPool.NewDefaultPoolConfig(),
	}
}

type Pool struct {
	Config PoolConfig
	*commonPool.ObjectPool
}

func NewPool(config PoolConfig) *Pool{
	return &Pool{
		Config:  config,
		ObjectPool: NewObjPool(config),
	}
}

func (p *Pool) BorrowObject(ctx context.Context) (*ThriftClient, error){
	client, err := p.ObjectPool.BorrowObject(ctx)
	if err != nil {
		return nil, err
	}
	return client.(*ThriftClient), nil
}

func (p *Pool) Destroy(ctx context.Context) {
	p.ObjectPool.Close(ctx)
}


type ThriftClient struct {
	transPort thrift.TTransport
	Client    *thrift.TStandardClient
}

type ThriftClientFactory struct {
	conf PoolConfig
}

func (f *ThriftClientFactory) MakeObject(ctx context.Context) (*commonPool.PooledObject, error) {
	var transportFactory thrift.TTransportFactory
	if f.conf.IsTransportBuffer {
		transportFactory = thrift.NewTBufferedTransportFactory(f.conf.BufferSize)
	} else {
		transportFactory = thrift.NewTTransportFactory()
	}
	if f.conf.IsTransportFrame {
		transportFactory = thrift.NewTFramedTransportFactoryConf(transportFactory, f.conf.TConfiguration)
	}

	var transport thrift.TTransport
	if f.conf.Secure {
		transport = thrift.NewTSSLSocketConf(f.conf.Address, f.conf.TConfiguration)
	} else {
		transport = thrift.NewTSocketConf(f.conf.Address, f.conf.TConfiguration)
	}
	transport, err := transportFactory.GetTransport(transport)
	if err != nil {
		return nil, err
	}
	err = transport.Open()
	if err != nil {
		return nil, err
	}

	var protocolFactory thrift.TProtocolFactory
	switch f.conf.Protocol {
	case POOL_THRIFT_PROTOCOL_COMPACT:
		protocolFactory = thrift.NewTCompactProtocolFactoryConf(nil)
	case POOL_THRIFT_PROTOCOL_SIMPLEJSON:
		protocolFactory = thrift.NewTSimpleJSONProtocolFactoryConf(nil)
	case POOL_THRIFT_PROTOCOL_JSON:
		protocolFactory = thrift.NewTJSONProtocolFactory()
	case POOL_THRIFT_PROTOCOL_BINARY:
		protocolFactory = thrift.NewTBinaryProtocolFactoryConf(nil)
	default:
		return nil, errors.New("invalid protol conf")
	}

	iprot := protocolFactory.GetProtocol(transport)
	oprot := protocolFactory.GetProtocol(transport)
	client := thrift.NewTStandardClient(iprot, oprot)
	obj := commonPool.NewPooledObject(&ThriftClient{transPort: transport, Client: client})
	return obj, nil
}

func (f *ThriftClientFactory) DestroyObject(ctx context.Context, object *commonPool.PooledObject) error {
	return object.Object.(*ThriftClient).transPort.Close()
}

func (f *ThriftClientFactory) ValidateObject(ctx context.Context, object *commonPool.PooledObject) bool {
	return object.Object.(*ThriftClient).transPort.IsOpen()
}

func (f *ThriftClientFactory) ActivateObject(ctx context.Context, object *commonPool.PooledObject) error {
	// do activate
	return nil
}

func (f *ThriftClientFactory) PassivateObject(ctx context.Context, object *commonPool.PooledObject) error {
	// do passivate
	return nil
}

func NewObjPool(config PoolConfig) *commonPool.ObjectPool{
	factory := ThriftClientFactory{conf: config}
	return commonPool.NewObjectPool(context.Background(), &factory, config.ObjectPoolConfig)
}