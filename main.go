package main

import (
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"

	"fmt"
	"net/http"
	"time"
	"os"
	"log"
	"github.com/dgrijalva/jwt-go"
	"encoding/json"
	"github.com/gorilla/context"
)

type datastore struct {
	MySql *sqlx.DB
}


var ds datastore

func InitMySqlConn(dsn string) (*sqlx.DB, error) {
	var err error
	ds.MySql, err = sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	ds.MySql.SetMaxOpenConns(50)

	ds.MySql.SetMaxIdleConns(20)
	ds.MySql.SetConnMaxLifetime(time.Hour)

	if err = ds.MySql.Ping(); err != nil {
		return nil, err
	}

	return ds.MySql, nil
}

func main(){
	mysqlDsn :=  os.Getenv("MYSQL_DSN")
	//mysqlDsn := "%s:%s@tcp(%s:%s)/%s?timeout=30s&readTimeout=1s&writeTimeout=1s"
	//mysqlDsn = fmt.Sprintf(
	//	mysqlDsn,
	//	os.Getenv("MYSQL_ROOT"),
	//	os.Getenv("MYSQL_ROOT_PASSWORD"),
	//	os.Getenv("MYSQL_HOST"),
	//	os.Getenv("MYSQL_PORT"),
	//	os.Getenv("MYSQL_DATABASE"))

	fmt.Printf("Initializing MySql connection to: %s\n", mysqlDsn)

	trials, maxTrials := 0, 15 // TODO import maxTrials value from a config file

	for {
		db, err := InitMySqlConn(mysqlDsn)
		if err != nil {
			log.Printf("Unable to connect to MySql (trial %d): %s\n", trials, err)
			time.Sleep(time.Duration(1) * time.Second)
			trials++
			if trials >= maxTrials {
				os.Exit(1)
			}
		} else {
			fmt.Printf("Connected to db\n")
			defer db.Close()
			break
		}
	}


	adserver := mux.NewRouter()

	adserver.HandleFunc("/", HomeHandler)
	adserver.HandleFunc("/login", LoginHandler)
	adserver.HandleFunc("logout", LogoutHandler)
	adserver.HandleFunc("/register", RegisterHandler)
	adserver.Handle("/first", ValidateMiddleware(http.HandlerFunc(FirstHandler)))


	fmt.Println("Server started on 3003...")
	http.ListenAndServe(":3003", adserver)
}

func ValidateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("j")
		if err != nil {
			fmt.Printf("Error at cookie:  %v \n", err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		bearerToken := c.Value
		if bearerToken != "" {
			token, err := jwt.Parse(bearerToken, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}
				return []byte("secret"), nil
			})
			if err != nil {
				json.NewEncoder(w).Encode(err.Error())
				return
			}
			if token.Valid {
				context.Set(r, "decoded", token.Claims)
				next(w, r)
			} else {
				json.NewEncoder(w).Encode("Invalid authorization token")
			}

		} else {
			json.NewEncoder(w).Encode("An authorization header is required")
		}
	})
}
