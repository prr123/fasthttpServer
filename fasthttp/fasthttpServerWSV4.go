// fastHttpServerWsV3
// building a webserver based on wasgob and fasthttp
//
// Author: prr, azul software
// Date 29 June 2024
// copyright (c) 2024 prr, azul software
//

// V2 add RTyp and expand handler type
// add router
// test websocket
//
// v3 create upgrader routine
//

package main

import (
	"os"
	"log"
	"fmt"
//	"bytes"
	"strings"
	"net"
	"io"

	"server/http/fasthttp/upgrader"
//	"server/http/fasthttp/wsnonce"
	"server/http/fasthttp/pathparser"
	"github.com/gobwas/ws"
	"github.com/valyala/fasthttp"
    util "github.com/prr123/utility/utilLib"
)

type Rtyp struct {
    fil *os.File
    ftyp string
    }

type Handler struct {
	dbg bool
    router map[string] Rtyp
	p pparse.Path
}


func main() {

	wwwBase := "/home/peter/www/azultest/"

    numarg := len(os.Args)
    flags:=[]string{"dbg", "port", "index"}

    useStr := " /port=portstr [/index=idxfil [/dbg]"
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

    rootFil := "indexjsV4"
    rval, ok := flagMap["index"]
    if ok {
        if rval.(string) == "none" {log.Fatalf("error: no index file name provided!\n")}
        rootFil = rval.(string)
    }



	han := Handler{
		dbg: dbg,
	}
    han.router = make(map[string]Rtyp)

    idxFilnam := wwwBase + "html/" + rootFil + ".html"
    fil, err := os.Open(idxFilnam)
    if err != nil {log.Fatalf("error -- cannot open index file: %v\n", err)}
    idx := Rtyp {
        fil: fil,
        ftyp: "text/html",
    }
    han.router["index.html"] = idx

    libfilnam := wwwBase + "js/azulLibV8.js"
    libfil, err := os.Open(libfilnam)
    if err != nil {log.Fatalf("error -- cannot open index file: %v\n", err)}
    defer libfil.Close()
    idx.fil = libfil
    idx.ftyp = "text/javascript"
    han.router["azulLib.js"] = idx

    stfilnam := wwwBase + "js/azulstartV2.js"
    stfil, err := os.Open(stfilnam)
    if err != nil {log.Fatalf("error -- cannot open index file: %v\n", err)}
    defer stfil.Close()
    idx.fil = stfil
    han.router["azulstart.js"] = idx

   	xpfilnam := wwwBase + "js/azulxp.js"
    xpfil, err := os.Open(xpfilnam)
    if err != nil {log.Fatalf("error -- cannot open xp file: %v\n", err)}
    defer xpfil.Close()
//  han.start = stfil
    idx.fil = xpfil
    han.router["azulxp.js"] = idx


	if dbg {
		fmt.Println("************ setup ****************")

		fmt.Println("********** end setup **************")
	}

	log.Printf("info -- starting to listen at port: %s\n", portStr)
	fasthttp.ListenAndServe(":"+portStr, han.requestHandler)

}


	// the corresponding fasthttp request handler
func (han Handler)requestHandler(ctx *fasthttp.RequestCtx) {

	if han.dbg {log.Printf("dbg request: %s method: %s path: %s\n", ctx.RequestURI(), ctx.Method(), ctx.Path())}

	// find etension and folder path -> parse ctx.PATH

	p := pparse.Pparse(ctx.Path())
	if han.dbg {log.Printf("dbg Fold: %s Fnam: %s Ext: %s\n", p.Fold, p.Fnam, p.Ext)}
	han.p = p

	switch string(p.Fold) {
		case "/","":
			han.idxHandler(ctx)
		case "/foo":
			han.fooHandler(ctx)
		case "/bar":
			han.barHandler(ctx)
		case "/hijack":
			// Note that the connection is hijacked only after
			// returning from requestHandler and sending http response.
			han.wsHandler(ctx)

		default:
			if han.dbg {log.Println("unsupported path!")}

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

func (han Handler)idxHandler(ctx *fasthttp.RequestCtx) {

	if han.dbg {log.Printf("dbg index %s %s %s\n", han.p.Fold, han.p.Fnam, han.p.Ext)}

	tgt := "index.html"
	if len(han.p.Fnam) > 1 {tgt=string(han.p.Fnam)}

    if han.dbg {fmt.Printf("dbg -- tgt: %s\n", tgt)}
    ridx, ok := han.router[tgt]
    if !ok {
        fmt.Printf("error -- invalid req: %s\n", tgt)
        ctx.SetStatusCode(404)
        ctx.SetContentType("text/plain; charset=utf-8")
        fmt.Fprintf(ctx, "invalid req: %s\n",tgt)
        return
    }
    if han.dbg {fmt.Printf("dbg -- %s\n", ridx.ftyp)}
    ctx.SetContentType(ridx.ftyp + "; charset=utf-8")

//  m, err := io.Copy(os.Stdout, ridx.fil)
    ridx.fil.Seek(0,0)
    n, err := io.Copy(ctx, ridx.fil)
    if err != nil {
        log.Printf("error -- io.Copy: %v\n", err)
        return
    }
	return
    if han.dbg {fmt.Printf("dbg -- sent: %d\n", n)}


}

func (han Handler)barHandler(ctx *fasthttp.RequestCtx) {
//	fmt.Fprintf(ctx, "Hi there! bar here! RequestURI is %q", ctx.RequestURI())
//	resp:= fasthttp.AcquireResponse()
	ctx.SetStatusCode(401)
    ctx.SetContentType("text/plain; charset=utf8")
	ctx.Response.Header.Set("Authorisation", "abcdefg")
//	ctx.Response.
	ctx.SetBodyString("hello -- this is a test string!\n")
}

/* 
// web socket upgrade
GET /chat HTTP/1.1
Host: example.com:8000
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
Sec-WebSocket-Version: 13
*/

func (han Handler)wsHandler(ctx *fasthttp.RequestCtx) {


	if han.dbg {
		log.Printf("ws upgrade!\n")
		upgVal:= ctx.Request.Header.Peek("Upgrade")
		fmt.Printf("upgrade value: %s\n", upgVal)
		conVal:= ctx.Request.Header.Peek("Connection")
		fmt.Printf("connection value: %s\n", conVal)
		wsKeyVal:= ctx.Request.Header.Peek("Sec-WebSocket-Key")
		fmt.Printf("ws Key value: %s\n", wsKeyVal)
		wsVerVal:= ctx.Request.Header.Peek("Sec-WebSocket-Version")
		fmt.Printf("ws version: %s\n", wsVerVal)
	}

	// Upgrade(ctx *fasthttp.RequestCtx, dbg bool)(err error)
	err := upgrader.Upgrade(ctx, han.dbg)
	if err != nil {
		log.Printf("error -- upgrade error: %v\n", err)
		return
	}
	// return hhtp response
	// ctx.Response.Header.Set(key, value)
	// this will get the raw net.Conn

	if han.dbg {log.Printf("upgrade success; hijacking conn\n")}

	ctx.Hijack(hijackHandler)
	fmt.Fprintf(ctx, "Hijacked the connection!")
}


// hijackHandler is called on hijacked connection.
func hijackHandler(c net.Conn) {

	log.Printf("hello hijack handler\n")

/*
	hs, err := ws.Upgrade(c)
	if err != nil {
		fmt.Printf("upgrade error: %v\n", err)
		return
	}

	fmt.Printf("handshake: %v\n", hs)
//	fmt.Fprintf(c, "This message is sent over a hijacked connection to the client %s\n", c.RemoteAddr())
*/
	defer c.Close()

//	for {
		header, err := ws.ReadHeader(c)
		if err != nil {
					// handle error
			log.Printf("error -- read header: %v\n", err)
		}

		log.Printf("header [%d]: %v\n",header.Length, header)
		PrintWSHeader(header)
/*
		if header.OpCode == ws.OpClose {
			log.Printf("socket close\n")
//			break
		}
*/
		payload := make([]byte, header.Length)
		_, err = io.ReadFull(c, payload)
		if err != nil {
					// handle error
					// handle error
			log.Printf("error -- read full: %v\n", err)
		}
		if header.Masked {
			ws.Cipher(payload, header.Mask, 0)
		}
		log.Printf("payload: %s\n", payload)

		// Reset the Masked flag, server frames must not be masked as
		// RFC6455 says.
		header.Masked = false

		msg := []byte("hello client!")
		header.Length = int64(len(msg))

		if err := ws.WriteHeader(c, header); err != nil {
			log.Printf("error -- write header: %v\n", err)
		}
		if _, err := c.Write(msg); err != nil {
			log.Printf("error -- write payload: %v\n", err)
		}
		log.Printf("msg sent!")
//	}

}

func PrintWSHeader(h ws.Header) {

	fmt.Println("************* ws Header **************")
	fmt.Printf("Fin:    %t\n",h.Fin)
	fmt.Printf("Rsv:    %x\n",h.Rsv)
	fmt.Printf("OpC:    %x\n",h.OpCode)
	fmt.Printf("Masked: %t\n",h.Masked)
	fmt.Printf("Mask: 	%v\n",h.Mask)
	fmt.Printf("Length: %t\n",h.Fin)
	fmt.Println("*********** end ws Header ************")
}

