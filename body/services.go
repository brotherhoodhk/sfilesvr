package body

import (
	"io/ioutil"
	"os"
	"strconv"
)

var publicdir = DATADIR + "/public/"
var privatedir = DATADIR + "/private/"

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
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
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
			err = os.Mkdir(privatedir+rootdir, 0666)
			if err != nil {
				errorlog.Println(err)
				return
			} else if _, err := os.Stat(privatedir + rootdir + "/" + child); err != nil {
				err = os.Mkdir(privatedir+rootdir+"/"+child, 0666)
				if err != nil {
					errorlog.Println(err)
					return
				}
			}
		}
		de, _ := ioutil.ReadDir(privatedir + rootdir + "/" + child)
		filename := privatedir + rootdir + "/" + child + "/" + strconv.Itoa(len(de)+1) + ".sf"
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
		if err != nil {
			errorlog.Println(err)
			return
		}
		_, err = f.Write(content)
		if err != nil {
			errorlog.Println(err)
			return
		}
	}
}
