// fastHttpServer
// testing fast http
//
// Author: prr, azul software
// Date 5 February 2024
// copyright (c) 2024 prr, azul software
//

// v2: add cli
// V3: cleanup handlers

package main

import (
	"os"
	"log"
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"
    util "github.com/prr123/utility/utilLib"
)

type Handler struct {
	dbg bool
}

func main() {

    numarg := len(os.Args)
    flags:=[]string{"dbg", "port"}

    useStr := " /port=portstr [/dbg]"
    helpStr := "fasthttp server"

    if numarg > len(flags) +1 {
        fmt.Println("too many arguments in cl!")
        fmt.Println("usage: %s %s\n", os.Args[0], useStr)
        os.Exit(-1)
    }

    if numarg == 1 || (numarg > 1 && os.Args[1] == "help") {
        fmt.Printf("help: %s\n", helpStr)
        fmt.Printf("usage is: %s %s\n", os.Args[0], useStr)
        os.Exit(1)
    }

    flagMap, err := util.ParseFlags(os.Args, flags)
    if err != nil {log.Fatalf("util.ParseFlags: %v\n", err)}

	dbg:= false
    _, ok := flagMap["dbg"]
    if ok {dbg = true}

    portStr := ""
    pval, ok := flagMap["port"]
    if !ok {
        log.Fatalf(" error no port provided!\n")
    } else {
        if pval.(string) == "none" {log.Fatalf("error: no port value provided!\n")}
        portStr = pval.(string)
    }

	han := Handler{
		dbg: dbg,
	}


	log.Printf("info -- starting to listen at port: %s\n", portStr)
	fasthttp.ListenAndServe(":"+portStr, han.requestHandler)

}

	// the corresponding fasthttp request handler
func (han Handler)requestHandler(ctx *fasthttp.RequestCtx) {

	log.Printf("dbg: %t request: %s method: %s\n", han.dbg, ctx.RequestURI(), ctx.Method())

	switch string(ctx.Path()) {
		case "/foo":
			han.fooHandler(ctx)
		case "/bar":
			han.barHandler(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
	}
}

	// request handler in fasthttp style, i.e. just plain function.
func (han Handler)fooHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Hi there! foo here! RequestURI is %q! dbg: %t body:\n", ctx.RequestURI(), han.dbg)
	fmt.Fprintf(ctx, "Hello, world!\n\n")

	fmt.Fprintf(ctx, "Request method is %q\n", ctx.Method())
	fmt.Fprintf(ctx, "RequestURI is %q\n", ctx.RequestURI())
	fmt.Fprintf(ctx, "Requested path is %q\n", ctx.Path())
	fmt.Fprintf(ctx, "Host is %q\n", ctx.Host())
	fmt.Fprintf(ctx, "Query string is %q\n", ctx.QueryArgs())
	fmt.Fprintf(ctx, "User-Agent is %q\n", ctx.UserAgent())
	fmt.Fprintf(ctx, "Connection has been established at %s\n", ctx.ConnTime())
	fmt.Fprintf(ctx, "Request has been started at %s\n", ctx.Time())
	fmt.Fprintf(ctx, "Serial request number for the current connection is %d\n", ctx.ConnRequestNum())
//	fmt.Fprintf(ctx, "Your source adr is %q\n", ctx.RemoteAddr.String())
	con:= ctx.Conn()
	adr := con.RemoteAddr()
	fmt.Fprintf(ctx, "remote addr: %q \n", adr.String())
	idx := strings.Index(adr.String(), ":")
	port := adr.String()[idx+1:]
	fmt.Fprintf(ctx, "port: %s\n", port)
	// unique id
	fmt.Fprintf(ctx,"connection seq: %d\n",  ctx.ConnRequestNum())
	if ctx.ConnRequestNum() == 1 {fmt.Fprintf(ctx,"need to login!\n")}

	fmt.Fprintf(ctx,"connection id: %d\n\n", ctx.ConnID())
	
	authVal:= ctx.Request.Header.Peek("Authorization")
	fmt.Fprintf(ctx,"auth value: %s\n", authVal)

	numHeaders := ctx.Request.Header.Len()
	head := ctx.Request.Header.RawHeaders()
	fmt.Fprintf(ctx, "headers [%d]:\n%s\n", numHeaders, head)
	fmt.Fprintf(ctx, "end headers\n")
	fmt.Fprintf(ctx, "\nRaw request is:\n---START---\n%s\n---END---", &ctx.Request)

	ctx.SetContentType("text/plain; charset=utf8")

/*
	// Set arbitrary headers
	ctx.Response.Header.Set("X-My-Header", "my-header-value")

	// Set cookies
	var c fasthttp.Cookie
	c.SetKey("cookie-name")
	c.SetValue("cookie-value")
*/
}

func (han Handler)barHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Hi there! bar here! RequestURI is %q", ctx.RequestURI())
}

