《Network Programming With Go》不是很深入。


# 3. Socket level Programming

### socket
```
// resolve IP
package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	name := "www.baidu.com"
	addr, err := net.ResolveIPAddr("ip", name)
	if err != nil {
		fmt.Println("Resolution error", err.Error())
		os.Exit(1)
	}
	fmt.Println("addr ip is ", addr.String())
}
```

编写一个时间回显tcp服务器：

```
// DaytimeServer
// telnet localhost 1200
package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	service := ":1200"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		daytime := time.Now().String()
		conn.Write([]byte(daytime))
		conn.Close()
	}
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

编写一个简单的tcp回显服务器：
```
```
// SimpleEchoServer
package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	service := ":1201"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		handleClient(conn)
		conn.Close() // close  the client
	}
}
func handleClient(conn net.Conn) {
	var buf [512]byte
	for {
		n, err := conn.Read(buf[0:])
		if err != nil {
			return
		}
		fmt.Println(string(buf[0:]))
		_, err2 := conn.Write(buf[0:n])
		if err2 != nil {
			return
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
```

可以很容易修改成并发的，使用 goroutine

```
// SimpleEchoServer
package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	service := ":1201"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn) // NOTE: use goroutine
	}
}
func handleClient(conn net.Conn) {
	defer conn.Close() // NOTE: close connection on exit
	var buf [512]byte
	for {
		n, err := conn.Read(buf[0:])
		if err != nil {
			return
		}
		fmt.Println(string(buf[0:]))
		_, err2 := conn.Write(buf[0:n])
		if err2 != nil {
			return
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
```

### Controlling TCP connections

```go
// Timeout
func (c *TCPConn) SetTimeout(nsec int64) os.Error

// Staying alive, a client wish to stay connected to a server even if it has nothing to send
func (c *TCPConn) SetKeepAlive(keepalive bool) os.Error
```

UDP client server demo:

```go
// UDPDaytimeClient
package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	service := "localhost:1200"
	udpAddr, err := net.ResolveUDPAddr("udp4", service)

	conn, err := net.DialUDP("udp", nil, udpAddr)
	checkError(err)

	_, err = conn.Write([]byte("anything"))
	checkError(err)
	var buf [512]byte
	n, err := conn.Read(buf[0:])
	checkError(err)

	fmt.Println(string(buf[0:n]))
}
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

// udpserver.go
package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	service := ":1200"
	updAddr, err := net.ResolveUDPAddr("udp4", service)
	checkError(err)
	conn, err := net.ListenUDP("udp", updAddr)
	checkError(err)
	for {
		handleClient(conn)
	}
}

func handleClient(conn *net.UDPConn) {
	var buf [512]byte
	_, addr, err := conn.ReadFromUDP(buf[0:])
	if err != nil {
		return
	}
	daytime := time.Now().String()
	conn.WriteToUDP([]byte(daytime), addr)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
```

### Raw sockets and the type IPConn

```go
// Ping
package main

import (
	"fmt"
	"net"
	"os"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "host")
		os.Exit(1)
	}
	addr, err := net.ResolveIPAddr("ip", os.Args[1])
	if err != nil {
		fmt.Println("Resolution error", err.Error())
		os.Exit(1)
	}

	fmt.Println(addr)
	conn, err := net.DialIP("ip4:icmp", addr, addr)
	checkError(err)

	var msg [512]byte
	msg[0] = 8  //echo
	msg[1] = 0  //code 0
	msg[2] = 0  //checksum, fix later
	msg[3] = 0  //checksum, fix later
	msg[4] = 0  //identifier[0]
	msg[5] = 13 //identifier[1]
	msg[6] = 0  // sequence[0]
	msg[7] = 37 //sequence[1]
	len := 8

	check := checkSum(msg[0:len])
	msg[2] = byte(check >> 8)
	msg[3] = byte(check & 255)

	_, err = conn.Write(msg[0:len])
	checkError(err)

	_, err = conn.Read(msg[0:])
	checkError(err)

	fmt.Println("Got response")
	if msg[5] == 13 {
		fmt.Println("identifier matches")
	}
	if msg[7] == 37 {
		fmt.Println("Sequence matches")
	}
}

func checkSum(msg []byte) uint16 {
	sum := 0
	for n := 1; n < len(msg)-1; n += 2 {
		sum += int(msg[n])*256 + int(msg[n+1])
	}
	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)
	answer := uint16(^sum)
	return answer
}
```


# 4. Data serialisation

- ASN.1
- JSON
- gob
- base64


# 5. Application Level Protocols

- version control
- message format
- data format: byte encoded or character encoded
- state


# 6. Managing character sets and encodings

Go use utf8 encoded characters in its strings. Each character is of type rune(alias for int32)
as a Unicode character can be 1,2 or 4 bytes in UTF8 encoding. A string is an array of rune.


# 7. Security

### Data integrity (数据完整性)

```
// MD5 Has
package main

import (
	"crypto/md5"
	"fmt"
)

func main() {
	hash := md5.New()
	bytes := []byte("hello\n")
	hash.Write(bytes)
	hashValue := hash.Sum(nil)
	hashSize := hash.Size()
	// print out in ASCII from as four hexadcimal numbers
	for n := 0; n < hashSize; n += 4 {
		var val uint32
		val = uint32(hashValue[n])<<24 +
			uint32(hashValue[n+1])<<16 +
			uint32(hashValue[n+2])<<8 +
			uint32(hashValue[n+3])
		fmt.Printf("%x ", val)
	}
}
```

### Symmetric key encryption

- Blowfish
- DES

### Public key encryption

- crypto/rsa

### X.509 certificates

A public key infrastructure(PKI) is a framework for a collections of public keys,
along with additional information such as owner name and location.

### TLS

- crypto/tls


# 8. HTTP

# HTTP client
```
package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	url := "http://127.0.0.1:8000"
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
	if response.Status != "200 OK" {
		fmt.Println(response.Status)
		os.Exit(2)
	}
	// b, _ := httputil.DumpResponse(response, false)
	// fmt.Print(string(b))

	var buf [512]byte
	reader := response.Body
	for {
		n, err := reader.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		fmt.Print(string(buf[0:n]))
	}

}
```

Proxy handling:

```
proxyURL, err := url.Parse(proxyString) // http://proxy-host:port
transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
client := &http.Client{Transport: transport}
// func ProxyFromEnvironment(req *Request) (*url.URL, error)
```

### Servers

- File server

```
package main

import "net/http"

func main() {
	fileServer := http.FileServer(http.Dir("/home/httpd/html"))
	err := http.ListenAndServe(":8000", fileServer)
	checkError(err)
}
```

- Handler function

```
func Handle(pattern string, handler Handler)
func HandleFunc(pattern string, handler func(*Conn, *Request))
```


# 9. Templates

- html/template, text/template
- pipelines: `{{. | html}}`
- `template.FuncMap{"emailExpand": EmailExpand}`
- variables in templates are prefixed by '$'
- conditional
