package respository

import (
	"errors"
	"fmt"
	"sfilesvr/body/model"
	"text/template"

	"golang.org/x/crypto/bcrypt"
)

type Respository interface {
	Create()
	Find()
	Delete()
	Update()
}
type UserRespository struct {
	table_name string
}

func NewUserRespository(tablename string) *UserRespository {
	return &UserRespository{table_name: tablename}
}
func (s *UserRespository) FindUser(id string, usr *model.User) (err error) {
	esql := fmt.Sprintf("select id,name,password from %v where id=%v", s.table_name, id)
	err = dbcon.Get(&usr, esql)
	return
}
func (s *UserRespository) CreateUser(usr *model.User) (err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(usr.Password), bcrypt.DefaultCost)
	if err == nil {
		usr.Password = string(hash)
		// usr.Password = template.HTMLEscapeString(usr.Password)
	} else {
		return
	}
	esql := fmt.Sprintf("insert into %v (id,name,password)values('%v','%v','%v')", s.table_name, template.HTMLEscapeString(usr.Id), template.HTMLEscapeString(usr.Name), template.HTMLEscapeString(usr.Password))
	sr, err := dbcon.Exec(esql)
	if err == nil {
		rows, err := sr.RowsAffected()
		if err == nil && rows <= 0 {
			err = errors.New("create user failed,unknown error")
		}
	}
	if err != nil {
		fmt.Println(esql)
	}
	return
}
func (s *UserRespository) UpdateUser(usr *model.User) (err error) {
	esql := fmt.Sprintf("update %v set name=%v,password=%v where id=%v", s.table_name, usr.Name, usr.Password, usr.Id)
	_, err = dbcon.Exec(esql)
	return
}
