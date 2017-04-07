//
// @(#)Pool_test.go  Date: 3/23/17
//
//MIT License
//Copyright (c) 2017 Anirudh Vyas
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.
package networking

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"testing"
)

var p *Pool

// TODO: Figure out if Go has some sort of setUp() bound to 'a' group of tests already.
// NOTE: Assume that netcat is already installed on the machine, if you are on os x
// do brew install nc or netcat.
func SetUp() {
	// setup about 10 running servers
	for i := 0; i < 10; i++ {
		port := 12345
		c := exec.Command("nc", "-l", "-p", fmt.Sprintf("%v", port+i))
		c.Start()
	}
}

// Tests new pool so created.
func TestPool_New(t *testing.T) {
	cfg := PoolCfg{6, 10, func(cfg ConnCfg) (net.Conn, error) {
		log.Printf("cfg=%+v\n", cfg)
		c, err := net.Dial(cfg.Protocol, cfg.HostPort)
		if err != nil {
			log.Printf("failed to initialize connection for : %v due to: %v\n", c, err)
		}
		return c, nil
	}, []ConnCfg{
		{HostPort: "127.0.0.1:12345", Protocol: "tcp"},
		{HostPort: "127.0.0.1:12346", Protocol: "tcp"},
		{HostPort: "127.0.0.1:12347", Protocol: "tcp"},
		{HostPort: "127.0.0.1:12348", Protocol: "tcp"},
		{HostPort: "127.0.0.1:12349", Protocol: "tcp"},
		{HostPort: "127.0.0.1:12350", Protocol: "tcp"},
	}, true}
	_, p = New(cfg)
}

// Close a particular connection on pool
func TestPool_Close(t *testing.T) {
	SetUp()
	TestPool_New(t)
	fmt.Printf("pool size is=%+v\n", p.Size())
	c, _ := p.Get(ConnCfg{HostPort: "127.0.0.1:12345", Protocol: "tcp"})
	c.Close()
	fmt.Printf("pool size is=%+v\n", p.Size())
}
