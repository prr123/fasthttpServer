// fastHttpServerV6
// building a webserver based on wasgob and fasthttp
//
// Author: prr, azul software
// Date 13 July 2024
// copyright (c) 2024 prr, azul software
//

// V2 add RTyp and expand handler type
// add router
// test websocket
//
// v3 create upgrader routine
//
// v5: test performance by preloading html
//     preload scripts
// V6: add binary transport
//

package main

import (
	"os"
	"log"
	"fmt"
	"bytes"
	"strings"
	"net"
	"io"
	"unsafe"

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
	dbgout bool
    router map[string] Rtyp
	p pparse.Path
	index *[]byte
	idxLen int
}

type scrIns struct {
	filnam string
	st int
	end int
}

func main() {

	wwwBase := "/home/peter/www/azuldist/"

    numarg := len(os.Args)
    flags:=[]string{"dbg", "port", "index", "out"}

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

	dbgout:= false
    _, ok = flagMap["out"]
    if ok {dbgout = true}

    portStr := ""
    pval, ok := flagMap["port"]
    if !ok {
        log.Fatalf(" error no port provided!\n")
    } else {
        if pval.(string) == "none" {log.Fatalf("error: no port value provided!\n")}
        portStr = pval.(string)
    }

    rootFil := "indexjsV5"
    rval, ok := flagMap["index"]
    if ok {
        if rval.(string) == "none" {log.Fatalf("error: no index file name provided!\n")}
        rootFil = rval.(string)
    }



	han := Handler{
		dbg: dbg,
		dbgout: dbgout,
	}
    han.router = make(map[string]Rtyp)

    idxFilnam := wwwBase + "html/" + rootFil + ".html"
	idxfil, err := os.Open(idxFilnam)
    if err != nil {log.Fatalf("error -- cannot open index file: %v\n", err)}
    idx := Rtyp {
        fil: idxfil,
        ftyp: "text/html",
    }
    han.router["index.html"] = idx

/*
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
*/

//	preload
	idxlen := 1024*200
	idxbyt := make([]byte, idxlen)
	n, err:= idxfil.Read(idxbyt)
	if err != nil {log.Fatalf("error -- reading idx file: %v\n",err)} 
	if n == idxlen {log.Fatalf("error -- idxlen too small!\n")}

	if dbg {fmt.Printf("dbg -- idx [%d]: %s\n%s\n", n, idxFilnam, string(idxbyt))}

	res, err := parseScript(idxbyt)
	if err != nil {log.Fatalf("error -- parsing idx file: %v\n",err)} 
//fmt.Printf("dbg -- len(res): %d\n", len(res))


	if dbg {
		fmt.Println("************ setup ****************")
		for i:=0; i< len(res); i++ {
			fmt.Printf("  %d: -- [%d:%d] %s\n", i+1, res[i].st, res[i].end, res[i].filnam)
		}
		fmt.Println("********** end setup **************")
	}

//	read azul lib
	liblen := 1024*100
	libbyt := make([]byte, liblen)


//	isize = n

	pt := res[0].st
	idxbyt[pt-1] = '>'
	idxbyt[pt] = '\n'
	pt++

	for i:=0; i<len(res); i++ {
	    libfilnam := wwwBase + "js/" + res[i].filnam
		if dbg {log.Printf("dbg -- fil[%d]: %s\n",i+1, libfilnam)}

		libfil, err := os.Open(libfilnam)
		if err != nil {log.Fatalf("error -- cannot open index file: %v\n", err)}

		nlib, err:= libfil.Read(libbyt)
		if err != nil {log.Fatalf("error -- reading lib file: %v\n",err)} 
		if nlib == liblen {log.Fatalf("error -- liblen too small!\n")}
    	libfil.Close()

		for j:=0; j<nlib; j++ {
			idxbyt[pt] = libbyt[j]
			pt++
		}
	}

	rem := []byte("\n</script>\n</html>\n")
	for i:=0; i<len(rem); i++ {
		idxbyt[pt] = rem[i]
		pt++
	}
	lidx := pt

	han.index = &idxbyt
	han.idxLen = lidx

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

func (han Handler)idxHandlerOld(ctx *fasthttp.RequestCtx) {

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

func (han Handler)idxHandler(ctx *fasthttp.RequestCtx) {

	if han.dbg {log.Printf("dbg index %s %s %s\n", han.p.Fold, han.p.Fnam, han.p.Ext)}

/*
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
*/
    ctx.SetContentType("text/html; charset=utf-8")
	out := *han.index
	if han.dbgout {fmt.Printf("dbg -- out\n%s\n",string(out[:han.idxLen]))}
	n, err := ctx.Write(out[:han.idxLen])
	if err != nil {log.Fatalf("error -- ctx write: %v", err)}
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

	var ival int32
	var msg []byte

	ival = -1
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

	log.Println("*** ws handler start ***")

	binary := false
	for it:=0; it< 10; it++ {
		header, err := ws.ReadHeader(c)
		if err != nil {log.Printf("ws error -- read header: %v\n", err)}

		log.Printf("ws header [%d]: %x\n",header.Length,header.OpCode)
//		PrintWSHeader(header)

		if header.OpCode == ws.OpClose {
			log.Printf("ws info -- close code\n")
			break
		}
			// change
		payload := make([]byte, header.Length)
		_, err = io.ReadFull(c, payload)
		if err != nil {log.Printf("ws error -- read full: %v\n", err)}
		if header.Masked {
			ws.Cipher(payload, header.Mask, 0)
		}
		if header.OpCode == 1 {
			binary = false
			log.Printf("ws text payload: %s\n", payload)
		}

		if header.OpCode == 2 {
			binary = true
			ival = ByteSliceToInt32(payload[:4])
			log.Printf("ws binary payload: %d\n", ival)
		}

		if it >3 {binary = true}
		// Reset the Masked flag, server frames must not be masked as
		// RFC6455 says.
		header.Masked = false
		if binary {
			ival++
			msg = Int32ToByteSlice(ival)
			fmt.Printf("msg: %v\n", msg)
			header.OpCode = 2
			header.Length = 4
		} else {
			tstStr := fmt.Sprintf("hello client [%d]!", it)
			msg = []byte(tstStr)
			header.Length = int64(len(msg))
			header.OpCode = 1
		}

		if it == 7 {
			header.OpCode = 1
			binary = false
			msg = []byte("end")
			header.Length = int64(len(msg))
		}

		if err := ws.WriteHeader(c, header); err != nil {
			log.Printf("error -- write header: %v\n", err)
		}
		if _, err := c.Write(msg); err != nil {
			log.Printf("error -- write payload: %v\n", err)
		}
		if binary {
			log.Printf("bin msg sent [%d]: %d!", header.Length, ival)
		} else {
			log.Printf("txt msg sent [%d]: %s!", header.Length, msg)
		}
	}


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


func parseScript(idx []byte)(res []scrIns, err error) {

	ist :=0
	scIdx :=-1
	scEnd := -1
	for i:=0; i< 10; i++ {
		scIdx = bytes.Index(idx[ist:],[]byte("<script "))
		if scIdx == -1 {break}
		nist := ist + scIdx + 8
		scEnd = bytes.Index(idx[nist:],[]byte("</script>"))
		ist = nist + scEnd + 9
//fmt.Printf("%d: %s\n",i,string(idx[nist:ist -9]))
		jsfilnam, err := parseScriptFil(idx[nist:(ist -9)])
		if err != nil {return res, fmt.Errorf("parsing script: %v", err)}
//fmt.Printf("js file name: %s\n", jsfilnam)
		scr := scrIns {
			st: nist,
			end: ist-9,
			filnam: jsfilnam,
		}

		res = append(res, scr)
	}
//	if scIdx<0 {return res, fmt.Errorf("no script")}
//	if scEnd<0 {return res, fmt.Errorf("no /script")}
	return res, nil
}

func parseScriptFil(x []byte)(out string, err error) {

	istate:=0
	ist:= -1
	iend := -1
	for i:=0; i< len(x)-1; i++ {
		switch istate {
		case 0:
			if x[i] == '\'' {
				istate = 1
				ist = i+1
			}

		case 1:
			if x[i] == '\'' {
				istate = 2
				iend = i
			}
		default:
		}
		if istate == 2 {break}
	}

	if ist<0 {return "", fmt.Errorf("no start apost")}
	if iend<0 {return "", fmt.Errorf("no end apost")}

	out = string(x[ist:iend])
	return out, nil
}

func toInt(bytes []byte) int {
    result := 0
    for i := 0; i < 4; i++ {
        result = result << 8
        result += int(bytes[i])

    }

    return result
}

func Int32ToByteSlice(num int32) []byte {
    size := int(unsafe.Sizeof(num))
    arr := make([]byte, size)
    for i := 0 ; i < size ; i++ {
        byt := *(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&num)) + uintptr(i)))
        arr[i] = byt
    }
    return arr
}

func ByteSliceToInt32(arr []byte) int32{
    val := int32(0)
    size := 4
    for i := 0 ; i < size ; i++ {
        *(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&val)) + uintptr(i))) = arr[i]
    }
    return val
}
