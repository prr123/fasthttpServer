// fastHttpsMap 
// from fastHttpsWSNV12
//
// building a webserver based on wasgob and fasthttp tls
//
// Author: prr, azul software
// Date 13 July 2024
// copyright (c) 2024 prr, azul software
//
// embeds script files
//
// binary transport
//
// V2:
// add js file handler
//
// V3:
// add json file handler
// fix js prefix for embedding
//
// V4:
// added preload flag
// only embed scripts with src attribute
// problem other scripts are not included if preload flag is set
// solution requires rework of embedding
//
// V5:
// rework embedding
// check for psrc=
// if not embed psrc => src
// else embed
//
// v6
//
// V7:
// add post method to json
// explore using hash functions on router
// add json converter
// V8:
// V9: process cmds
// V10: added imgHandler
//      fmt -> log
// V11: add authorisation -Benderson
//
// V12: add authorisation with gologin
//		- first plain https
// 		- add domain
//
// map: replace switch with a map of funcs
//
// sni: server name indication
//

package main

import (
	"os"
	"log"
	"fmt"
	"bytes"
	"strings"
	"net"
//	"net/http"
	"io"
//	"io/ioutil"
	"unsafe"
//	"context"
//	"encoding/base64"

	"crypto/tls"
//	"crypto/rand"
//	"time"

	"github.com/prr123/fasthttpServer/fasthttp/upgrader"
	"github.com/prr123/fasthttpServer/fasthttp/pathparser"

//	"golang.org/x/oauth2"
//	"golang.org/x/oauth2/google"

	"github.com/gobwas/ws"
	"github.com/valyala/fasthttp"
	"github.com/goccy/go-json"

    util "github.com/prr123/utility/utilLib"
)

type address struct {
	Street string
	StNum string
	AptNum string
	City string
	Zip string
	Country string
}

type Rtyp struct {
    fil *os.File
    ftyp string
    }

type Handler struct {
	dbg bool
	test bool
    router map[string] Rtyp
	p pparse.Path
	index *[]byte
	idxLen int
	oaState string
	domain string
}

type scrIns struct {
	filnam string
	st int
	end int
	src int
}

/*
var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8000/auth/google/callback",
	ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

const GoogleClientID = "id"
const GoogleClientSecret = "secret"
*/

const wwwBase = "/home/peter/www/azuldist/"

var hmap = make(map[string]func(h Handler, ctx *fasthttp.RequestCtx))

var CertList []tls.Certificate
var	CertMap = make(map[string]tls.Certificate)


func main() {

    numarg := len(os.Args)
    flags:=[]string{"dbg", "domain", "port", "index", "test" ,"preload"}

    useStr := " /domain = domstr /port=portstr [/index=idxfil] [/preload] [/dbg]"
    helpStr := "fasthttps server"

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

	test:= false
    _, ok = flagMap["test"]
    if ok {test = true}

	preload:= false
    _, ok = flagMap["preload"]
    if ok {preload = true}

    domStr := ""
    domval, ok := flagMap["domain"]
    if !ok {
        log.Fatalf(" error -- no domain flag!\n")
    } else {
        if domval.(string) == "none" {log.Fatalf("error: no domain provided!\n")}
        domStr = domval.(string)
    }

	domList := strings.Split(domStr, ",")
	ndomList := make([]string, len(domList))
	for i:=0; i<len(domList); i++ {
		d := []byte(domList[i])
		for j:=0; j< len(d); j++ {
			if d[j] == '_' {d[j] = '.'}
		}
		ndomList[i] = string(d)
	}

    portStr := ""
    pval, ok := flagMap["port"]
    if !ok {
        log.Fatalf(" error port flag!\n")
    } else {
        if pval.(string) == "none" {log.Fatalf("error: no port provided!\n")}
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
		test: test,
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

	// check domain
/*
	certDir := "/home/peter/cloud/LEAccount/domains/" + domStr 
	 _, err = os.Stat(certDir)
    if os.IsNotExist(err) {
        log.Fatalf("error -- domain directory %s does not exist\n", domStr)
    } else if err != nil {
        log.Fatalf("error -- checking directory: %v\n", err)
    }
*/
	certDir := "/home/peter/cloud/LEAccount/certs/"
	 _, err = os.Stat(certDir)
    if os.IsNotExist(err) {
        log.Fatalf("error -- cert directory %s does not exist\n", certDir)
    } else if err != nil {
        log.Fatalf("error -- checking directory: %v\n", err)
    }

	fmt.Println("************ domains ***********")
	for i:=0; i< len(domList); i++ {
		fmt.Printf(" --%d: %s %s\n", i, domList[i], ndomList[i])
	}
	fmt.Println("********** end domains *********")

	//check whether all certs exist
	CertList = make([]tls.Certificate, len(domList))

	for i:=0; i< len(domList); i++ {
		dom := domList[i]
		ndom := ndomList[i]
		if dbg {log.Printf(" --%d: %s\n", i, domList[i])}
		cert := certDir + "/" + dom + ".crt"
		key := certDir + "/" + dom + ".key"
		_,err := os.Stat(cert)
		if err != nil {log.Fatalf("error -- not found cert %s! %v\n", cert, err)}
		_,err = os.Stat(key)
		if err != nil {log.Fatalf("error -- not found key %s! %v\n", key, err)}
		certs, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {log.Fatalf("error -- loading certs: %v\n", err)}
		CertList[i] = certs
		CertMap[ndom] = certs
	}


	if dbg {
		log.Printf("dbg -- testing CertMap\n")
		for i:=0; i< len(domList); i++ {
			ndom := ndomList[i]
			_, ok = CertMap[ndom]
			log.Printf(" --%d: %s found: %t\n", i, ndomList[i], ok)
		}
	}


	//	preload
	idxlen := 1024*100
	idxbyt := make([]byte, idxlen)
	n, err:= idxfil.Read(idxbyt)
	if err != nil {log.Fatalf("error -- reading idx file: %v\n",err)}
	if n == idxlen {log.Fatalf("error -- idxlen too small!\n")}

	if dbg {fmt.Printf("dbg -- idx [%d]: %s\n%s\n", n, idxFilnam, string(idxbyt))}
	han.index = &idxbyt
	idxlen = n
	han.idxLen = n

	res, err := parseScript(idxbyt)
	if err != nil {log.Fatalf("error -- parsing idx file: %v\n",err)} 

//fmt.Printf("dbg -- len(res): %d\n", len(res))

	if dbg {
		fmt.Println("************ setup ****************")
		for i:=0; i< len(res); i++ {
			fmt.Printf("  %d: -- [%d:%d] src: %d %s\n", i+1, res[i].st, res[i].end, res[i].src, res[i].filnam)
		}
		fmt.Println("********** end setup **************")
	}

	if preload {

		pre := han.idxLen
		if len(res) > 0 {pre = res[0].st -8}
//fmt.Printf("preamble: %d from %d\n", pre, han.idxLen)
		//create new index file
		nlib := 200
		nidxbyt := make([]byte, 1024*nlib)

		for i:= 0; i< pre; i++ {
			nidxbyt[i] = idxbyt[i]
		}
//fmt.Printf("preamble:\n%s\n", nidxbyt[:pre])

//	read azul lib
		liblen := 1024*100
		libbyt := make([]byte, liblen)

		pt := pre +1
		for i:=0; i<len(res); i++ {
			if res[i].src == -1 {
				for j:=res[i].st -8; j< res[i].end + 10; j++ {
					nidxbyt[pt] = idxbyt[j]
					pt++
				}

			} else {

	    		libfilnam := wwwBase + res[i].filnam
				if dbg {log.Printf("dbg -- fil[%d]: %s\n",i+1, libfilnam)}

				libfil, err := os.Open(libfilnam)
				if err != nil {log.Fatalf("error -- cannot open index file: %v\n", err)}

				nlib, err:= libfil.Read(libbyt)
				if err != nil {log.Fatalf("error -- reading lib file: %v\n",err)} 
				if nlib == liblen {log.Fatalf("error -- liblen too small!\n")}
    			libfil.Close()

				scrSt := []byte("<script>\n")
				for j:=0; j<len(scrSt); j++ {
					nidxbyt[pt] = scrSt[j]
					pt++
				}
				for j:=0; j<nlib; j++ {
					nidxbyt[pt] = libbyt[j]
					pt++
				}
				scrEnd := []byte("</script>\n")
				for j:=0; j<len(scrEnd); j++ {
					nidxbyt[pt] = scrEnd[j]
					pt++
				}

			}
//fmt.Printf("res %d:\n%s\n", i, nidxbyt[:pre])

		}
		rem := res[len(res)-1].end + 9
		for j:=rem; j<idxlen; j++ {
			nidxbyt[pt] = idxbyt[j]
			pt++
		}

		lenidx := pt
		han.index = &nidxbyt
		han.idxLen = lenidx

	} else {
		// convert psrc => src
		han.index = &idxbyt
		han.idxLen = n
		for i:=0; i<len(res); i++ {
			chp := res[i].src
			if chp > -1 {idxbyt[chp] = ' '}
		}
	}

	hmap["/"] = (Handler).idxHandler
	hmap["/js/"]	= (Handler).jsHandler
	hmap["/js/md/"]	= (Handler).jsHandler
	hmap["/pdf/"]	= (Handler).pdfHandler
	hmap["/hijack"]	= (Handler).wsHandler
	hmap["/hello"]	= (Handler).helloHandler

	log.Printf("info -- starting to listen at %s at port: %s\n", domStr, portStr)


	tlsconfig := &tls.Config{
		Certificates: CertList,
//		GetCertificate: 
		GetConfigForClient: GetConfigForClientHandler,
	}

	ln, err := tls.Listen("tcp", ":"+portStr , tlsconfig)
	if err != nil {log.Fatalf("error -- creating listener: %v\n", err)}
	defer ln.Close()

//	serverMain := &fasthttp.Server {Handler: han.requestHandler}
//	serverAdmin := &fasthttp.Server {Handler: han.reqHandlerAdmin}

	err = fasthttp.Serve(ln, han.requestHandler);
	if err != nil {log.Fatalf("error in Serve: %v\n", err)}
}
//	err = fasthttp.ListenAndServeTLS(":"+portStr, cert, key, han.requestHandler)
//	if err != nil {log.Printf("error Listen: %v\n", err)}


func GetConfigForClientHandler(info *tls.ClientHelloInfo) (*tls.Config, error) {
    log.Printf("GetConfigForClientHandler for: %s\n", info.ServerName)

	tlsconfig :=  &tls.Config{}
	cert, ok := CertMap[info.ServerName]
	if !ok {
		log.Printf("error -- no cert for %s\n", info.ServerName)
		tlsconfig.Certificates = CertList
	} else {
		log.Printf("info -- found cert for %s\n", info.ServerName)
		tlsconfig.Certificates = append(tlsconfig.Certificates, cert)
	}

	return tlsconfig, nil
}



	// the corresponding fasthttp request handler
func (han Handler)requestHandler(ctx *fasthttp.RequestCtx) {

	if han.dbg {log.Printf("dbg -- ctx main: %q method: %q path: %q\n", ctx.RequestURI(), ctx.Method(), ctx.Path())}

	// find etension and folder path -> parse ctx.PATH

	p := pparse.Pparse(ctx.Path())
	if han.dbg {log.Printf("dbg Fold: %s Fnam: %s Ext: %s\n", p.Fold, p.Fnam, p.Ext)}
	han.p = p

	fn, ok := hmap[string(p.Fold)]
	if !ok {
		if han.dbg {log.Println("unsupported path!")}
		ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		return
	}

	fn(han, ctx)
}

func (han Handler)reqHandlerAdmin(ctx *fasthttp.RequestCtx) {
	if han.dbg {log.Printf("dbg -- ctx admin: %q method: %q path: %q\n", ctx.RequestURI(), ctx.Method(), ctx.Path())}


}


func (han Handler)helloHandler(ctx *fasthttp.RequestCtx) {
	log.Printf("helloHandler\n")
//	chi := ctx.ClientHelloInfo()

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

}

/*
	// Set arbitrary headers
	ctx.Response.Header.Set("X-My-Header", "my-header-value")

	// Set cookies
	var c fasthttp.Cookie
	c.SetKey("cookie-name")
	c.SetValue("cookie-value")
*/


func (han Handler)idxHandler(ctx *fasthttp.RequestCtx) {

	if han.dbg {log.Printf("dbg -- index %s %s %s\n", han.p.Fold, han.p.Fnam, han.p.Ext)}

    ctx.SetContentType("text/html; charset=utf-8")
	out := *han.index
//	if han.test {fmt.Printf("dbg -- out\n%s\n",string(out[:han.idxLen]))}
	n, err := ctx.Write(out[:han.idxLen])
	if err != nil {log.Fatalf("error -- ctx write: %v", err)}
    if han.dbg {fmt.Printf("dbg index -- sent: %d\n", n)}
}

/*
func (han Handler) loginHandler(ctx *fasthttp.RequestCtx) {
	if han.dbg {log.Printf("dbg loginHandler -- enter\n")}

	// Create oauthState cookie
	oauthState := generateStateOauthCookie(ctx)

	var expiration = time.Now().Add(365 * 24 * time.Hour)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	oacookie := fasthttp.Cookie{}
	oacookie.SetKey("oauthstate")
	oacookie.SetValue(state)
	oacookie.SetMaxAge(3600000)
	oacookie.SetDomain("localhost")
	oacookie.SetPath(("/"))
	oacookie.SetSecure(false)

	ctx.Response.Header.SetCookie(&oacookie)

	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	han.oaState = state

	
//	AuthCodeURL receive state that is a token to protect the user from CSRF attacks. You must always provide a non-empty string and
//	validate that it matches the the state query parameter on your redirect callback.

	u := googleOauthConfig.AuthCodeURL(state)

//	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
//	ctx.Redirect(u,307) // 307 temporary redirect

}

func (han Handler)googleCBHandler(ctx *fasthttp.RequestCtx) {
	if han.dbg {log.Printf("dbg idHandler -- enter\n")}

	bodyToken := ctx.FormValue("oauthstate")
	if han.dbg {log.Printf("dbg -- body: %s\n",string(bodyToken))}

	// Read oauthState from Cookie
	cookieToken := ctx.Request.Header.Cookie("oauthstate")
	if han.dbg {log.Printf("dbg -- body: %s\n",string(cookieToken))}

    if bytes.Equal(cookieToken, bodyToken) {
//        http.Error(w, "token mismatch", http.StatusBadRequest)
        log.Printf("error idHandler -- invalid id: %s - %s\n", bodyToken, cookieToken)
        ctx.SetStatusCode(400) //badrequest
        ctx.SetContentType("text/plain; charset=utf-8")
        fmt.Fprintf(ctx, "bad request token mismatch: %s %s\n",bodyToken, cookieToken)
        return
    }

	code := ctx.FormValue("code")

	token, err := googleOauthConfig.Exchange(context.Background(), string(code))
    if err != nil {
        log.Printf("error OAuthExchange: %v\n", err)
        ctx.SetStatusCode(400) //badrequest
        ctx.SetContentType("text/plain; charset=utf-8")
        fmt.Fprintf(ctx, "exchange error: %v\n", err)
        return
    }


	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
        log.Printf("error http.Get: %v\n", err)
        ctx.SetStatusCode(400) //badrequest
        ctx.SetContentType("text/plain; charset=utf-8")
        fmt.Fprintf(ctx, "error -- http.Get: %v\n", err)
        return
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)

	if err != nil {
        log.Printf("error read resp: %v\n", err)
        ctx.SetStatusCode(400) //badrequest
        ctx.SetContentType("text/plain; charset=utf-8")
        fmt.Fprintf(ctx, "response read error: %v\n", err)
        return
	}

	fmt.Fprintf(ctx, "UserInfo: %s\n", contents)

}
*/

func (han Handler)jsHandler(ctx *fasthttp.RequestCtx) {

	if han.dbg {log.Printf("dbg -- index %s %s %s\n", han.p.Fold, han.p.Fnam, han.p.Ext)}

	if !bytes.Equal(han.p.Ext, []byte("js"))  {
        log.Printf("error jsHandler -- invalid req: %s\n", ctx.Path)
        ctx.SetStatusCode(404)
        ctx.SetContentType("text/plain; charset=utf-8")
        fmt.Fprintf(ctx, "invalid req: %s\n",ctx.Path)
        return
	}

    ctx.SetContentType("application/javascript; charset=utf-8")

	filnam := wwwBase + string(ctx.Path())
	if han.dbg {log.Printf("dbg -- info jsHandler -- filnam: %s\n",filnam)}

	fil, err := os.Open(filnam)
	if err != nil {
		log.Printf("error -- jsHandler -- ctx open: %v", err)
	    ctx.SetStatusCode(404)
        ctx.SetContentType("text/plain; charset=utf-8")
        fmt.Fprintf(ctx, "could not read file: %s\n",ctx.Path)
		return
	}

    fil.Seek(0,0)
    n, err := io.Copy(ctx, fil)
	if err != nil {
		log.Printf("error -- jsHandler -- ctx copy: %v", err)
	    ctx.SetStatusCode(404)
        ctx.SetContentType("text/plain; charset=utf-8")
        fmt.Fprintf(ctx, "could not read file: %s\n",ctx.Path)
		return
	}

	if han.dbg {log.Printf("dbg info -- jsHandler -- sent %d \n", n)}
//	n, err := ctx.Write(out[:han.idxLen])
//	if err != nil {log.Fatalf("error jsHandler -- ctx write: %v", err)}

}

func (han Handler)imgHandler(ctx *fasthttp.RequestCtx) {

	if han.dbg {log.Printf("dbg imgHandler -- method: %q index %s %s %s\n", ctx.Method(), han.p.Fold, han.p.Fnam, han.p.Ext)}


	switch string(han.p.Ext) {
	case "png": 
    	ctx.SetContentType("image/png")
	case "jpg", "jpeg":
    	ctx.SetContentType("image/jpeg")
	case "gif":
    	ctx.SetContentType("image/gif")

	default:
        log.Printf("error imgHandler -- invalid type: %s\n", han.p.Ext)
        ctx.SetStatusCode(404)
        ctx.SetContentType("text/plain; charset=utf-8")
        fmt.Fprintf(ctx, "invalid req: %s\n",ctx.Path)
        return
	}

	filnam := wwwBase + string(ctx.Path())
	if han.dbg {log.Printf("dbg info imgHandler -- filnam: %s\n",filnam)}

	fil, err := os.Open(filnam)
	if err != nil {
		log.Printf("error imgHandler -- open file: %v", err)
	    ctx.SetStatusCode(404)
        ctx.SetContentType("text/plain; charset=utf-8")
        fmt.Fprintf(ctx, "could not read file: %s\n",ctx.Path)
		return
	}
	info, err := fil.Stat()
	if err != nil {
		log.Printf("error imgHandler -- file Stat: %v", err)
	    ctx.SetStatusCode(404)
        ctx.SetContentType("text/plain; charset=utf-8")
        fmt.Fprintf(ctx, "could not get size of file: %s\n",ctx.Path)
		return
	}

	ctx.Response.Header.SetContentLength(int(info.Size()))
    fil.Seek(0,0)
    n, err := io.Copy(ctx, fil)
	if err != nil {
		log.Printf("error imgHandler -- ctx copy: %v", err)
	    ctx.SetStatusCode(404)
        ctx.SetContentType("text/plain; charset=utf-8")
        fmt.Fprintf(ctx, "could not copy file: %s\n",ctx.Path)
		return
	}

	if han.dbg {log.Printf("dbg info imgHandler -- sent %d \n", n)}
	return
}

func (han Handler)jsonHandler(ctx *fasthttp.RequestCtx) {

	if han.dbg {log.Printf("dbg jsonHandler -- method: %q index %s %s %s\n", ctx.Method(), han.p.Fold, han.p.Fnam, han.p.Ext)}

	if !bytes.Equal(han.p.Ext, []byte("json"))  {
        fmt.Printf("error jsonHandler -- invalid req: %s\n", ctx.Path)
        ctx.SetStatusCode(404)
        ctx.SetContentType("text/plain; charset=utf-8")
        fmt.Fprintf(ctx, "invalid req: %s\n",ctx.Path)
        return
	}

    ctx.SetContentType("application/json; charset=utf-8")

	filnam := wwwBase + string(ctx.Path())
	if han.dbg {fmt.Printf("info jsonHandler -- filnam: %s\n",filnam)}

	fil, err := os.Open(filnam)
	if err != nil {
		log.Printf("error jsonHandler -- open file: %v", err)
	    ctx.SetStatusCode(404)
        ctx.SetContentType("text/plain; charset=utf-8")
        fmt.Fprintf(ctx, "could not read file: %s\n",ctx.Path)
		return
	}
    fil.Seek(0,0)
    n, err := io.Copy(ctx, fil)
	if err != nil {
		log.Printf("error jsonHandler -- ctx copy: %v", err)
	    ctx.SetStatusCode(404)
        ctx.SetContentType("text/plain; charset=utf-8")
        fmt.Fprintf(ctx, "could not read file: %s\n",ctx.Path)
		return
	}

	if han.dbg {log.Printf("info jsonHandler -- sent %d \n", n)}

}

func (han Handler)xjsonHandler(ctx *fasthttp.RequestCtx) {

	var adL []address

	if han.dbg {log.Printf("dbg xjson -- method: %q index %s %s %s\n", ctx.Method(), han.p.Fold, han.p.Fnam, han.p.Ext)}

	if !bytes.Equal(ctx.Method(), []byte("POST")) {
		log.Printf("error jsonHandler -- method not post!\n")
		return
	}

	res := ctx.PostBody()
	log.Printf("rec: %s\n",string(res))

 	err := json.Unmarshal(res, &adL)
	if err != nil {
		log.Printf("error jsonHandler -- conversion: %v!\n", err)
		return
	}
	PrintADL(adL)
}

func PrintADL (adl []address) {

	fmt.Printf("******** address records: %d *********\n",len(adl))
	for i:=0; i< len(adl); i++ {
		fmt.Printf("Street:  %s\n",adl[i].Street)
		fmt.Printf("Number:  %s\n",adl[i].StNum)
		fmt.Printf("Apt:     %s\n",adl[i].AptNum)
		fmt.Printf("City:    %s\n",adl[i].City)
		fmt.Printf("Zip:     %s\n",adl[i].Zip)
		fmt.Printf("Country: %s\n",adl[i].Country)

	}
	fmt.Printf("******** end address records *********\n")
}

func (han Handler)pdfHandler(ctx *fasthttp.RequestCtx) {

	if han.dbg {log.Printf("dbg index %s %s %s\n", han.p.Fold, han.p.Fnam, han.p.Ext)}

	if !bytes.Equal(han.p.Ext, []byte("pdf"))  {
        fmt.Printf("error pdfHandler -- invalid req: %s\n", ctx.Path)
        ctx.SetStatusCode(404)
        ctx.SetContentType("text/plain; charset=utf-8")
        fmt.Fprintf(ctx, "invalid req: %s\n",ctx.Path)
        return
	}

    ctx.SetContentType("application/pdf; charset=utf-8")

	filnam := wwwBase + string(ctx.Path())
	if han.dbg {fmt.Printf("info pdfHandler -- filnam: %s\n",filnam)}

	fil, err := os.Open(filnam)
	if err != nil {
		log.Printf("error pdfHandler -- open file: %v", err)
	    ctx.SetStatusCode(404)
        ctx.SetContentType("text/plain; charset=utf-8")
        fmt.Fprintf(ctx, "could not read file: %s\n",ctx.Path)
		return
	}
    fil.Seek(0,0)
    n, err := io.Copy(ctx, fil)
	if err != nil {
		log.Printf("error pdfHandler -- ctx copy: %v", err)
	    ctx.SetStatusCode(404)
        ctx.SetContentType("text/plain; charset=utf-8")
        fmt.Fprintf(ctx, "could not read file: %s\n",ctx.Path)
		return
	}

	if han.dbg {log.Printf("info pdfHandler -- sent %d \n", n)}
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

	defer c.Close()

	log.Println("*** ws handler start ***")

	binary := false
	ival = 0
	for it:=0; it< 10; it++ {
		header, err := ws.ReadHeader(c)
		if err != nil {log.Printf("ws error -- read header: %v\n", err)}

		log.Printf("<ws rec msg header [%d]: %x\n",header.Length,header.OpCode)
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
			log.Printf("<ws rec text payload [%d]: >%s<\n", header.Length, payload)
		}

		if header.OpCode == 2 {
			binary = true
// x
			inval := BArToInt32(payload[:4])
//			ival = ByteSliceToInt32(payload[:4])
			log.Printf("<ws rec binary payload [%d]: >%d<\n", header.Length, inval)
//			ival = inval
		}

		if it >3 {
			binary = true
		}
		// Reset the Masked flag, server frames must not be masked as
		// RFC6455 says.
		header.Masked = false
		if binary {
//			msg = Int32ToByteSlice(ival)
			ar := Int32ToBAr(ival)
			msg = ar[:]
			fmt.Printf("msg: %v\n", msg)
			header.OpCode = 2
			header.Length = 4
		} else {
			tstStr := fmt.Sprintf("hello client %d!", it)
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
			log.Printf(">bin msg sent [%d]: %d!", header.Length, ival)
			ival++
		} else {
			log.Printf(">txt msg sent [%d]: >%s<", header.Length, msg)
		}
	}


}

func PrintCtx(ctx *fasthttp.RequestCtx) {

	fmt.Println("******************** CTX request ******************")
	fmt.Printf("RequestURI is %q! Method %q\n", ctx.RequestURI(), ctx.Method())
	fmt.Printf("Requested path is %q\n", ctx.Path())
	fmt.Printf("Host is %q\n", ctx.Host())
	fmt.Printf("Query string is %q\n", ctx.QueryArgs())
	fmt.Printf("User-Agent is %q\n", ctx.UserAgent())
	fmt.Printf("Connection has been established at %s\n", ctx.ConnTime())
	fmt.Printf("Request has been started at %s\n", ctx.Time())
	fmt.Printf("Serial request number for the current connection is %d\n", ctx.ConnRequestNum())
//	fmt.Printf("Your source adr is %q\n", ctx.RemoteAddr.String())
	con:= ctx.Conn()
	adr := con.RemoteAddr()
	fmt.Printf("remote addr: %q \n", adr.String())
	idx := strings.Index(adr.String(), ":")
	port := adr.String()[idx+1:]
	fmt.Printf("port: %s\n", port)
	// unique id
	fmt.Printf("connection seq: %d\n",  ctx.ConnRequestNum())
	if ctx.ConnRequestNum() == 1 {fmt.Printf("need to login!\n")}

	fmt.Printf("connection id: %d\n\n", ctx.ConnID())

	authVal:= ctx.Request.Header.Peek("Authorization")
	fmt.Printf("auth value: %s\n", authVal)

	numHeaders := ctx.Request.Header.Len()
	head := ctx.Request.Header.RawHeaders()
	fmt.Printf("headers [%d]:\n%s\n", numHeaders, head)
	fmt.Printf("end headers\n")
	fmt.Printf("\nRaw request is:\n---START---\n%s\n---END---", &ctx.Request)

	fmt.Println("****************** End CTX request ****************")

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
		fmt.Printf("check src:  %s\n", idx[nist:nist+scEnd])
		jsfilnam, err := parseScriptFil(idx[nist:(ist -9)])
		if err != nil {return res, fmt.Errorf("parsing script: %v", err)}
//fmt.Printf("js file name: %s\n", jsfilnam)
		srcIdx := bytes.Index(idx[nist:nist+scEnd], []byte("psrc="))
		if srcIdx>-1 {srcIdx=nist+ srcIdx}
		scr := scrIns {
			st: nist,
			end: ist-9,
			filnam: jsfilnam,
			src: srcIdx,
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

func Int32ToBAr(x int32) (ar [4]byte) {
	ar = *(*[4]byte)(unsafe.Pointer(&x))
	return ar
}

func BArToInt32(ar []byte) (x int32) {
	x = *(*int32)(unsafe.Pointer(&ar[0]))
	return x
}
// (*type)(unsafe.Pointer()) casts a pointer into a pointer to type 
