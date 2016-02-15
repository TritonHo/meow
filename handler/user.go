package handler

import (
	//	"log"
	"net/http"

	"meow/lib/auth"
	"meow/lib/httputil"
	"meow/lib/middleware"
	"meow/model"

	"github.com/go-xorm/xorm"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
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

func UserCreate(w http.ResponseWriter, r *http.Request, urlValues map[string]string, db *xorm.Engine) {
	user := struct {
		model.User `xorm:"extends"`
		Password   string `xorm:"-" json:"password" validate:"required"`
	}{}

	if err := httputil.Bind(r.Body, &user); err != nil {
		middleware.Send(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	user.Id = uuid.NewV4().String()

	if digest, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost); err != nil {
		middleware.Send(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	} else {
		user.PasswordDigest = string(digest)
	}

	session := db.NewSession()
	if err := session.Begin(); err != nil {
		middleware.Send(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer session.Close()

	if statusCode, err := createRecord(&user, session); err != nil {
		middleware.Send(w, statusCode, map[string]string{"error": err.Error()})
		return
	}

	if err := session.Commit(); err != nil {
		middleware.Send(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if newToken, err := auth.Sign(user.Id); err != nil {
		middleware.Send(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	} else {
		// update JWT Token
		w.Header().Add("Authorization", newToken)
		//allow CORS
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")
		middleware.Send(w, http.StatusOK, map[string]string{"userId": user.Id})
	}
}
