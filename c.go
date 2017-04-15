package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type Info struct {
	Id   string
	Host string
}

type trade struct {
	From     string
	To       string
	Quantity int
	Sort     string
}

type clidata struct {
	Host   string
	First  int
	Second int
}

type message struct {
	Kind      string
	Verifykey bool
	Jlmessage *Info
	Peerinfo  map[string]Info
	Tradeinfo trade `json:"body,omitempty"`
}

//定义通道
var ch chan int = make(chan int)

//定义ID
var id, address string

func reader(conn *net.TCPConn) {
	buff := make([]byte, 128)
	for {
		j, err := conn.Read(buff)
		if err != nil {
			ch <- 1
			break
		}

		fmt.Println(string(buff[0:j]))
	}
}

func input() (id string, address string) {
	fmt.Println("Input your id")
	fmt.Scanln(&id)
	fmt.Println("Input your listenport")
	fmt.Scanln(&address)
	fmt.Println("目标port为:", address, "目标id为", id)
	return id, address
}

//main
func main() {
	var key bool
	key = true
	fmt.Println("Input your server's listen port")
	fmt.Scanln(&address)
	tcpAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:"+address)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	// var port string
	// port = "127.0.0.1:9999"
	if err != nil {
		fmt.Println("Server is not starting")
		os.Exit(0)
	}

	defer conn.Close()
	go reader(conn)

	for {
		var msg string
		fmt.Scanln(&msg)
		//输入数据，仅设计加入离开即可，复杂的消息由单独client完成

		//if msg==join
		//if msg==leave
		// b := []byte(b)
		if msg == "join" { //以后会提取出独立的函数
			id, address = input()
			a := &Info{
				id,
				address,
			}
			b := &message{
				Kind:      "join",
				Verifykey: key,
				Jlmessage: a,
			}
			c, err := json.Marshal(b)
			if err != nil {
				fmt.Println("err:", err)
			}
			conn.Write(c)
		}
		if msg == "leave" {
			id, address = input()
			a := &Info{
				id,
				address,
			}
			b := message{
				Kind:      "leave",
				Verifykey: key,
				Jlmessage: a,
			}
			c, err := json.Marshal(b)
			if err != nil {
				fmt.Println("err:", err)
			}
			conn.Write(c)
		}
		if msg == "quit" {
			break
		}
		//select 为非阻塞的
		select {
		case <-ch:
			fmt.Println("Server错误!请重新连接!")
			os.Exit(1)
		default:

		}

	}

}
