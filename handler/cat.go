package handler

import (
	"io"
	"net/http"

	"framework-demo/lib/httputil"
	"framework-demo/model"

	"github.com/go-xorm/xorm"
	"github.com/satori/go.uuid"
)

func CatGetOne(r *http.Request, urlValues map[string]string, db *xorm.Engine, userId string) (int, error, interface{}) {
	//FIXME: add object level privilege checking

	cat := model.Cat{}
	statusCode, err := getRecordDirect(&cat, urlValues["catId"], db)

	return statusCode, err, cat
}

func CatUpdate(r *http.Request, urlValues map[string]string, session *xorm.Session, userId string) (int, error, interface{}) {
	cat := model.Cat{}
	dbUpdateFields, _, err := httputil.BindForUpdate(r.Body, &cat)
	if err != nil {
		return http.StatusBadRequest, err, nil
	}
	statusCode, err := updateRecordWithUserId(&cat, dbUpdateFields, urlValues["catId"], userId, session)
	return statusCode, err, nil
}

func CatCreate(r io.Reader, urlValues map[string]string, session *xorm.Session, userId string) (int, error, interface{}) {
	cat := model.Cat{}
	if err := httputil.Bind(r, &cat); err != nil {
		return http.StatusBadRequest, err, nil
	}

	cat.Id = uuid.NewV4().String()
	cat.UserId = userId

	if statusCode, err := createRecord(&cat, session); err != nil {
		return statusCode, err, nil
	} else {
		return http.StatusOK, nil, map[string]string{"id": cat.Id}
	}
}

func CatDelete(urlValues map[string]string, session *xorm.Session, userId string) (int, error, interface{}) {
	statusCode, err := deleteRecordWithUserId(&model.Cat{}, urlValues["catId"], userId, session)
	return statusCode, err, nil
}
