package body

import (
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
		// fmt.Println(err)
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
					goto passthroug
				}
				fileid := filemap[msg.MessBox]
				filepath, ok := getfilepath(fileid, version)
				if !ok {
					statusresp.StatusCode = 400
					goto passthroug
				}
				f, err := os.OpenFile(filepath, os.O_RDONLY, 0666)
				if err != nil {
					statusresp.StatusCode = 400
					goto passthroug
				}
				lang, err := f.Read(buff)
				if err != nil {
					statusresp.StatusCode = 400
					goto passthroug
				}
				statusresp.StatusCode = 200
				statusresp.Content = buff[:lang]
				statusresp.Footer = msg.MessBox
				ws.WriteJSON(statusresp)
			case 41:
				if !strings.ContainsRune(msg.MessBox, '/') || len(msg.MessBox) < 3 {
					statusresp.StatusCode = 401
					goto passthroug
				}
				dirarr := strings.Split(msg.MessBox, "/")
				if len(dirarr) != 2 || len(dirarr[0]) < 1 || len(dirarr[1]) < 1 {
					//不符合 dirname/filename的规范
					statusresp.StatusCode = 401
					goto passthroug
				}
				filemap := ParseList(filemappath)
				if _, ok := filemap[dirarr[0]]; !ok || len(msg.Content) < 1 {
					statusresp.StatusCode = 400
					goto passthroug
				}
				if saveprivatefile(filemap[dirarr[0]], dirarr[1], msg.Content) {
					statusresp.StatusCode = 200
				} else {
					statusresp.StatusCode = 400
				}
			}
		passthroug:
			ws.WriteJSON(statusresp)
		}
	}()
}

// 接受其他指令
func OtherCommand(w http.ResponseWriter, r *http.Request) {
	upgrade.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		// fmt.Println(err)
		errorlog.Println(err)
	}
	con := &Connection{con: ws, send: make(chan []byte)}
	hub.register <- con
	defer func() {
		hub.unregister <- con
		ws.Close()
	}()
	func() {
		var cmd CommonCommand
		var resp = new(Response)
		for {
			resp.Footer = ""
			resp.Content = nil
			err := ws.ReadJSON(&cmd)
			if err != nil {
				resp.StatusCode = 500
				goto passthrough
			}
			switch cmd.Actionid {
			case 3:
				//删除指定文件
				if len(cmd.Header) == 0 {
					resp.StatusCode = 400
					goto passthrough
				}
				filemap := ParseList(filemappath)
				if _, ok := filemap[cmd.Header]; !ok {
					resp.StatusCode = 200
					goto passthrough
				}
				if deletefile(filemap[cmd.Header]) {
					delete(filemap, cmd.Header)
					FormatList(filemap, filemappath)
					resp.StatusCode = 200
				} else {
					//delete file failed
					resp.StatusCode = 400
				}
				ws.WriteJSON(resp)
			case 40:
				//在private 目录下新建子目录
				if len(cmd.Header) == 0 {
					resp.StatusCode = 400
					goto passthrough
				}
				if mkdirinprivate(cmd.Header) {
					resp.StatusCode = 200
					goto passthrough
				} else {
					resp.StatusCode = 400
					goto passthrough
				}
			case 42:
				//获取指定private file
				if !isprivatefilename(cmd.Header) {
					resp.StatusCode = 401
					goto passthrough
				}
				namearr := strings.Split(cmd.Header, "/")
				name, ver := GetVersion(namearr[1])
				namearr[1] = name
				cmd.Header = strings.Join(namearr, "/")
				if fileid, ok := isexistprivatefile(cmd.Header); ok {
					content, okk := getfilecontent(fileid, ver)
					if !okk {
						resp.StatusCode = 400
						goto passthrough
					}
					resp.StatusCode = 200
					resp.Content = content
					resp.Footer = namearr[1]
					goto passthrough
				} else {
					resp.StatusCode = 401
					goto passthrough
				}
			case 43:
				//删除private区指定目录
				if len(cmd.Header) == 0 {
					resp.StatusCode = 400
					goto passthrough
				}
				filelist := ParseList(filemappath)
				if dirid, ok := filelist[cmd.Header]; ok && len(dirid) == 3 {
					//目录名存在
					_, err := os.Stat(privatedir + dirid)
					if err == nil {
						err = os.RemoveAll(privatedir + dirid)
						if err != nil {
							errorlog.Println(err)
							resp.StatusCode = 400
							goto passthrough
						}
						err = os.Remove(privatemapdir + dirid)
						if err != nil {
							errorlog.Println(err)
							resp.StatusCode = 400
							goto passthrough
						}
					}
					delete(filelist, cmd.Header)
					FormatList(filelist, filemappath)
					resp.StatusCode = 200
					goto passthrough
				}
			case 431:
				if len(cmd.Header) == 0 {
					resp.StatusCode = 400
					goto passthrough
				}
				if !isprivatefilename(cmd.Header) {
					resp.StatusCode = 401
					goto passthrough
				}
				if deletefilefromprivate(cmd.Header) {
					resp.StatusCode = 200
				} else {
					resp.StatusCode = 400
				}
				goto passthrough
			case 900:
				//test function
				if AuthServe(&cmd.Auth) {
					resp.StatusCode = 200
				} else {
					resp.StatusCode = 400
				}
				goto passthrough
			case 840:
				//认证版新建目录
				if AuthServe(&cmd.Auth) && len(cmd.Header) > 0 {
					if mkdirinprivate(cmd.Header) {
						filelist := ParseList(filemappath)
						fileid := filelist[cmd.Header]
						oke := AddToFPMS(fileid, string(cmd.Auth.Usrname), "760", FILEPMS)
						if !oke {
							resp.StatusCode = 400
						} else {
							resp.StatusCode = 200
						}
					}
				} else {
					resp.StatusCode = 400
				}
				goto passthrough
			case 842:
				//拉取私立目录中的指定文件(认证版)
				usrinfo, ok := GetUserInfo(string(cmd.Auth.Usrname))
				if !AuthServe(&cmd.Auth) || !ok {
					resp.StatusCode = 400
					goto passthrough
				}
				if !isprivatefilename(cmd.Header) {
					resp.StatusCode = 401
					goto passthrough
				}
				namearr := strings.Split(cmd.Header, "/")
				name, ver := GetVersion(namearr[1])
				namearr[1] = name
				cmd.Header = strings.Join(namearr, "/")
				if fileid, ok := isexistprivatefile(cmd.Header); ok {
					rwxpms := CheckPmsForFile(string(cmd.Auth.Usrname), usrinfo, fileid[3:], fileid[:3])
					if !rwxpms[0] {
						resp.StatusCode = 402
						processlog.Println(string(cmd.Auth.Usrname), "dont have", cmd.Header, "'s read permission")
						goto passthrough
					}
					content, okk := getfilecontent(fileid, ver)
					if !okk {
						resp.StatusCode = 400
						processlog.Println(cmd.Header, "get file failed")
						goto passthrough
					}
					resp.StatusCode = 200
					resp.Content = content
					resp.Footer = namearr[1]
					goto passthrough
				} else {
					resp.StatusCode = 401
					errorlog.Println(cmd.Header, "is not correct private filename")
					goto passthrough
				}
			case 843:
				//删除指定目录(认证版)
				usrinfo, oke := GetUserInfo(string(cmd.Auth.Usrname))
				if !AuthServe(&cmd.Auth) || !oke {
					resp.StatusCode = 400
					goto passthrough
				}
				if len(cmd.Header) == 0 {
					resp.StatusCode = 400
					goto passthrough
				}
				filelist := ParseList(filemappath)
				if dirid, ok := filelist[cmd.Header]; ok && len(dirid) == 3 {
					//目录名存在
					rwxpms := CheckPmsForFile(string(cmd.Auth.Usrname), usrinfo, "", dirid)
					if !rwxpms[2] {
						resp.StatusCode = 402
						processlog.Println(string(cmd.Auth.Usrname), "dont have execute permission on", cmd.Header)
						goto passthrough
					}
					_, err := os.Stat(privatedir + dirid)
					if err == nil {
						err = os.RemoveAll(privatedir + dirid)
						if err != nil {
							errorlog.Println(err)
							resp.StatusCode = 400
							goto passthrough
						}
						err = os.Remove(privatemapdir + dirid)
						if err != nil {
							errorlog.Println(err)
							resp.StatusCode = 400
							goto passthrough
						}
					}
					pmsfspath := buildpmspath(dirid)
					_, err = os.Stat(pmsfspath)
					if err == nil {
						err = os.RemoveAll(pmsfspath)
						if err != nil {
							errorlog.Println(err)
							resp.StatusCode = 400
							goto passthrough
						}
					}
					delete(filelist, cmd.Header)
					FormatList(filelist, filemappath)
					filelist = ParseList(FILEPMS)
					delete(filelist, dirid)
					FormatList(filelist, FILEPMS)
					resp.StatusCode = 200
					goto passthrough
				} else {
					resp.StatusCode = 400
					goto passthrough
				}
			case 8431:
				usrinfo, oke := GetUserInfo(string(cmd.Auth.Usrname))
				if !AuthServe(&cmd.Auth) && !oke {
					resp.StatusCode = 400
					goto passthrough
				}
				if len(cmd.Header) == 0 {
					resp.StatusCode = 400
					goto passthrough
				}
				if !isprivatefilename(cmd.Header) {
					resp.StatusCode = 401
					goto passthrough
				}
				if statuscode, ok := deletefilefromprivateplus(cmd.Header, string(cmd.Auth.Usrname), usrinfo); ok {
					resp.StatusCode = 200
				} else {
					resp.StatusCode = statuscode
				}
				goto passthrough
			}
		passthrough:
			ws.WriteJSON(resp)
		}
	}()
}

// add auth
func AcceptFilePlus(w http.ResponseWriter, r *http.Request) {
	upgrade.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		// fmt.Println(err)
		errorlog.Println(err)
	}
	con := &Connection{con: ws, send: make(chan []byte)}
	hub.register <- con
	defer func() {
		hub.unregister <- con
		ws.Close()
	}()
	func() {
		var msg SendMsgPlus
		statusresp := &Response{}
		for {
			statusresp.Content = nil
			statusresp.Footer = ""
			err := ws.ReadJSON(&msg)
			if err != nil {
				//如果解析不出来，则为数据受损，向客户端发送状态码500，跳过本回合
				statusresp.StatusCode = 500
				goto passthroug
			}
			if !AuthServe(&msg.Auth) {
				statusresp.StatusCode = 400
				goto passthroug
			}
			switch msg.Action {
			case 841:
				//客户端向服务端上传私立目录文件
				if !strings.ContainsRune(msg.MessBox, '/') || len(msg.MessBox) < 3 {
					statusresp.StatusCode = 401
					goto passthroug
				}
				dirarr := strings.Split(msg.MessBox, "/")
				if len(dirarr) != 2 || len(dirarr[0]) < 1 || len(dirarr[1]) < 1 {
					//不符合 dirname/filename的规范
					statusresp.StatusCode = 401
					goto passthroug
				}
				filemap := ParseList(filemappath)
				//debugline
				// processlog.Println(filemap)
				if _, ok := filemap[dirarr[0]]; !ok || len(msg.Content) < 1 {
					processlog.Println(dirarr[0], " dont exist in filemap msg content length ", len(msg.Content))
					statusresp.StatusCode = 400
					goto passthroug
				}
				//检查用户权限是否符合要求
				usfinfo, ok := GetUserInfo(string(msg.Auth.Usrname))
				if !ok {
					processlog.Println("some error occur")
					statusresp.StatusCode = 400
					goto passthroug
				}
				pmsarr := CheckPmsForFile(string(msg.Auth.Usrname), usfinfo, "", filemap[dirarr[0]])
				if !pmsarr[1] {
					statusresp.StatusCode = 402
					processlog.Println(string(msg.Auth.Usrname), "dont have write permission on", msg.MessBox)
					goto passthroug
				}
				if saveprivatefileplus(filemap[dirarr[0]], dirarr[1], string(msg.Auth.Usrname), msg.Content, "740") {
					statusresp.StatusCode = 200
				} else {
					statusresp.StatusCode = 400
				}
			}
		passthroug:
			ws.WriteJSON(statusresp)
		}
	}()
}
