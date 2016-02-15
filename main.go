package main

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"log"

	"framework-demo/handler"
	"framework-demo/lib/auth"
	"framework-demo/lib/config"
	"framework-demo/lib/httputil"
	"framework-demo/lib/lock"
	"framework-demo/lib/middleware"
	"framework-demo/setting"

	jwt "github.com/dgrijalva/jwt-go"
	xormCore "github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	redis "gopkg.in/redis.v3"
)

func main() {
	showDevAuth()
	initDependency()

	//in old go compiler, it is a must to enable multithread processing
	runtime.GOMAXPROCS(runtime.NumCPU())

	router := mux.NewRouter()

	router.HandleFunc("/v1/auth", middleware.Plain(handler.Login)).Methods("POST")

	router.HandleFunc("/v1/user", middleware.Plain(handler.UserCreate)).Methods("POST")

	router.HandleFunc("/v1/cats/{catId}", middleware.Auth(handler.CatGetOne)).Methods("GET")
	router.HandleFunc("/v1/cats/{catId}", middleware.AuthAndTx(handler.CatUpdate)).Methods("PUT")
	router.HandleFunc("/v1/cats/{catId}", middleware.AuthAndTx(middleware.DoubleDeleteIntercept(handler.CatDelete))).Methods("DELETE")
	router.HandleFunc("/v1/cats", middleware.AuthAndTx(middleware.DoublePostIntercept(handler.CatCreate))).Methods("POST")

	http.Handle("/", router)
	s := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Fatal(s.ListenAndServe())
}

// init the various object and inject the database object to the modules
func initDependency() {
	//the postgresql connection string
	connectStr := "host=" + config.GetStr(setting.DB_HOST) +
		" port=" + strconv.Itoa(config.GetInt(setting.DB_PORT)) +
		" dbname=" + config.GetStr(setting.DB_NAME) +
		" user=" + config.GetStr(setting.DB_USERNAME) +
		" password='" + config.GetStr(setting.DB_PASSWORD) +
		"' sslmode=disable"

	//db, err := gorm.Open("postgres", connectStr)
	db, err := xorm.NewEngine("postgres", connectStr)

	if err != nil {
		log.Panic("DB connection initialization failed", err)
	}

	db.SetMaxIdleConns(config.GetInt(setting.DB_MAX_IDLE_CONN))
	db.SetMaxOpenConns(config.GetInt(setting.DB_MAX_OPEN_CONN))
	db.SetColumnMapper(xormCore.SnakeMapper{})
	//uncomment it if you want to debug
	// db.ShowSQL = true
	// db.ShowErr = true

	//setup the redis
	redisOptions := redis.Options{
		Addr:     config.GetStr(setting.REDIS_ENDPOINT),
		PoolSize: config.GetInt(setting.REDIS_POOL_SIZE),
		Network:  "tcp",

		//FIXME: check the purpose of these timeout
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	redisClient := redis.NewClient(&redisOptions)

	//load the RSA key from the file system, for the jwt auth
	var err1 error
	var currentKey *rsa.PrivateKey = nil
	var oldKey *rsa.PrivateKey = nil

	currentKeyBytes, _ := ioutil.ReadFile(config.GetStr(setting.JWT_RSA_KEY_LOCATION))
	currentKey, err1 = jwt.ParseRSAPrivateKeyFromPEM(currentKeyBytes)
	if err1 != nil {
		log.Panic(err1)
	}
	if location := config.GetStr(setting.JWT_OLD_RSA_KEY_LOCATION); location != `` {
		oldKeyBytes, _ := ioutil.ReadFile(location)
		oldKey, err1 = jwt.ParseRSAPrivateKeyFromPEM(oldKeyBytes)
		if err1 != nil {
			log.Panic(err1)
		}
	}
	lifetime := time.Duration(config.GetInt(setting.JWT_TOKEN_LIFETIME)) * time.Minute
	auth.Init(currentKey, oldKey, lifetime)

	httputil.Init(xormCore.SnakeMapper{})

	//add the db dependency to middleware module
	middleware.Init(db, redisClient)

	//add the redis dependency to lock module
	lock.Init(redisClient)
}

func showDevAuth() {
	currentKeyBytes, _ := ioutil.ReadFile(config.GetStr(setting.JWT_RSA_KEY_LOCATION))
	currentKey, err1 := jwt.ParseRSAPrivateKeyFromPEM(currentKeyBytes)
	if err1 != nil {
		log.Panic(err1)
	}

	token := jwt.New(jwt.SigningMethodRS512)

	// Set some claims
	token.Claims["userId"] = `eeee1df4-9fae-4e32-98c1-88f850a00001`
	token.Claims["exp"] = time.Now().Add(time.Minute * 60 * 24 * 30).Unix()

	// Sign and get the complete encoded token as a string
	tokenString, _ := token.SignedString(currentKey)
	fmt.Println("Please put the following string into http 'Authorization' header:")
	fmt.Println(tokenString)
}
