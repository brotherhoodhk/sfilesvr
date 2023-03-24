package service

import (
	"net/http"
	"sfilesvr/body/model"

	"github.com/oswaldoooo/octools/jwttoken"
	"github.com/oswaldoooo/octools/toolsbox"
)

var jwtservice = jwttoken.NewJwt()

// verify the user whether exist,return user's userid and username information
func Verify(r *http.Request) (usr *model.User) {
	token := r.Header.Get("token")
	claim, err := jwtservice.ParseToken(token)
	if err == nil && toolsbox.CheckArgs([]string{"usrid", "usrname"}, claim.Args) {
		usr = &model.User{Name: claim.Args["usrname"], Id: claim.Args["usrid"]}
	}
	return
}
