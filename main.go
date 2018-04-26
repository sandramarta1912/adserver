package main

import (
	"github.com/gorilla/mux"
	"fmt"
	"net/http"
)

func main(){
	mockedserver := mux.NewRouter()

	mockedserver.HandleFunc("/", AdServerHandler)

	fmt.Println("Server started on 3003...")
	http.ListenAndServe(":3003", mockedserver)
}
