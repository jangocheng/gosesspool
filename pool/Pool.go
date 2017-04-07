//
// @(#)Pool.go  Date: 3/23/17
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

// TODO: This library needs to be tested for leaks.
// Note - this is taken mostly from ideas and improved upon
// using https://github.com/fatih as a sample pooling lib
// Differences
// 1.Are mostly in how connections are configured from outside
// 2. Connection configurators
// 3. Adjustments to how factory is invoked for each connection to be live.
// 4. Pre and Post initializers if any available to run for each of the resources
// 	- Connection `PoolConn`
//	- `Pool`
// 5. Ping and Keep-Alive semantics added
// 6. Connection and Pool Debug Mode enables more logs to show incase a leak is suspected
// from the client application.
package networking

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// PoolConn is a wrapper around net.Conn to modify the the behavior of
// net.Conn's Close() method.
type PoolConn struct {
	net.Conn // makes it an enhanced connection.
	mu       sync.RWMutex
	pool     *Pool
	unusable bool
}

// Connection configuration is a dependent of `PoolCfg`.
type ConnCfg struct {
	HostPort            string
	Protocol            string
	PreConnInitializers []Before
}

// Connection creator.
type ConnectionFactory func(c ConnCfg) (net.Conn, error)

// TODO: Figure out a better name for this.
type Before func(interface{}) (interface{}, error)
type After func(interface{}) (interface{}, error)

// Pool bootstrapping configuration
type PoolCfg struct {
	InitialCap int
	MaxCap     int
	Factory    ConnectionFactory
	//PrePoolInitializers  []Before
	//PostPoolInitializers []After
	ConnectionConfigs []ConnCfg
	DebugMode         bool
}

// A session spans across many active connections
// Exposes `ActiveConnections` channel that allows connections to be gotten.
type Pool struct {
	// Pool identifier if multiple pools are needed.
	Id string
	// Mutex for getting active connection & size
	Mu sync.Mutex
	// channel storage for our net.Conn connections
	ActiveConnections chan net.Conn
	PoolCfg           PoolCfg
}

// An ordered pool that has priority associated with each connection.
// In effect all connections may not be considered equal TODO: Not implemented yet!
type PriorityPool struct {
	Pool
}

// A `io.Closer` type contract that allows pooled connections
// to be opened, closed, passivated at will.
type Pooled interface {
	// Gets the connection or errors out if none is available.
	Get(cfg ConnCfg) (net.Conn, error)
	// Places the connection into pool
	Put(conn net.Conn) error
	// Provides size of the pool i.e. # of connections.
	Size() int
	// TODO: Investigate io.Closer to be added instead - need to test client code.
	// TODO: There's some complications to client if we do that.
	Close() error
}

// Bootstraps a new pool with pool configuration provided.
// This in effect is the entry point for users of the library.
func New(cfg PoolCfg) (error, *Pool) {
	if cfg.InitialCap < 0 || cfg.MaxCap <= 0 || cfg.InitialCap > cfg.MaxCap {
		return errors.New(fmt.Sprintf("Invalid capacity setting provided : %+v\n", cfg)), nil
	}
	c := &Pool{
		ActiveConnections: make(chan net.Conn, cfg.MaxCap),
		PoolCfg:           cfg,
	}
	// create initial connections, if something goes wrong,
	// just close the pool error out.
	// Note that connection configuration is used without giving order or priority
	// An alternative way could be to use connection pools with each connection - and an associated
	// Keep-alive priority.
	for i := 0; i < cfg.InitialCap; i++ {
		if cfg.DebugMode {
			log.Println("[New]: initializing connections...")
		}
		conn, err := cfg.Factory(c.PoolCfg.ConnectionConfigs[i])
		if cfg.DebugMode {
			// FIXME: leak potentially since it passes a copy? Needed only for time being
			// FIXME: Perhaps
			log.Printf("[New]: initialized connection = %+v\n", &conn)
		}
		if err != nil {
			return fmt.Errorf("factory is not able to fill the pool: %s", err), nil
		}
		c.ActiveConnections <- conn
	}
	return nil, c
}

// Pooled connection `PoolConn` Close() puts the given connects
// back to the pool instead of closing it.
func (p *PoolConn) Close() error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.unusable {
		if p.Conn != nil {
			return p.Conn.Close()
		}
		return nil
	}
	return nil
}

// MarkUnusable() marks the connection not usable any more, to let the pool close it instead of returning it to pool.
func (p *PoolConn) MarkUnusable() {
	p.mu.Lock()
	p.unusable = true
	p.mu.Unlock()
}

// newConn wraps a standard net.Conn to a poolConn net.Conn.
func (c *Pool) wrapConn(conn net.Conn) net.Conn {
	p := &PoolConn{pool: c}
	p.Conn = conn
	return p
}

// Gets the active connections channel safely using `sync.Mutex` Lock() and Unlock() semantics.
func (p *Pool) getSafeActiveConnections() chan net.Conn {
	p.Mu.Lock()
	activeConnections := p.ActiveConnections
	p.Mu.Unlock()
	return activeConnections
}

// Gets the connection from the pool in a safe way.
// The connection so retrieved is an enriched version of connections that could use Close() effectively.
func (p *Pool) Get(cfg ConnCfg) (net.Conn, error) {
	activeConnections := p.getSafeActiveConnections()
	if activeConnections == nil {
		return nil, errors.New("no connections available!")
	}
	if activeConnections == nil {
		return nil, errors.New("no connections available!")
	}
	select {
	case conn := <-p.ActiveConnections:
		if conn == nil {
			return nil, errors.New("no connections available!")
		}
		return p.wrapConn(conn), nil
	default:
		conn, err := p.PoolCfg.Factory(cfg)
		if err != nil {
			return nil, err
		}
		return p.wrapConn(conn), nil
	}
}

// put puts the connection back to the pool. If the pool is full or closed,
// conn is simply closed. A nil conn will be rejected.
func (c *Pool) put(conn net.Conn) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}
	c.Mu.Lock()
	defer c.Mu.Unlock()

	if c.ActiveConnections == nil {
		// pool is closed, close passed connection
		return conn.Close()
	}
	// put the resource back into the pool. If the pool is full, this will
	// block and the default case will be executed.
	select {
	case c.ActiveConnections <- conn:
		return nil
	default:
		// pool is full, close passed connection
		return conn.Close()
	}
}

// Closes the pool - systematically closing all connections.
func (p *Pool) Close() {
	p.Mu.Lock()
	activeConnections := p.ActiveConnections
	p.ActiveConnections = nil
	p.PoolCfg.Factory = nil
	p.Mu.Unlock()
	if activeConnections == nil {
		return
	}
	close(activeConnections)
	for conn := range activeConnections {
		conn.Close()
	}
}

// Gives size of pool - as in number of active connections pool holds at any time.
// Note that this operation locks/unlocks to get active connections to avoid any race conditions
// for client applications.
func (p *Pool) Size() int {
	return len(p.getSafeActiveConnections())
}
