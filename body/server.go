package body

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

var list = ParseList(ROOTPATH + "/conf/site.cnf")
var port = 8001

func init() {
	fmt.Println("start initiazation...")
	_, err := os.Stat(ROOTPATH + "/conf/site.cnf")
	if err != nil {
		fmt.Println("site config file dont exist")
		os.Exit(-1)
	}
	_, err = os.Stat(ROOTPATH + "/data/public")
	if err != nil {
		fmt.Println("public dic dont exist")
		err = os.Mkdir(ROOTPATH+"/data/public", 0666)
		if err != nil {
			fmt.Println("cant create public data dir")
			os.Exit(-1)
		}
	}
	if pt, ok := list["port"]; ok {
		ports, err := strconv.Atoi(pt)
		if err != nil {
			fmt.Println("your port set is not correct")
		} else {
			port = ports
		}
	}
}
func ServerStart() {
	go hub.Run()
	http.HandleFunc("/singlefile", AcceptFile)
	fmt.Println("listen at port ", port)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
