//
// @(#)TCPSessions_test.go  Date: 3/23/17
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
	"net"
	"testing"
)

func TestHash(t *testing.T) {
	c, _ := net.Dial("tcp", "127.0.0.1:12345")
	fmt.Println(Hash(&c))
	// prints 02080828193f034f7d90f9f4b559c9a8d2f3be806a71f58f8b5948f9cfb5bc7b35c900bdb55a006f152ccde6502c12853ff2786a176b2c9b115d09baad531d82
}
