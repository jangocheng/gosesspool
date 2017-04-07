//
// @(#)TCPSessions.go  Date: 3/23/17
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
	"crypto/sha512"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// A session is logical representation
// of what user is doing with peer that it connects to
// which maintains session information on the user.
// This is useful when providing history of commands
// User details updates, session times and historical record.
type Session struct {
	// SHA512 identifier of the session based on connections
	Id string
	// Username provided
	// This is scoped to IP Address provided(RemoteAddr)
	Username string
	// Commands user executes
	CommandLog []string
	// Session Start time
	StartTime string
	// Session Finish Time
	FinishTime string
	Conn       *net.Conn
}

// Sessions provides engines, TCP reactors
// with modeling capability for many sessions.
type Sesspool struct {
	Mu             sync.RWMutex
	ActiveSessions map[string]*Session
}

// Contract that `Session` aims to implement.
type Sessioned interface {
	// Starts a new Session
	New(c *net.Conn) (error, Session)
	// Flushes command log to a/many permanent store(s) using sequence flushingMeadowsClosures provided.
	// NOTE : If command list is 15 - this will flush automatically all commands to a store using closure.
	Flush(flushingMeadows ...func(*Session) error) (error, string)
	// Closes an active session
	Close() (error, string)
}

// Hashes connection using remote address to SHA-512 string or errors out.
func Hash(c *net.Conn) (error, string) {
	h := sha512.New()
	hostPort := (*c).RemoteAddr().String()
	h.Write([]byte(strings.Split(hostPort, ":")[0]))
	return nil, fmt.Sprintf("%x", h.Sum(nil))
}

// Starts a session, scopes to a session.
func NewSession(c *net.Conn) (error, Session) {
	_, h := Hash(c)
	return nil, Session{Id: h, Username: (*c).RemoteAddr().String(), CommandLog: []string{}, StartTime: time.Now().String(), FinishTime: "", Conn: c}
}

// Appends a new session to existing sessions being maintained by a networking node.
func (sessPool *Sesspool) Append(s *Session) error {
	log.Printf("[BEFORE]: lock acquired for session ~~~ %+v\n", s)
	sessPool.Mu.Lock()
	defer sessPool.Mu.Unlock()
	log.Printf("lock acquired for session ~~~ %+v\n", s)
	if sessPool.ActiveSessions[s.Id] != nil {
		return errors.New("Session Already Exists, use that connection instead!")
	} else {
		sessPool.ActiveSessions[s.Id] = s
	}
	log.Printf("lock released for session ~~~ %+v\n", s)
	return nil
}

// Associate a session with connection
// Need to investigate Yamux more, felt like implementing thyself be easier.
func Associate(c *net.Conn) Session {
	_, h := Hash(c)
	return Session{Id: h, Username: (*c).RemoteAddr().String(), CommandLog: []string{}, StartTime: time.Now().String(), FinishTime: "", Conn: c}
}
