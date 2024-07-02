// fastHttpServerWsV1
// building a webserver based on wasgob and fasthttp
//
// Author: prr, azul software
// Date 29 June 2024
// copyright (c) 2024 prr, azul software
//


package main

import (
	"os"
	"log"
	"fmt"
	"strings"
	"net"
	"bufio"
 	"unsafe"
   "crypto/sha1"
    "encoding/base64"

	"server/http/fasthttp/wsnonce"
	"github.com/gobwas/ws"
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
		case "/hijack":
			// Note that the connection is hijacked only after
			// returning from requestHandler and sending http response.
			han.wsHandler(ctx)

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
		upgVal:= ctx.Request.Header.Peek("Upgrade")
		fmt.Printf("upgrade value: %s\n", upgVal)
		conVal:= ctx.Request.Header.Peek("Connection")
		fmt.Printf("connection value: %s\n", conVal)
		wsKeyVal:= ctx.Request.Header.Peek("Sec-WebSocket-Key")
		fmt.Printf("ws Key value: %s\n", wsKeyVal)
		wsVerVal:= ctx.Request.Header.Peek("Sec-WebSocket-Version")
		fmt.Printf("ws version: %s\n", wsVerVal)
	}

	// return hhtp response
	// ctx.Response.Header.Set(key, value)
/*
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
*/
	wsKeyVal:= ctx.Request.Header.Peek("Sec-WebSocket-Key")

    acceptStr, err := wsnonce.GenNonce([]byte(wsKeyVal))
    if err != nil {log.Fatalf("cannot gen accept: %v", err)}
	if han.dbg {
		fmt.Printf("key: %s\naccept: %s\n", wsKeyVal, acceptStr)
	}
	ctx.SetStatusCode(101)
//    ctx.SetContentType("text/plain; charset=utf8")
	ctx.Response.Header.Set("Upgrade", "websocket")
	ctx.Response.Header.Set("Connection", "Upgrade")
	ctx.Response.Header.Set("Sec-Websocket-Accept", string(acceptStr))
//	ctx.Response.Header.Set("Sec-WebSocket-Protocol", *)
//	ctx.Response.Header.Set("Sec-WebSocket-Extensions", *)

	ctx.SetBodyString("ws resp sent!\n")

	// The connection will be hijacked after sending this response.
//	fmt.Fprintf(ctx, "Hijacked the connection!")
	// this will get the raw net.Conn
//	ctx.Hijack(hijackHandler)

}

// hijackHandler is called on hijacked connection.
func hijackHandler(c net.Conn) {

	fmt.Printf("hello hijack handler\n")

	hs, err := ws.Upgrade(c)
	if err != nil {
		fmt.Printf("upgrade error: %v\n", err)
		return
	}

	fmt.Printf("handshake: %v\n", hs)
//	fmt.Fprintf(c, "This message is sent over a hijacked connection to the client %s\n", c.RemoteAddr())

/*
	fmt.Fprintf(c, "Send me something and I'll echo it to you\n")
	var buf [1]byte
	for {
		if _, err := c.Read(buf[:]); err != nil {
			log.Printf("error when reading from hijacked connection: %v", err)
			return
		}
		fmt.Fprintf(c, "You sent me %q. Waiting for new data\n", buf[:])
	}
*/

}


func initAcceptFromNonce(accept, nonce []byte) {
	const (
	magic = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	// RFC6455: The value of this header field MUST be a nonce consisting of a
	// randomly selected 16-byte value that has been base64-encoded (see
	// Section 4 of [RFC4648]).  The nonce MUST be selected randomly for each
	// connection.
	nonceKeySize = 16
	nonceSize    = 24 // base64.StdEncoding.EncodedLen(nonceKeySize)

	// RFC6455: The value of this header field is constructed by concatenating
	// /key/, defined above in step 4 in Section 4.2.2, with the string
	// "258EAFA5- E914-47DA-95CA-C5AB0DC85B11", taking the SHA-1 hash of this
	// concatenated value to obtain a 20-byte value and base64- encoding (see
	// Section 4 of [RFC4648]) this 20-byte hash.
	acceptSize = 28 // base64.StdEncoding.EncodedLen(sha1.Size)
	)

	if len(accept) != acceptSize {
		panic("accept buffer is invalid")
	}
	if len(nonce) != nonceSize {
		panic("nonce is invalid")
	}

	p := make([]byte, nonceSize+len(magic))
	copy(p[:nonceSize], nonce)
	copy(p[nonceSize:], magic)

	sum := sha1.Sum(p)
	base64.StdEncoding.Encode(accept, sum[:])
}

func writeAccept(bw *bufio.Writer, nonce []byte) (int, error) {
	// temporary solution
	const acceptSize = 28
	accept := make([]byte, acceptSize)
	initAcceptFromNonce(accept, nonce)
	// NOTE: write accept bytes as a string to prevent heap allocation –
	// WriteString() copy given string into its inner buffer, unlike Write()
	// which may write p directly to the underlying io.Writer – which in turn
	// will lead to p escape.
//	return bw.WriteString(btsToString(accept))
	return bw.WriteString(string(accept))
}

func b2s(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

