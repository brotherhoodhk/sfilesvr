package model

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/oswaldoooo/octools/toolsbox"
)

var ROOTPATH = os.Getenv("SFILESVR_HOME")
var DatabaseConf = make(map[string]*Dbinfo)
var errorlog = toolsbox.LogInit("error", ROOTPATH+"/logs/error.log")
var AdditionalExtenFunc = make(map[string]func(http.ResponseWriter, *http.Request)) //storage addtional extension function
func init() {
	//read site.xml
	if len(ROOTPATH) < 1 {
		fmt.Println("error>> dont set home environment variable")
		os.Exit(-1)
	}
	content, err := ioutil.ReadFile(ROOTPATH + "/conf/site.xml")
	if err == nil {
		fmt.Println("=====start init model=====")
		var cnfinfo = new(CnfInfo)
		err = xml.Unmarshal(content, cnfinfo)
		if err == nil {
			badconf := 0
			//read database configure information
			for _, info := range cnfinfo.DatabaseInfo.Dbinfo {
				switch info.Class {
				case "mysql":
					DatabaseConf[info.Class] = &info
				default:
					badconf++
				}
			}
			fmt.Printf("read db info %v,bad conf %v\n", len(cnfinfo.DatabaseInfo.Dbinfo), badconf)
			//read plugin information from site.xml
			badconf = 0
			for _, plugin_info := range cnfinfo.PluginInfo.Plugin_Info {
				switch strings.ToLower(plugin_info.ClassName) {
				case "additional extension":
					err = loadextension(plugin_info)
				default:
					err = errors.New(plugin_info.ClassName + " is not support extension class")
				}
				if err != nil {
					//write error into errorlog instead print
					errorlog.Println(err.Error())
					badconf++
				}
			}
			fmt.Printf("read db info %v,bad conf %v\n", len(cnfinfo.PluginInfo.Plugin_Info), badconf)
		}
	}
	if err != nil {
		fmt.Println("error>> ", err)
	}
}

// parse the additional extension
func loadextension(plugin_info plugin) (err error) {
	pluginer, err := toolsbox.ScanPluginByName(plugin_info.Name, ROOTPATH+"/plugin/")
	if err == nil {
		srm, err := pluginer.Lookup("Pattern")
		pattern := *srm.(*string)
		if err == nil {
			if _, ok := AdditionalExtenFunc[pattern]; !ok {
				srm, err = pluginer.Lookup("MainFunc")
				if err == nil {
					AdditionalExtenFunc[pattern] = srm.(func(http.ResponseWriter, *http.Request))
				}
			} else {
				err = errors.New(pattern + " method already used")
			}
		}
	}
	return
}
