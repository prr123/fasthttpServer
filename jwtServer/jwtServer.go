// https://golangr.com/golang-http-serve
// modified
//
// Author: prr
// Date: 13 June 2022
// copyright prr, azul software
//


package main

import (
	"fmt"
    "log"
    "net/http"
	"bytes"
	"io"

	"github.com/goccy/go-json"
)

type userInfo struct {
    User string `json:"username"`
    Pwd string `json:"password"`
}


type myHandler struct{}

func (*myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("received %s from %s\n", r.Method, r.RemoteAddr)
	fmt.Printf("Proto:  %s\n", r.Proto)
	fmt.Printf("Request URI:  %s\n", r.RequestURI)

	fmt.Println("Header")
	for k,v := range r.Header {
		fmt.Printf("key: %s value: %s\n", k, v)
	}
    log.Printf("URL: %s\n",r.URL.String())
	authStr := r.Header["Authorization"]
	fmt.Printf("Auth[%d]: %s\n", len(authStr), authStr)
	for i:=0; i<len(authStr); i++ {
		fmt.Printf(" %d: %s\n",i , authStr[i])
	}
	if len(authStr) != 1 {
		log.Printf("invalid auth string!\n")
		return
	}

	bear:= []byte(authStr[0])
	idx:=bytes.Index(bear,[]byte("Bearer "))
	if idx == -1 {
		fmt.Printf("no Bearer keyword!")
		fmt.Fprintf(w,"no Bearer keyword!\n")
		return
	}

	token := string(bear[idx+7:])
	log.Printf("token: %s\n", token)
	fmt.Fprintf(w,"hello client\nauth: %s\n", token)
}

func Tmp(w http.ResponseWriter, r *http.Request) {
    log.Printf("tmp\n")
    fmt.Fprintf(w,"tmp\n")
}

func Signin(w http.ResponseWriter, r *http.Request) {
    log.Printf("signin\n")
	fmt.Printf("received %s from %s\n", r.Method, r.RemoteAddr)
	fmt.Printf("Proto:  %s\n", r.Proto)
	fmt.Printf("Request URI:  %s\n", r.RequestURI)
	fmt.Printf("content-type: %s\n", r.Header.Get("content-type"))

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("server: could not read request body: %s\n", err)
		return
	}

	fmt.Printf("Request body:\n%s\n*** end request body ***\n", string(reqBody))

	userData := userInfo{}

	if err := json.Unmarshal(reqBody, &userData); err != nil {
        log.Fatalf("error Json Unmarshal: %v\n", err)
    }

	fmt.Printf("User Data: %v\n",userData)
    fmt.Fprintf(w,"signin success\n")
}


func main(){


    mux := http.NewServeMux()

    mux.HandleFunc("/tmp", Tmp)

    mux.HandleFunc("/signin", Signin)


    // Register routes and register handlers in this form.
    mux.Handle("/",&myHandler{})


    //http.ListenAndServe uses the default server structure.
	portStr := "12011"
	log.Printf("listening on port: %s!\n", portStr)
    err := http.ListenAndServe(":"+portStr, mux)
    if err != nil {log.Fatalf("ListenAndServe: %v\n", err)}
}
