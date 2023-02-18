package body

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func ParseList(path string) map[string]string {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		errorlog.Println(err)
		return nil
	} else if len(f) < 3 {
		return make(map[string]string)
	}
	content := string(f)
	basicarr := strings.Split(content, "\n")
	var namelist = make(map[string]string)
	for _, v := range basicarr {
		if len(v) > 2 {
			resarr := strings.Split(v, "=")
			if len(resarr) == 2 {
				//name=path
				namelist[resarr[0]] = resarr[1]
			}
		}
	}
	return namelist
}
func FormatList(origin map[string]string, path string) bool {
	recordmsg := ""
	for k, v := range origin {
		recordmsg += k + "=" + v + "\n"
	}
	err := ioutil.WriteFile(path, []byte(recordmsg), 0666)
	if err != nil {
		fmt.Println("write list to file error")
		errorlog.Println(err)
		return false
	}
	return true
}

// 通过fileid 反向分析出文件实际物理地址,若使用默认版本，version值应该<0
func getfilepath(fileid string, version int) (string, bool) {
	var parentpath string
	if len(fileid) == 7 {
		//private zone
		firdic := fileid[:3]
		secdic := fileid[3:]
		parentpath = privatedir + firdic + "/" + secdic
		if version < 0 {
			//default return last version
			de, err := ioutil.ReadDir(parentpath)
			if err != nil {
				errorlog.Println(parentpath, " is not exist")
				return "", false
			}
			parentpath += "/" + strconv.Itoa(len(de)) + ".sf"
		} else {
			parentpath += "/" + strconv.Itoa(version) + ".sf"
		}
	} else if len(fileid) == 6 {
		//public zone
		parentpath = publicdir + fileid
		if version < 0 {
			//default return last version
			de, err := ioutil.ReadDir(parentpath)
			if err != nil {
				errorlog.Println(parentpath, " is not exist")
				return "", false
			}
			parentpath += "/" + strconv.Itoa(len(de)) + ".sf"
		} else {
			parentpath += "/" + strconv.Itoa(version) + ".sf"
		}
	} else {
		return "", false
	}
	return parentpath, true
}

// get file real path by file id
func getfilepathbyid(fileid string) (string, bool) {
	var parentpath string
	if len(fileid) == 7 {
		//private zone
		firdir := fileid[:3]
		secdir := fileid[3:]
		parentpath = privatedir + firdir + "/" + secdir
	} else if len(fileid) == 6 {
		//public zone
		parentpath = publicdir + fileid
	} else {
		return "", false
	}
	return parentpath, true
}

// 删除云端filesystem 中指定文件
func deletefile(fileid string) bool {
	filepath, ok := getfilepathbyid(fileid)
	if !ok {
		return false
	}
	_, err := os.Stat(filepath)
	if err != nil {
		//file dont exist
		return true
	}
	err = os.RemoveAll(filepath)
	if err != nil {
		errorlog.Println(err)
		return false
	}
	return true
}
