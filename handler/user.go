package handler

import (
	//"errors"
	//	"log"
	"net/http"
	//	"strings"

	//"framework-demo/lib/httputil"
	"framework-demo/model"

	"github.com/go-xorm/xorm"

	//"golang.org/x/crypto/bcrypt"
)

func UserGetOne(res http.ResponseWriter, req *http.Request, urlValues map[string]string, session *xorm.Session, userId string) (error, int, interface{}) {
	//FIXME: add object level privilege checking

	user := model.User{}
	statusCode, err := getRecord(&user, urlValues["userId"], session)

	return err, statusCode, user
}

/*
func UserUpdate(w http.ResponseWriter, r *http.Request, urlValues map[string]string, authInput *auth.AuthInput) {
	//TODO: not handled subuser concept
	if authInput.User.Id != urlValues["userId"] {
		hu.SendErr(w, http.StatusUnauthorized, errors.New("Not allowed access to other users"))
		return
	}
	user := model.User{}
	dbUpdateFields, _, err := hu.BindForUpdate(r, &user)
	if err != nil {
		hu.SendErr(w, http.StatusBadRequest, err)
		return
	}
	if statusCode, err := updateRecord(&user, dbUpdateFields, urlValues["userId"]); err != nil {
		hu.SendErr(w, statusCode, err)
	} else {
		hu.Send(w, http.StatusOK, hu.M{"id": urlValues["userId"]})
	}
}
*/
