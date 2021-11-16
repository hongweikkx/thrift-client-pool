package main

import (
	"github.com/apache/thrift/lib/go/thrift"
	pool2 "github.com/hongweikkx/thrift-client-pool"
	"testing"
)

func TestRunClientsPool(t *testing.T) {
	err := runClientsPool("localhost:9090", pool2.POOL_THRIFT_PROTOCOL_BINARY, false, false, false, 1000)
	if err != nil {
		t.Fatalf("err:%+v", err)
	}
}

func TestRunClients(t *testing.T) {
	transportFactory := thrift.NewTTransportFactory()
	protocolFactory := thrift.NewTBinaryProtocolFactoryConf(nil)
	if err := runClients(transportFactory, protocolFactory,"localhost:9090", false, &thrift.TConfiguration{}, 1000); err != nil {
		t.Fatalf("err:%+v", err)
	}
}