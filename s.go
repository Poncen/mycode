//多个服务器和客户端。客户端向服务器发出不同的请求
//一个中心节点用于审批节点的加入离开和网络的维护
//根据报文头来分类处理不同的消息（handle），用json来序列化结构体
//0.简单字符串[default]
//1.请求加入：id listenport————添加表单并转发（dial）&&把表单整体传递给新节点（datatrans）&&通知节点已加入网络[1]
//		可能的错误：id重复，服务器不存在
//2.请求离开：id listenport————删除表单并转发[2]
//3.来自转发的加入/离开请求：id listenport————仅添加/删除表单[3,4]
//4.连通确认[5]
//5.复杂的高级消息：id num1 num2 time etc.[6,7...]
//另：
//反馈确认函数（check）
//超时处理函数（time）
//心跳检查函数（heartbeating）
//错误检查函数（checkerr）
//建立本地数据库（mysql）

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

//定义节点信息结构体info
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
	Tradeinfo *trade
}

// var i bool
var address, id string
var InfoMap map[string]Info
var CliMap map[string]clidata

//错误检查函数
func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

//连接建立函数
func dial(c []byte) {
	for _, con := range InfoMap { //根据infomap的信息转发
		//返回一个计数值与map的项数对比，判断是否全部传达到？
		if con.Host == address {
			//跳过自身
			continue
		}
		tcpAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:"+con.Host)
		conn, err := net.DialTCP("tcp", nil, tcpAddr) //建立通信
		if err != nil {
			fmt.Println("err", err)
			os.Exit(0)
		}
		conn.Write(c)
		conn.Close()
	}
}

//申请函数，向中心节点发去报文，返回一个bool参数
// func request(data []byte, long int) (i bool) {
// 	tcpAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:10000")
// 	conn, err := net.DialTCP("tcp", nil, tcpAddr)
// 	conn.Write(data)
// 	buff := make([]byte, 128)
// 	for {
// 		j, err := conn.Read(buff)
// 		if err != nil {
// 			ch <- 1
// 			break
// 		}
// 		if string(buff) == "yes" {
// 			i = true
// 			break
// 		}
// 		if string(buff) == "no" {
// 			i = false
// 			break
// 		}
// 	}
// 	conn.Close()
// 	return i
// }

//分类处理函数
func handle(data []byte, long int) {
	v := &message{}
	err := json.Unmarshal(data[:long], v)
	checkErr(err)
	//加入请求
	if v.Kind == "join" {
		// i=request(data,long)
		if v.Verifykey == true {
			InfoMap[v.Jlmessage.Host] = Info{v.Jlmessage.Id, v.Jlmessage.Host}
			v.Verifykey = false
			//改为转发的标识
			b, _ := json.Marshal(v)
			dial(b)
			tcpAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:"+v.Jlmessage.Host)
			conn, err := net.DialTCP("tcp", nil, tcpAddr)
			//建立通信
			checkErr(err)
			v.Kind = "receive"
			v.Peerinfo = InfoMap
			c, _ := json.Marshal(v)
			conn.Write(c)
			conn.Close()
		}
		if v.Verifykey == false {
			InfoMap[v.Jlmessage.Host] = Info{v.Jlmessage.Id, v.Jlmessage.Host}
			fmt.Println(v.Jlmessage.Id + "joined the web and map updated")
		}
	}
	//接受节点拓扑表
	if v.Kind == "receive" {
		b, _ := json.Marshal(v.Peerinfo)
		err := json.Unmarshal(b, &InfoMap)
		fmt.Println("get the map")
		checkErr(err)
	}
	if v.Kind == "leave" {
		// i=request(data,long)
		if v.Verifykey == true {
			delete(InfoMap, v.Jlmessage.Host)
			fmt.Println(v.Jlmessage.Id + "want to leave")
			v.Verifykey = false
			c, _ := json.Marshal(v)
			dial(c)
		}
		if v.Verifykey == false {
			delete(InfoMap, v.Jlmessage.Host)
			fmt.Println(v.Jlmessage.Id + "leaved the net and map updated")
		}
	}
}

// 消息接收函数
func chat(tcpConn *net.TCPConn) {
	//区分出来自中心节点的消息
	ipStr := tcpConn.RemoteAddr().String()
	// 中断处理
	defer func() {
		fmt.Println("disconnected :" + ipStr)
		tcpConn.Close()
	}()

	for {
		data := make([]byte, 10000)
		total, err := tcpConn.Read(data)
		if err != nil {
			break
		}
		if tcpConn.RemoteAddr().String() != "127.0.0.1:10000" {
			handle(data, total)
			e := []byte("finish")
			tcpConn.Write(e)
		}
	}
}

func main() {
	//输入节点信息
	fmt.Println("Input your listen port")
	fmt.Scanln(&address)
	fmt.Println("Input your id")
	fmt.Scanln(&id)
	tcpAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:"+address)
	tcpListener, _ := net.ListenTCP("tcp", tcpAddr)
	//建立自身的map表格并初始化，以id为索引
	InfoMap = map[string]Info{
		address: Info{
			id,
			address,
		},
	}
	//开始监听消息
	for {
		tcpConn, _ := tcpListener.AcceptTCP()
		defer tcpConn.Close()
		fmt.Println("连接的客服端信息:", tcpConn.RemoteAddr().String())
		//消息处理
		go chat(tcpConn)
	}
}
