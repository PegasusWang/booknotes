// 聊天室 demo
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

func main() {
	// tcp  server 监听
	listener, err := net.Listen("tcp", ":2020")
	if err != nil {
		panic(err)
	}

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConn(conn)
	}
}

var (
	enteringChannel = make(chan *User)     // 记录新用户到来
	leavingChannel  = make(chan *User)     // 离开用户通过该 channel 等级
	messageChannel  = make(chan string, 8) // 广播用户普通消息. 8 需要自测调整
)

// 广播消息：1.新用户进来 2.用户普通消息 3.用户离开
func broadcaster() {
	users := make(map[*User]struct{})
	for {
		select {
		case user := <-enteringChannel:
			users[user] = struct{}{}
		case user := <-leavingChannel:
			delete(users, user)
			close(user.MessageChannel) // 避免 goroutine 泄露
		case msg := <-messageChannel:
			// 给所有在线用户发送消息
			for user := range users {
				user.MessageChannel <- msg
			}
		}
	}
}

type User struct {
	ID             int
	Addr           string
	EnterAt        time.Time
	MessageChannel chan string // 当前用户发送消息的通道
}

func (u *User) String() string {
	return fmt.Sprintf("User:[%d]", u.ID)
}

var i int

func GenUserID() int {
	i++
	return i
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	user := &User{
		ID:             GenUserID(),
		Addr:           conn.RemoteAddr().String(),
		EnterAt:        time.Now(),
		MessageChannel: make(chan string, 8),
	}

	// 新开一个 goroutine 给用户发送消息
	go sendMessage(conn, user.MessageChannel)

	// 给当前用户发送欢迎消息，然后向所有用户告知新用户到来
	user.MessageChannel <- "Welcome, " + user.String()
	messageChannel <- "user:" + strconv.Itoa(user.ID) + " has enter"

	// 记录到全局用户列表
	enteringChannel <- user

	// 循环读取用户输入
	input := bufio.NewScanner(conn)
	for input.Scan() {
		messageChannel <- strconv.Itoa(user.ID) + ": " + input.Text()
	}

	if err := input.Err(); err != nil {
		log.Println("读取错误:", err)
	}

	// 用户离开
	leavingChannel <- user
	messageChannel <- "user:" + strconv.Itoa(user.ID) + " has left"
}

func sendMessage(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}
