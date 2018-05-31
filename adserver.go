package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"html/template"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
	"github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"
	"github.com/gorilla/context"
)

type targetingInfo struct {
	IP      string
	Countries []string
}

type partner struct {
	Id      string        `db:"id"`
	IsSsp   bool          `db:"is_ssp"`
	IsDsp   bool          `db:"is_dsp"`
	Name    string        `db:"name"`
	Timeout time.Duration `db:"timeout"`
	URL     string        `db:"url"`
	Method  string        `db:"method"`
}

type userCollection struct {
	Users []User
}

type User struct {
	Id int `db:"id"`
	Name string `db:"name"`
	Email string `db:"email"`
	Password string `db:"password"`
}

type bid struct {
	Id        string
	URL       string
	Value     float64
	PartnerId string
}

type JwtToken struct {
	Token string `json:"token"`
}
type Data struct {
	Email string
}

var myTemplates = template.Must(template.ParseGlob("tpl/*"))

func FirstHandler(w http.ResponseWriter, r *http.Request){
	decoded := context.Get(r, "decoded")
	var data Data
	mapstructure.Decode(decoded.(jwt.MapClaims), &data)
	json.NewEncoder(w).Encode(data)
}


func HomeHandler(w http.ResponseWriter, r *http.Request){
	data := `{"user_id":"a1b2c3","username":"nikola"}`
	uEnc := base64.URLEncoding.EncodeToString([]byte(data))
	fmt.Println(uEnc)
	err := myTemplates.ExecuteTemplate(w, "home", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request){
	if r.Method == http.MethodGet {
		err := myTemplates.ExecuteTemplate(w, "login", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			fmt.Printf("Cannot parse the form %s \n",err)
			return
		}

		email := r.PostFormValue("email")
		password := r.PostFormValue("password")

		selectAUserQuery := "SELECT * FROM Users WHERE email=?"
		var u User
		err = ds.MySql.Get(&u, selectAUserQuery, email)
		if err != nil {
			fmt.Printf("Cannot get user form db %s \n",err)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
		if err != nil {
			fmt.Printf("Bad password %s \n",err)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		//TOKEN :{header, claims, method}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"email": email,
		})

		tokenString, err := token.SignedString([]byte("secret"))
		if err != nil {
			fmt.Printf("Cannot get the complete signed token %s \n",err)
			return
		}

		cookie := &http.Cookie{Name:"j", Value:tokenString}
		http.SetCookie(w, cookie)

		http.Redirect(w, r, "/first", http.StatusSeeOther)
	}
}

func RegisterHandler(w http.ResponseWriter, r *http.Request){
	if r.Method == http.MethodGet {
		err := myTemplates.ExecuteTemplate(w, "register", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}

		name := r.PostFormValue("name")
		email := r.PostFormValue("email")
		password := r.PostFormValue("password")


		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}

		var createUserQuery = "INSERT Users SET name=?,email=?, password=?"
		_, err = ds.MySql.Exec(createUserQuery, name, email, hashedPassword)
		if err != nil {
			fmt.Printf("Cannot insert user into database:  %v \n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)

		}
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   "e",
		MaxAge: -1}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/login", http.StatusSeeOther)

}

//func AdServerHandler(w http.ResponseWriter, r *http.Request) {
//	if r.Method == http.MethodGet {
//		//_ = r.URL.Query()["ip"][0]
//		//_ = r.URL.Query()["country"][0]
//	}
//	if r.Method == http.MethodPost {
//		body, err := ioutil.ReadAll(r.Body)
//		defer r.Body.Close()
//		if err != nil {
//			http.Error(w, err.Error(), 500)
//			return
//		}
//
//		var targetingInfo targetingInfo
//		err = json.Unmarshal(body, &targetingInfo)
//		if err != nil {
//			http.Error(w, err.Error(), 500)
//			return
//		}
//		fmt.Print(targetingInfo)
//	}
//
//	p := partner{"p1", true, false, "partner1", 200, "", ""}
//
//	partners := []partner{{"dp1", false, true, "dpartner1", 0, "http://localhost:3002", "GET"},
//		{"dp2", false, true, "dpartner2", 0, "http://localhost:3002", "GET"},
//		{"dp3", false, true, "dpartner3", 0, "http://localhost:3002", "GET"},
//		{"dp4", false, true, "dpartner4", 0, "http://localhost:3002", "GET"}}
//
//	ctx, cancel := context.WithTimeout(context.Background(), p.Timeout*time.Millisecond)
//	defer cancel()
//
//	var wg sync.WaitGroup
//	wg.Add(len(partners) + 1)
//
//	bids := make([]bid, 0)
//	receivedBids := make(chan bid, len(partners))
//
//	go func() {
//		defer wg.Done()
//		for range partners {
//			select {
//			case a := <-receivedBids:
//				bids = append(bids, a)
//			case <-ctx.Done():
//				fmt.Println(ctx.Err())
//				return
//			}
//		}
//	}()
//
//	for _, p := range partners {
//		go func() {
//			defer wg.Done()
//			select {
//			case a := <-MakeRequest(p.URL, p.Method, p.Timeout):
//				receivedBids <- a
//			case <-ctx.Done():
//				fmt.Println(ctx.Err())
//			}
//		}()
//	}
//	wg.Wait()
//	bestBid := Max(bids)
//
//	bidJson, err := json.Marshal(bestBid)
//	if err != nil {
//		panic(err)
//	}
//	w.Write(bidJson)
//}

func Max(bids []bid) bid {
	max := 0.0
	for _, v := range bids {
		if v.Value > max {
			max = v.Value
		}
	}
	var bestBid bid
	for _, v := range bids {
		if v.Value == max {
			bestBid = v
		}
	}
	return bestBid
}

func MakeRequest(urlStr, method string, timeout time.Duration) chan bid {
	r, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		panic(err)
	}

	client := &http.Client{Timeout: timeout * time.Millisecond}
	resp, err := client.Do(r)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	var receivedBid bid
	responseBody, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(responseBody, &receivedBid)
	if err != nil {
		fmt.Print(err)
	}
	c := make(chan bid, 1)
	c <- receivedBid

	return c
}
