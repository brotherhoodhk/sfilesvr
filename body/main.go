package body

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func AcceptFile(w http.ResponseWriter, r *http.Request) {
	upgrade.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		errorlog.Println(err)
	}
	con := &Connection{con: ws, send: make(chan []byte)}
	hub.register <- con
	defer func() {
		hub.unregister <- con
		ws.Close()
	}()
	func() {
		var msg SendMsg
		var buff = make([]byte, wrbuffsize)
		statusresp := &Response{}
		for {
			statusresp.Content = nil
			statusresp.Footer = ""
			err := ws.ReadJSON(&msg)
			if err != nil {
				//如果解析不出来，则为数据受损，向客户端发送状态码500，跳过本回合
				statusresp.StatusCode = 500
				ws.WriteJSON(statusresp)
				goto passthroug
			} else {
				if msg.Action != 2 {
					statusresp.StatusCode = 200
					ws.WriteJSON(statusresp)
				}
			}
			switch msg.Action {
			case 1:
				filemap := ParseList(filemappath)
				if _, ok := filemap[msg.MessBox]; !ok {
					//如果filesystem中没有对应file，就新建联系
					rand.Seed(time.Now().UnixNano())
					fileid := rand.Intn(899999) + 100000
					filemap[msg.MessBox] = strconv.Itoa(fileid)
					FormatList(filemap, filemappath)
				}
				SaveFile(msg.Content, filemap[msg.MessBox], "")
			case 2:
				filemap := ParseList(filemappath)
				version := -1
				if strings.ContainsRune(msg.MessBox, '@') {
					itres := strings.Split(msg.MessBox, "@")
					version, err = strconv.Atoi(itres[len(itres)-1])
					if err != nil {
						//version不符合规范
						version = -1
						processlog.Println(err)
					}
					msg.MessBox = strings.Join(itres[:len(itres)-1], "")
				}
				if _, ok := filemap[msg.MessBox]; !ok {
					statusresp.StatusCode = 400
					ws.WriteJSON(statusresp)
					goto passthroug
				}
				fileid := filemap[msg.MessBox]
				filepath, ok := getfilepath(fileid, version)
				if !ok {
					statusresp.StatusCode = 400
					ws.WriteJSON(statusresp)
					goto passthroug
				}
				f, err := os.OpenFile(filepath, os.O_RDONLY, 0666)
				if err != nil {
					statusresp.StatusCode = 400
					ws.WriteJSON(statusresp)
					goto passthroug
				}
				lang, err := f.Read(buff)
				if err != nil {
					statusresp.StatusCode = 400
					ws.WriteJSON(statusresp)
					goto passthroug
				}
				statusresp.StatusCode = 200
				statusresp.Content = buff[:lang]
				statusresp.Footer = msg.MessBox
				ws.WriteJSON(statusresp)
			}
		passthroug:
		}
	}()
}