package respository

import (
	"fmt"
	"sfilesvr/body/model"

	"github.com/jmoiron/sqlx"
)

var dbcon *sqlx.DB

// respository init
func init() {
	fmt.Println("=====start init respository=====")
	dbinfo := model.DatabaseConf["mysql"]
	url := fmt.Sprintf("%v:%v@tcp(%v)/%v", dbinfo.User, dbinfo.Password, dbinfo.Address, dbinfo.DBname)
	con, err := sqlx.Connect("mysql", url)
	if err == nil {
		dbcon = con
	} else {
		fmt.Println("init database failed,error>> ", err)
	}

}
