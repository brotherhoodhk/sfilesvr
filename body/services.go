package body

import (
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var publicdir = DATADIR + "/public/"
var privatedir = DATADIR + "/private/"
var privatemapdir = CONFDIR + "/privatemap/"

// save file to data dir,ispublic is true,it will put into public dir,or it will put into private
func SaveFile(content []byte, fileidpath string, privatedirname string) {
	if len(fileidpath) == 6 {
		//this is public zone
		_, err := os.Stat(publicdir + fileidpath)
		if err != nil {
			err = os.Mkdir(publicdir+fileidpath, 0755)
			if err != nil {
				errorlog.Println(err)
				return
			}
		}
		de, _ := ioutil.ReadDir(publicdir + fileidpath)
		filename := publicdir + fileidpath + "/" + strconv.Itoa(len(de)+1) + ".sf"
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0700)
		if err != nil {
			errorlog.Println(err)
			return
		}
		_, err = f.Write(content)
		if err != nil {
			errorlog.Println(err)
			return
		}
	} else if len(fileidpath) == 7 {
		//this is private zone
		rootdir := fileidpath[:3]
		child := fileidpath[3:]
		if _, err := os.Stat(privatedir + rootdir); err != nil {
			err = os.Mkdir(privatedir+rootdir, 0744)
			if err != nil {
				errorlog.Println(err)
				return
			}
		}
		if _, err := os.Stat(privatedir + rootdir + "/" + child); err != nil {
			err = os.Mkdir(privatedir+rootdir+"/"+child, 0744)
			if err != nil {
				errorlog.Println(err)
				return
			}
		}
		de, _ := ioutil.ReadDir(privatedir + rootdir + "/" + child)
		filename := privatedir + rootdir + "/" + child + "/" + strconv.Itoa(len(de)+1) + ".sf"
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0740)
		if err != nil {
			errorlog.Println(err)
			return
		}
		defer f.Close()
		_, err = f.Write(content)
		if err != nil {
			errorlog.Println(err)
			return
		}
	}
}

// 在private 区新建子目录
func mkdirinprivate(dirname string) bool {
	if strings.ContainsRune(dirname, '/') {
		return false
	}
	filelist := ParseList(filemappath)
	if _, ok := filelist[dirname]; ok {
		return true
	}
remakeid:
	rand.Seed(time.Now().UnixNano())
	dirid := rand.Intn(899) + 100
	dirpath := privatedir + strconv.Itoa(dirid)
	_, err := os.Stat(dirpath)
	if err == nil {
		//已经有这个id的目录，则重新生成目录id
		goto remakeid
	}
	err = os.Mkdir(dirpath, 0744)
	if err != nil {
		//无法建立目录，返回失败
		errorlog.Println(err)
		return false
	}
	filelist[dirname] = strconv.Itoa(dirid)
	FormatList(filelist, filemappath)
	return true
}
func saveprivatefile(dirid, filename string, content []byte) bool {
	originpath := privatedir + dirid
	_, err := os.Stat(originpath)
	if err != nil {
		err = os.Mkdir(originpath, 0744)
		if err != nil {
			errorlog.Println(err)
			return false
		}
	}
	_, err = os.OpenFile(privatemapdir+dirid, os.O_CREATE|os.O_RDONLY, 0740)
	filelist := ParseList(privatemapdir + dirid)
	if _, ok := filelist[filename]; !ok {
		fileid := rand.Intn(8999) + 1000
		filelist[filename] = strconv.Itoa(fileid)
		FormatList(filelist, privatemapdir+dirid)
	}
	filelist = ParseList(privatemapdir + dirid)
	SaveFile(content, dirid+filelist[filename], "")
	return true
}

// 能够设立文件权限
func saveprivatefileplus(dirid, filename, usrname string, content []byte, pmsnumber string) bool {
	if !checkpmsnum(pmsnumber) {
		processlog.Println(pmsnumber, " is not pmsnumber")
		return false
	}
	//检查用户对目录操作的权限
	rwxarr, oke := CheckPmsForFileInterface(usrname, "", dirid)
	if !oke || !rwxarr[2] || !rwxarr[1] {
		processlog.Println(usrname, " permission denied")
		return false
	}
	originpath := privatedir + dirid
	_, err := os.Stat(originpath)
	if err != nil {
		err = os.Mkdir(originpath, 0744)
		if err != nil {
			errorlog.Println(err)
			return false
		}
	}
	_, err = os.OpenFile(privatemapdir+dirid, os.O_CREATE|os.O_RDONLY, 0740)
	_, err = os.OpenFile(privatemapdir+dirid+".pmsfs", os.O_CREATE|os.O_RDONLY, 0740)
	filelist := ParseList(privatemapdir + dirid)
	if _, ok := filelist[filename]; !ok {
		fileid := rand.Intn(8999) + 1000
		filelist[filename] = strconv.Itoa(fileid)
		FormatList(filelist, privatemapdir+dirid)
	}
	filelist = ParseList(privatemapdir + dirid)
	SaveFile(content, dirid+filelist[filename], "")
	AddToFPMS(filelist[filename], usrname, "740", privatemapdir+dirid+".pmsfs")
	return true
}
func deletefilefromprivate(heads string) bool {
	namearr := strings.Split(heads, "/")
	filelist := ParseList(filemappath)
	if _, ok := filelist[namearr[0]]; !ok {
		return true
	}
	firid := filelist[namearr[0]]
	filelist = ParseList(privatemapdir + firid)
	if _, ok := filelist[namearr[1]]; !ok {
		return true
	}
	secid := filelist[namearr[1]]
	completeid := firid + secid
	if deletefile(completeid) {
		//删除对应目录对应文件联系
		delete(filelist, namearr[1])
		FormatList(filelist, privatemapdir+firid)
		return true
	} else {
		return false
	}
}

// 需要验证权限
func deletefilefromprivateplus(heads, usrname string, usrinfo *UserInfo) (int, bool) {
	namearr := strings.Split(heads, "/")
	filelist := ParseList(filemappath)
	if _, ok := filelist[namearr[0]]; !ok {
		return 200, true
	}
	firid := filelist[namearr[0]]
	filelist = ParseList(privatemapdir + firid)
	if _, ok := filelist[namearr[1]]; !ok {
		return 200, true
	}
	rwxpms := CheckPmsForFile(usrname, usrinfo, "", firid)
	if !rwxpms[2] {
		processlog.Println(usrname, "dont have execute permission on", namearr[0])
		return 402, false
	}
	secid := filelist[namearr[1]]
	rwxpms = CheckPmsForFile(usrname, usrinfo, secid, firid)
	if !rwxpms[2] {
		processlog.Println(usrname, "dont have execute permission on", namearr[1])
		return 402, false
	}
	completeid := firid + secid
	if deletefile(completeid) {
		//删除对应目录对应文件联系
		delete(filelist, namearr[1])
		FormatList(filelist, privatemapdir+firid)
		firidpath := buildpmspath(firid)
		filelist = ParseList(firidpath)
		delete(filelist, secid)
		FormatList(filelist, firidpath)
		return 200, true
	} else {
		return 400, false
	}
}
