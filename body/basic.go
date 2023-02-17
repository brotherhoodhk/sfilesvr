package body

import (
	"fmt"
	"log"
	"os"

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
type Response struct {
	StatusCode int
	Content    []byte
	Footer     string
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
