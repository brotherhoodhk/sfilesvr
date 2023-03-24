package service

import (
	"encoding/json"
	"net/http"
	"sfilesvr/body/model"
	"sfilesvr/body/respository"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

var usrrespository = respository.NewUserRespository("user_info")

func Login(w http.ResponseWriter, r *http.Request) {
	usrid := r.Form.Get("usrid")
	passwd := r.Form.Get("password")
	//verify the user is valid
	var code int
	var token string
	res := make(map[string]string)
	usr := new(model.User)
	err := usrrespository.FindUser(usrid, usr)
	if err == nil {
		err = bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(passwd))
		if err == nil {
			code = 200
			//generate jwt token
			token, err = jwtservice.GenerateToken(map[string]string{"usrid": usr.Id, "usrname": usr.Name})
		}
	}
	resbytes := make([]byte, 8<<20)
	if err == nil {
		res["token"] = token
	} else {
		code = 400
		res["error"] = err.Error()
	}
	res["code"] = strconv.Itoa(code)
	resbytes, _ = json.Marshal(res)
	w.Write(resbytes)
}
