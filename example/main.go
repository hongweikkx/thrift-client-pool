/*
 copy from https://thrift.apache.org/tutorial/go.html
 modified by hongweigaokkx@163.com
*/
package main

import (
	"fmt"
	pool2 "github.com/hongweikkx/thrift-client-pool"
)

/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import (
	"crypto/tls"
	"flag"
	"os"

	"github.com/apache/thrift/lib/go/thrift"
)

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

func main() {
	flag.Usage = Usage
	mod := flag.String("mod", "server", "Run server or client")
	protocol := flag.String("P", "binary", "Specify the protocol (binary, compact, json, simplejson)")
	framed := flag.Bool("framed", false, "Use framed transport")
	buffered := flag.Bool("buffered", false, "Use buffered transport")
	addr := flag.String("addr", "localhost:9090", "Address to listen to")
	secure := flag.Bool("secure", false, "Use tls secure transport")
	num := flag.Int("num", 1000, "test req num")

	flag.Parse()
	var protocolFactory thrift.TProtocolFactory
	protocolInt := pool2.POOL_THRIFT_PROTOCOL_BINARY
	switch *protocol {
	case "compact":
		protocolFactory = thrift.NewTCompactProtocolFactoryConf(nil)
		protocolInt = pool2.POOL_THRIFT_PROTOCOL_COMPACT
	case "simplejson":
		protocolFactory = thrift.NewTSimpleJSONProtocolFactoryConf(nil)
		protocolInt = pool2.POOL_THRIFT_PROTOCOL_SIMPLEJSON
	case "json":
		protocolFactory = thrift.NewTJSONProtocolFactory()
		protocolInt = pool2.POOL_THRIFT_PROTOCOL_JSON
	case "binary", "":
		protocolFactory = thrift.NewTBinaryProtocolFactoryConf(nil)
		protocolInt = pool2.POOL_THRIFT_PROTOCOL_BINARY
	default:
		fmt.Fprint(os.Stderr, "Invalid protocol specified", protocol, "\n")
		Usage()
		os.Exit(1)
	}

	if *mod == "client-pool" {
		if err := runClientsPool(*addr, protocolInt, *framed, *buffered, *secure, *num); err != nil {
			fmt.Println("error running client:", err)
		}
	}else {
		var transportFactory thrift.TTransportFactory
		cfg := &thrift.TConfiguration{
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		if *buffered {
			transportFactory = thrift.NewTBufferedTransportFactory(8192)
		} else {
			transportFactory = thrift.NewTTransportFactory()
		}

		if *framed {
			transportFactory = thrift.NewTFramedTransportFactoryConf(transportFactory, cfg)
		}
		switch *mod {
		case "client":
			if err := runClients(transportFactory, protocolFactory,*addr, *secure, cfg, *num); err != nil {
				fmt.Println("error running client:", err)
			}
		default:
			if err := runServer(transportFactory, protocolFactory, *addr, *secure); err != nil {
				fmt.Println("error running server:", err)
			}
		}
	}
}