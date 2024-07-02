package wsnonce

import (

//    "bufio"
	"fmt"
    "unsafe"
	"crypto/sha1"
    "encoding/base64"

)

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

func b2s(b []byte) string {
    return unsafe.String(unsafe.SliceData(b), len(b))
}
