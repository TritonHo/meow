package handler

import (
	"errors"
	"net/http"

	"github.com/go-xorm/xorm"
	uuid "github.com/satori/go.uuid"
)

var (
	errNotFound     = errors.New("The record is not found.")
	errUuidNotValid = errors.New("The provided uuid is invalid.")
)

//the id should be a uuid
func getRecord(out interface{}, id string, session *xorm.Session) (statusCode int, err error) {
	if _, err := uuid.FromString(id); err != nil {
		return http.StatusBadRequest, errUuidNotValid
	}

	found, err := session.Id(id).Get(out)

	if err != nil {
		return http.StatusInternalServerError, err
	}
	if found == false {
		return http.StatusNotFound, errNotFound
	}

	return http.StatusOK, nil
}

func getRecordDirect(out interface{}, id string, db *xorm.Engine) (statusCode int, err error) {
	if _, err := uuid.FromString(id); err != nil {
		return http.StatusBadRequest, errUuidNotValid
	}

	found, err := db.Id(id).Get(out)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if found == false {
		return http.StatusNotFound, errNotFound
	}

	return http.StatusOK, nil
}

func getRecordWithUserIdDirect(out interface{}, id, userId string, db *xorm.Engine) (statusCode int, err error) {
	if _, err := uuid.FromString(id); err != nil {
		return http.StatusBadRequest, errUuidNotValid
	}

	found, err := db.Where("id = ? and user_id = ?", id, userId).Get(out)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if found == false {
		return http.StatusNotFound, errNotFound
	}

	return http.StatusOK, nil
}

func deleteRecordWithUserId(input interface{}, id, userId string, session *xorm.Session) (statusCode int, err error) {
	if _, err := uuid.FromString(id); err != nil {
		return http.StatusBadRequest, errUuidNotValid
	}
	if _, err := uuid.FromString(userId); err != nil {
		return http.StatusBadRequest, errUuidNotValid
	}
	affectedCount, err := session.Where("id = ? and user_id = ?", id, userId).Delete(input)

	if err != nil {
		return http.StatusInternalServerError, err
	}
	if affectedCount == 0 {
		return http.StatusNotFound, errors.New("The record is not found.")
	}

	return http.StatusNoContent, err
}

func updateRecordWithUserId(input interface{}, fieldNames map[string]bool, id, userId string, session *xorm.Session) (statusCode int, err error) {
	if _, err := uuid.FromString(id); err != nil {
		return http.StatusBadRequest, errUuidNotValid
	}
	if _, err := uuid.FromString(userId); err != nil {
		return http.StatusBadRequest, errUuidNotValid
	}

	//convert the fields set to array
	array := []string{}
	for k, _ := range fieldNames {
		array = append(array, k)
	}

	//update the database
	affected, err := session.Where("id = ? and user_id = ?", id, userId).Cols(array...).Update(input)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if affected == 0 {
		return http.StatusNotFound, errors.New("The record is not found.")
	}

	return http.StatusNoContent, nil
}

func createRecord(input interface{}, session *xorm.Session) (statusCode int, err error) {
	_, err = session.Insert(input)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, err
}
