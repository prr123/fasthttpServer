// fasthttpUpgrader
// upgrader of fasthttp for websockets
//
// Author: prr, azul software
// Date 12 July 2024
// copyright (c) 2024 prr, azul software
//

package upgrader

import (
	"log"
	"fmt"
//    "unsafe"
    "crypto/sha1"
    "encoding/base64"
	"bytes"
//	"net"
//	"io"
//    "github.com/gobwas/ws"
    "github.com/valyala/fasthttp"
)

/*
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
*/
/* 
// web socket upgrade
GET /chat HTTP/1.1
Host: example.com:8000
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
Sec-WebSocket-Version: 13
*/

func Upgrade(ctx *fasthttp.RequestCtx, dbg bool)(err error) {


	upgVal:= ctx.Request.Header.Peek("Upgrade")
	conVal:= ctx.Request.Header.Peek("Connection")
	wsKeyVal:= ctx.Request.Header.Peek("Sec-WebSocket-Key")
	wsVerVal:= ctx.Request.Header.Peek("Sec-WebSocket-Version")

	if dbg {
		log.Printf("*** ws upgrade ***\n")
		fmt.Printf("Upgrade: %s\n", upgVal)
		fmt.Printf("Connection: %s\n", conVal)
		fmt.Printf("ws Key: %s\n", wsKeyVal)
		fmt.Printf("ws version: %s\n", wsVerVal)
	}

	// tests
	if !bytes.Equal(upgVal,[]byte("websocket")) {return fmt.Errorf("Upgrade Header Value not websocket")}
//	if !bytes.Equal(conVal,[]byte("Upgrade")) {return fmt.Errorf("Connection Header Value no Upgrade")}
	if idx := bytes.Index(conVal,[]byte("Upgrade")); idx>0 {return fmt.Errorf("Connection Header Value no Upgrade")}


	if !bytes.Equal(wsVerVal,[]byte("13")) {return fmt.Errorf("Sec-WebSocket-Version Header Value not 13")}
//	if len(wsKeyVal) != noncekeySize {return fmt.Errorf("Sec-WebSocket-Key Header Value length is not 13!")}

    acceptStr, err := GenNonce([]byte(wsKeyVal))
    if err != nil {return fmt.Errorf("cannot generate nonce accept string: %v", err)}

	if dbg {log.Printf("dbg -- key: %s\naccept: %s\n", wsKeyVal, acceptStr)}

	ctx.SetStatusCode(101)
//    ctx.SetContentType("text/plain; charset=utf8")
	ctx.Response.Header.Set("Upgrade", "websocket")
	ctx.Response.Header.Set("Connection", "Upgrade")
	ctx.Response.Header.Set("Sec-Websocket-Accept", string(acceptStr))
//	ctx.Response.Header.Set("Sec-WebSocket-Protocol", *)//	ctx.Response.Header.Set("Sec-WebSocket-Extensions", *)
//	ctx.SetBodyString("ws resp sent!\n")

	return nil
}


func GenNonce(nonce []byte)(accept []byte, err error) {

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

    accept = make([]byte, acceptSize)

    if len(accept) != acceptSize {
        return accept, fmt.Errorf("accept buffer is invalid")
    }
    if len(nonce) != nonceSize {
        return accept, fmt.Errorf("nonce is invalid")
    }
    p := make([]byte, nonceSize+len(magic))
    copy(p[:nonceSize], nonce)
    copy(p[nonceSize:], magic)

    sum := sha1.Sum(p)
    base64.StdEncoding.Encode(accept, sum[:])

    return accept, nil
}
