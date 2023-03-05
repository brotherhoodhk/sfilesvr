package body

import (
	"strconv"
	"strings"
)

var USERPATH = CONFDIR + "/users.cnf"
var FILEPMS = CONFDIR + "/filepms.cnf"

// 获取全部用户和其权限,用户授权码应为固定的4位
func ReadUsers() map[string]*UserInfo {
	allusers := ParseList(USERPATH)
	usrmap := make(map[string]*UserInfo)
	for k, v := range allusers {
		if strings.ContainsRune(v, '@') && len(v) == 12 {
			usrinfoarr := strings.Split(v, "@")
			//判断授权码和权限设置是否符合要求
			if len(strings.Join(usrinfoarr[:len(usrinfoarr)-1], "@")) == 8 && len(usrinfoarr[len(usrinfoarr)-1]) == 3 {
				_, err := strconv.Atoi(usrinfoarr[1])
				if err != nil {
					// processlog.Println("dont suit 2nd step")
					continue
				} else {
					newuser := &UserInfo{Authkey: strings.Join(usrinfoarr[:len(usrinfoarr)-1], "@"), Writeable: [3]bool{}, Readable: [3]bool{}, Execable: [3]bool{}}
					for i := 0; i < 3; i++ {
						permissionnum, _ := strconv.Atoi(usrinfoarr[1][i : i+1])
						switch permissionnum {
						case 0:
							newuser.Readable[i] = false
							newuser.Writeable[i] = false
							newuser.Execable[i] = false
						case 1:
							newuser.Readable[i] = false
							newuser.Writeable[i] = false
							newuser.Execable[i] = true
						case 2:
							newuser.Readable[i] = false
							newuser.Writeable[i] = true
							newuser.Execable[i] = false
						case 3:
							newuser.Readable[i] = false
							newuser.Writeable[i] = true
							newuser.Execable[i] = true
						case 4:
							newuser.Readable[i] = true
							newuser.Writeable[i] = false
							newuser.Execable[i] = false
						case 5:
							newuser.Readable[i] = true
							newuser.Writeable[i] = false
							newuser.Execable[i] = true
						case 6:
							newuser.Readable[i] = true
							newuser.Writeable[i] = true
							newuser.Execable[i] = false
						case 7:
							newuser.Readable[i] = true
							newuser.Writeable[i] = true
							newuser.Execable[i] = true
						}
					}
					usrmap[k] = newuser
				}
			} else {
				// passlangth := len(usrinfoarr[:len(usrinfoarr)-1])
				// permissionnumlangth := len(usrinfoarr[len(usrinfoarr)-1])
				// processlog.Println("dont fit 1st step,and passlangth , permissionlangth", passlangth, "  ", permissionnumlangth)
				// processlog.Println(usrinfoarr[:len(usrinfoarr)-1])
			}
		}
	}
	return usrmap
}
func ReadUsersSecure() map[string]*UserInfo {
	permissionfile := CONFDIR + "/usrpermission.cnf"
	usrauthkey := CONFDIR + "/usrkey.cnf"
	usrpermissionmap := ParseList(permissionfile)
	usrauthkeymap := ParseList(usrauthkey)
	usrmap := make(map[string]*UserInfo)
	for k, permissionnum := range usrpermissionmap {
		if key, ok := usrauthkeymap[k]; ok {
			newuser := &UserInfo{Authkey: key, Writeable: [3]bool{}, Readable: [3]bool{}, Execable: [3]bool{}}
			for i := 0; i < 3; i++ {
				permissionnum, _ := strconv.Atoi(permissionnum)
				switch permissionnum {
				case 0:
					newuser.Readable[i] = false
					newuser.Writeable[i] = false
					newuser.Execable[i] = false
				case 1:
					newuser.Readable[i] = false
					newuser.Writeable[i] = false
					newuser.Execable[i] = true
				case 2:
					newuser.Readable[i] = false
					newuser.Writeable[i] = true
					newuser.Execable[i] = false
				case 3:
					newuser.Readable[i] = false
					newuser.Writeable[i] = true
					newuser.Execable[i] = true
				case 4:
					newuser.Readable[i] = true
					newuser.Writeable[i] = false
					newuser.Execable[i] = false
				case 5:
					newuser.Readable[i] = true
					newuser.Writeable[i] = false
					newuser.Execable[i] = true
				case 6:
					newuser.Readable[i] = true
					newuser.Writeable[i] = true
					newuser.Execable[i] = false
				case 7:
					newuser.Readable[i] = true
					newuser.Writeable[i] = true
					newuser.Execable[i] = true
				}
			}
			usrmap[k] = newuser
		}
	}
	return usrmap
}

// 获得指定用户信息
func GetUserInfo(usrname string) (*UserInfo, bool) {
	usrmap := ReadUsers()
	if _, ok := usrmap[usrname]; !ok {
		return nil, false
	}
	return usrmap[usrname], true
}

// get file or directory permission information
func GetFilePermission(fileid, filepath string) *FilePmsInfo {
	fileinfo := ParseList(filepath)
	if _, ok := fileinfo[fileid]; !ok && strings.Count(fileinfo[fileid], "&&") != 2 {
		return nil
	}
	fileab := strings.Split(fileinfo[fileid], "&&")
	if len(fileab) != 3 {
		return nil
	}
	_, err := strconv.Atoi(fileab[2])
	if err != nil {
		processlog.Println(fileid, "'s permission number is incorrect")
		return nil
	}
	writeable := [3]bool{}
	readable := [3]bool{}
	execable := [3]bool{}
	for i := 0; i < 3; i++ {
		permissionnum, _ := strconv.Atoi(fileab[2][i : i+1])
		switch permissionnum {
		case 0:
			readable[i] = false
			writeable[i] = false
			execable[i] = false
		case 1:
			readable[i] = false
			writeable[i] = false
			execable[i] = true
		case 2:
			readable[i] = false
			writeable[i] = true
			execable[i] = false
		case 3:
			readable[i] = false
			writeable[i] = true
			execable[i] = true
		case 4:
			readable[i] = true
			writeable[i] = false
			execable[i] = false
		case 5:
			readable[i] = true
			writeable[i] = false
			execable[i] = true
		case 6:
			readable[i] = true
			writeable[i] = true
			execable[i] = false
		case 7:
			readable[i] = true
			writeable[i] = true
			execable[i] = true
		}
	}
	newuser := &FilePmsInfo{Owner: fileab[0], Group: fileab[1], Writeable: writeable, Readable: readable, Execable: execable}
	return newuser
}

// check usr permission for file,return [r,w,x]
func CheckPmsForFile(usrname string, usr *UserInfo, fileid, dirid string) [3]bool {
	var filepms = new(FilePmsInfo)
	if len(fileid) == 0 && len(dirid) == 3 {
		//debug line
		// processlog.Println("start search dirid pms")
		filepms = GetFilePermission(dirid, FILEPMS)
		//debug line
		// processlog.Println(filepms)
	} else if len(fileid) == 4 && len(dirid) == 3 {
		filepms = GetFilePermission(fileid, buildpmspath(dirid))
	}
	if filepms == nil {
		errorlog.Println("cant find", dirid+fileid, "info")
		return [3]bool{false, false, false}
	}
	//the extension version 1.0,only support owner compare
	if usrname == filepms.Owner {
		return [3]bool{true, true, true}
	} else {
		respms := [3]bool{}
		respms[0] = filepms.Readable[2]
		respms[1] = filepms.Writeable[2]
		respms[2] = filepms.Execable[2]
		return respms
	}
}
func CheckPmsForFileInterface(usrname, fileid, dirid string) ([3]bool, bool) {
	usrinfo, ok := GetUserInfo(usrname)
	if !ok {
		return [3]bool{}, false
	}
	resarr := CheckPmsForFile(usrname, usrinfo, fileid, dirid)
	return resarr, true
}

// add fileid permission to filepms
func AddToFPMS(fileid, usrname, pmsn, filepath string) bool {
	fpms := ParseList(filepath)
	if _, ok := fpms[fileid]; ok {
		processlog.Println(fileid, " is already in filepms")
		return false
	}
	_, err := strconv.Atoi(pmsn)
	if len(pmsn) != 3 || err != nil {
		processlog.Println("permission number is incorrect")
		return false
	}
	info := usrname + "&&" + usrname + "&&" + pmsn
	fpms[fileid] = info
	FormatList(fpms, filepath)
	return true
}
