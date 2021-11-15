/*
 copy from https://thrift.apache.org/tutorial/go.html
 modified by hongweigaokkx@163.com
*/
package main

import (
	"context"
	"fmt"
	pool2 "github.com/hongweikkx/thrift-client-pool"
	"sync"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/hongweikkx/thrift-client-pool/example/gen-go/tutorial"
)

var defaultCtx = context.Background()

func handleClient(client *tutorial.CalculatorClient) (err error) {
	client.Ping(defaultCtx)
	fmt.Println("ping()")

	sum, _ := client.Add(defaultCtx, 1, 1)
	fmt.Print("1+1=", sum, "\n")

	work := tutorial.NewWork()
	work.Op = tutorial.Operation_DIVIDE
	work.Num1 = 1
	work.Num2 = 0
	quotient, err := client.Calculate(defaultCtx, 1, work)
	if err != nil {
		switch v := err.(type) {
		case *tutorial.InvalidOperation:
			fmt.Println("Invalid operation:", v)
		default:
			fmt.Println("Error during operation:", err)
		}
	} else {
		fmt.Println("Whoa we can divide by 0 with new value:", quotient)
	}

	work.Op = tutorial.Operation_SUBTRACT
	work.Num1 = 15
	work.Num2 = 10
	diff, err := client.Calculate(defaultCtx, 1, work)
	if err != nil {
		switch v := err.(type) {
		case *tutorial.InvalidOperation:
			fmt.Println("Invalid operation:", v)
		default:
			fmt.Println("Error during operation:", err)
		}
		return err
	} else {
		fmt.Print("15-10=", diff, "\n")
	}
	log, err := client.GetStruct(defaultCtx, 1)
	if err != nil {
		fmt.Println("Unable to get struct:", err)
		return err
	} else {
		fmt.Println("Check log:", log.Value)
	}
	return err
}

func runClients(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, addr string, secure bool, cfg *thrift.TConfiguration, num int) error {
	now := time.Now().UnixNano()
	for i := 0; i < num; i++ {
		err := runClient(transportFactory, protocolFactory, addr, secure, cfg)
		if err != nil {
			fmt.Printf("runClient error:%+v\n", err)
		}
	}
	fmt.Printf("finish time:%+vms\n", (time.Now().UnixNano() - now)/1000000)
	return nil
}

func runClient(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, addr string, secure bool, cfg *thrift.TConfiguration) error {
	var transport thrift.TTransport
	if secure {
		transport = thrift.NewTSSLSocketConf(addr, cfg)
	} else {
		transport = thrift.NewTSocketConf(addr, cfg)
	}
	transport, err := transportFactory.GetTransport(transport)
	if err != nil {
		return err
	}
	defer transport.Close()
	if err := transport.Open(); err != nil {
		return err
	}
	iprot := protocolFactory.GetProtocol(transport)
	oprot := protocolFactory.GetProtocol(transport)
	return handleClient(tutorial.NewCalculatorClient(thrift.NewTStandardClient(iprot, oprot)))
}

func runClientsPool(addr string, protocol int, framed bool, buffered bool,secure bool, num int) error {
	conf := pool2.NewPoolDefaultConfig()
	conf.Address = addr
	conf.Protocol = protocol
	conf.IsTransportFrame = framed
	conf.IsTransportBuffer = buffered
	conf.Secure = secure

	pool := pool2.NewPool(conf)
	now := time.Now().UnixNano()
	var wg sync.WaitGroup
	ctx := context.Background()
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func(i int) {
			obj, err := pool.BorrowObject(ctx)
			if err != nil {
				fmt.Println("brrow obj err", err.Error())
			}
			err = handleClient(tutorial.NewCalculatorClient(obj.Client))
			if err != nil {
				fmt.Println("handle client err", err.Error())
			}
			err = pool.ReturnObject(context.Background(), obj)
			if err != nil {
				fmt.Println("return client err", err.Error())
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Printf("finish time:%+vms\n", (time.Now().UnixNano() - now)/1000000)
	pool.Destroy(ctx)
	return nil
}
