package body

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

var ROOTPATH = os.Getenv("SFILESVR_HOME")
var LOGPATH = ROOTPATH + "/logs/"
var CONFDIR = ROOTPATH + "/conf"
var DATADIR = ROOTPATH + "/data"
var errorlog = LogInit("error")
var filemappath = CONFDIR + "/filemap"

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

var connecntionpool = make(map[*Connection]struct{})
var wrbuffsize = 100 * MB
var upgrade = websocket.Upgrader{ReadBufferSize: wrbuffsize, WriteBufferSize: wrbuffsize}
var hub = Hub{register: make(chan *Connection), unregister: make(chan *Connection), broadcast: make(chan []byte)}
var processlog = LogInit("process") //记录过程记录，非报错，类似于警告
type Connection struct {
	con  *websocket.Conn
	send chan []byte
}
type Hub struct {
	register   chan *Connection
	unregister chan *Connection
	broadcast  chan []byte
}

// 客户端发向服务端格式
type SendMsg struct {
	Content []byte `json:"content"`
	Action  int    `json:"action"`
	MessBox string `json:"messbox"`
}

// 客户端发向服务端格式(加强版)
type SendMsgPlus struct {
	Content []byte     `json:"content"`
	Action  int        `json:"action"`
	MessBox string     `json:"messbox"`
	Auth    AuthMethod `json:"auth"`
}
type Response struct {
	StatusCode int
	Content    []byte
	Footer     string
}

// 通用指令协议
type CommonCommand struct {
	Header   string
	Cmd      map[string]string
	Actionid int
	//version 2.1
	Auth AuthMethod
}
type AuthMethod struct {
	Key     []byte
	Usrname []byte
}

// user information
type UserInfo struct {
	Authkey   string
	Writeable [3]bool
	Readable  [3]bool
	Execable  [3]bool
}

// file permission info
type FilePmsInfo struct {
	Owner     string
	Group     string
	Writeable [3]bool
	Readable  [3]bool
	Execable  [3]bool
}

func (s *Hub) Run() {
	for {
		select {
		case c := <-s.register:
			connecntionpool[c] = struct{}{}
		case c := <-s.unregister:
			if _, ok := connecntionpool[c]; ok {
				delete(connecntionpool, c)
			}
		case <-s.broadcast:
			for client, _ := range connecntionpool {
				select {
				case <-client.send:
				default:
					delete(connecntionpool, client)
				}
			}
		}
	}
}
func LogInit(logname string) *log.Logger {
	f, err := os.OpenFile(LOGPATH+logname+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("open file failed,", err)
		return nil
	}
	newlog := log.New(f, "["+logname+"]", log.LUTC|log.Lshortfile|log.LstdFlags)
	return newlog
}
func Exist_File(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}
func GetVersion(filename string) (string, int) {
	if !strings.ContainsRune(filename, '@') {
		return filename, -1
	}
	namearr := strings.Split(filename, "@")
	if len(namearr) < 2 && (len(namearr[0]) < 1 || len(namearr[1]) < 1) {
		return filename, -1
	}
	ver, err := strconv.Atoi(namearr[len(namearr)-1])
	if err != nil {
		return filename, -1
	}
	return strings.Join(namearr[:len(namearr)-1], ""), ver
}
func (s *UserInfo) String() string {
	res := s.Authkey
	return res
}
