package middleware

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"meow/lib/auth"
	"meow/lib/lock"

	"github.com/go-xorm/xorm"
	"github.com/gorilla/mux"
	redis "gopkg.in/redis.v3"
)

const (
	DOUBLE_DETECTION_PERIOD = time.Second * 10
	MAX_PROCESS_TIME        = time.Second * 5
)

var (
	db          *xorm.Engine
	redisClient *redis.Client
)

func Init(database *xorm.Engine, client *redis.Client) {
	db = database
	redisClient = client
}

type HandlerWithTx func(r *http.Request, urlValues map[string]string, session *xorm.Session, userId string) (statusCode int, err error, output interface{})
type Handler func(r *http.Request, urlValues map[string]string, db *xorm.Engine, userId string) (statusCode int, err error, output interface{})
type PlainHandler func(res http.ResponseWriter, req *http.Request, urlValues map[string]string, db *xorm.Engine)

type PostHandler func(r io.Reader, urlValues map[string]string, session *xorm.Session, userId string) (statusCode int, err error, output interface{})
type DeleteHandler func(urlValues map[string]string, session *xorm.Session, userId string) (statusCode int, err error, output interface{})

// send a http response to the user with JSON format
func Send(res http.ResponseWriter, statusCode int, data interface{}) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.WriteHeader(statusCode)
	if d, ok := data.([]byte); ok {
		res.Write(d)
	} else {
		json.NewEncoder(res).Encode(data)
	}
}

type cachedResponse struct {
	StatusCode int
	//since golang doesn't have same OOP concept as java
	//thus it is impossible to perfrom serialization of generic error object and then and deserialize back into the same error object
	//And We used a string to store the error message
	ErrMessage *string
	Output     []byte
}

func DoublePostIntercept(f PostHandler) HandlerWithTx {
	return func(r *http.Request, urlValues map[string]string, session *xorm.Session, userId string) (int, error, interface{}) {
		// split the input stream into two
		buffer := new(bytes.Buffer)
		tee := io.TeeReader(r.Body, buffer)

		//perform hashing on one of the stream
		h := md5.New()
		io.Copy(h, tee)
		md5Hash := hex.EncodeToString(h.Sum(nil))

		//the key is request userId + requestUrl + method + hash of request body
		lockName := userId + `-` + r.URL.Path + `-` + r.Method + `-` + md5Hash + `-LOCK`
		resultName := userId + `-` + r.URL.Path + `-` + r.Method + `-` + md5Hash + `-RESULT`

		//ensure that in case of double request, only one thread can get processed
		if ok, err := lock.AcquireLock(lockName, MAX_PROCESS_TIME, MAX_PROCESS_TIME); err != nil {
			return http.StatusInternalServerError, err, nil
		} else if ok == false {
			return http.StatusConflict, err, nil
		}
		defer lock.ReleaseLock(lockName)

		//after entering the critical zone, check if it is a duplicated request
		//if yes, then use the previous output stored in redis, and then do nothing
		if b, err := redisClient.Get(resultName).Bytes(); err != nil && err != redis.Nil {
			return http.StatusInternalServerError, err, nil
		} else if err != redis.Nil {
			c := cachedResponse{}
			json.Unmarshal(b, &c)
			var err error = nil
			if c.ErrMessage != nil {
				err = errors.New(*c.ErrMessage)
			}
			return c.StatusCode, err, c.Output
		}

		//it is not a duplicated request.
		//perform normal processing and then store the result in the redis
		statusCode, err, output := f(buffer, urlValues, session, userId)
		outputBytes, _ := json.Marshal(output)
		c := cachedResponse{StatusCode: statusCode, Output: outputBytes}
		if err != nil {
			s := err.Error()
			c.ErrMessage = &s
		}
		b, _ := json.Marshal(c)
		redisClient.Set(resultName, b, DOUBLE_DETECTION_PERIOD).Result()

		return statusCode, err, output
	}
}

func DoubleDeleteIntercept(f DeleteHandler) HandlerWithTx {
	return func(r *http.Request, urlValues map[string]string, session *xorm.Session, userId string) (int, error, interface{}) {
		//the key is request userId + requestUrl + method + hash of request body
		lockName := userId + `-` + r.URL.Path + `-` + r.Method + `-LOCK`
		resultName := userId + `-` + r.URL.Path + `-` + r.Method + `-RESULT`

		//ensure that in case of double request, only one thread can get processed
		if ok, err := lock.AcquireLock(lockName, MAX_PROCESS_TIME, MAX_PROCESS_TIME); err != nil {
			return http.StatusInternalServerError, err, nil
		} else if ok == false {
			return http.StatusConflict, err, nil
		}
		defer lock.ReleaseLock(lockName)

		//after entering the critical zone, check if it is a duplicated request
		//if yes, then use the previous output stored in redis, and then do nothing
		if b, err := redisClient.Get(resultName).Bytes(); err != nil && err != redis.Nil {
			return http.StatusInternalServerError, err, nil
		} else if err != redis.Nil {
			c := cachedResponse{}
			json.Unmarshal(b, &c)
			var err error = nil
			if c.ErrMessage != nil {
				err = errors.New(*c.ErrMessage)
			}
			return c.StatusCode, err, c.Output
		}

		//it is not a duplicated request.
		//perform normal processing and then store the result in the redis
		statusCode, err, output := f(urlValues, session, userId)
		outputBytes, _ := json.Marshal(output)
		c := cachedResponse{StatusCode: statusCode, Output: outputBytes}
		if err != nil {
			s := err.Error()
			c.ErrMessage = &s
		}
		b, _ := json.Marshal(c)
		redisClient.Set(resultName, b, DOUBLE_DETECTION_PERIOD).Result()

		return statusCode, err, output
	}
}

// a middleware to handle user authorization
func AuthAndTx(f HandlerWithTx) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		userId, err := auth.Verify(req.Header.Get("Authorization"))
		if err != nil {
			Send(res, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		} else {
			if newToken, err := auth.Sign(userId); err != nil {
				Send(res, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			} else {
				res.Header().Add("Authorization", newToken) // update JWT Token
			}
		}

		//prepare a database session for the handler
		session := db.NewSession()
		if err := session.Begin(); err != nil {
			Send(res, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		defer session.Close()

		//everything seems fine, goto the business logic handler
		if statusCode, err, output := f(req, mux.Vars(req), session, userId); err == nil {
			//the business logic handler return no error, then try to commit the db session
			if err := session.Commit(); err != nil {
				Send(res, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			} else {
				Send(res, statusCode, output)
			}
		} else {
			session.Rollback()
			Send(res, statusCode, map[string]string{"error": err.Error()})
		}
	}
}

// a middleware to handle user authorization
func Auth(f Handler) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		userId, err := auth.Verify(req.Header.Get("Authorization"))
		if err != nil {
			Send(res, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		} else {
			if newToken, err := auth.Sign(userId); err != nil {
				Send(res, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			} else {
				res.Header().Add("Authorization", newToken) // update JWT Token
			}
		}

		//everything seems fine, goto the business logic handler
		if statusCode, err, output := f(req, mux.Vars(req), db, userId); err == nil {
			Send(res, statusCode, output)
		} else {
			Send(res, statusCode, map[string]string{"error": err.Error()})
		}
	}
}

//do nothing and provide injection of database object only
//normally it is used by public endpoint
func Plain(f PlainHandler) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		f(res, req, mux.Vars(req), db)
	}
}
