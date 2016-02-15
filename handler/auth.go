package handler

import (
	//"errors"
	//	"log"
	"net/http"

	"framework-demo/lib/auth"
	"framework-demo/lib/httputil"
	"framework-demo/lib/middleware"
	"framework-demo/model"

	"github.com/go-xorm/xorm"
	"golang.org/x/crypto/bcrypt"
)

func Login(w http.ResponseWriter, r *http.Request, urlValues map[string]string, db *xorm.Engine) {
	//handle the input
	var input struct {
		Email    string `json:"email" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
	if err := httputil.Bind(r.Body, &input); err != nil {
		middleware.Send(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	user := model.User{}
	found, err := db.Where("email = ?", input.Email).Get(&user)
	if err != nil {
		middleware.Send(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if found == false || bcrypt.CompareHashAndPassword([]byte(user.PasswordDigest), []byte(input.Password)) != nil {
		middleware.Send(w, http.StatusUnauthorized, map[string]string{"error": "Incorrect Email / Password"})
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
