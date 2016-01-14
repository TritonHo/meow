package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"framework-demo/handler"
	"framework-demo/lib/config"
	"framework-demo/lib/httputil"
	"framework-demo/lib/middleware"
	"framework-demo/setting"

	jwt "github.com/dgrijalva/jwt-go"
	xormCore "github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	showDevAuth()
	initDependency()

	//in old go compiler, it is a must to enable multithread processing
	runtime.GOMAXPROCS(runtime.NumCPU())

	router := mux.NewRouter()

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

	httputil.Init(xormCore.SnakeMapper{})

	//add the db dependency to middleware module
	middleware.Init(db)
}

func showDevAuth() {

	//FIXME switch to RS256
	secret := `abc123`
	token := jwt.New(jwt.SigningMethodHS256)

	// Set some claims
	token.Claims["userId"] = `eeee1df4-9fae-4e32-98c1-88f850a00001`
	token.Claims["exp"] = time.Now().Add(time.Minute * 60 * 24 * 30).Unix()

	// Sign and get the complete encoded token as a string
	tokenString, _ := token.SignedString([]byte(secret))
	fmt.Println("Please put the following string into http 'Authorization' header:")
	fmt.Println(tokenString)
}
