package embed

import (
	"fmt"
	"log"
	"bytes"
	"os"
)

type scrIns struct {
    filnam string
    st int
    end int
}


func EmbedScript(idxbyt []byte, wwwBase string) (out []byte, err error) {

	dbg:=false
//	idxLen := len(idx)

	res, err := parseScript(idxbyt)
	if err != nil {return out, fmt.Errorf("parseScript: %v", err)}

    if dbg {
        fmt.Println("************ setup ****************")
        for i:=0; i< len(res); i++ {
            fmt.Printf("  %d: -- [%d:%d] %s\n", i+1, res[i].st, res[i].end, res[i].filnam)
        }
        fmt.Println("********** end setup **************")
    }

    pt := res[0].st
    idxbyt[pt-1] = '>'
    idxbyt[pt] = '\n'
    pt++

    nidxbyt := make([]byte,1024*100)
    copy(nidxbyt, idxbyt[:pt])
    if dbg {fmt.Printf("nidx [%d]:\n%s\n", len(nidxbyt), string(nidxbyt))}

//  read azul lib
    liblen := 1024*100
    libbyt := make([]byte, liblen)


    for i:=0; i<len(res); i++ {
        libfilnam := wwwBase + "js/" + res[i].filnam
        if dbg {log.Printf("dbg -- fil[%d]: %s\n",i+1, libfilnam)}

        libfil, err := os.Open(libfilnam)
        if err != nil {log.Fatalf("error -- cannot open index file: %v\n", err)}

        nlib, err:= libfil.Read(libbyt)
        if err != nil {log.Fatalf("error -- reading lib file: %v\n",err)}
        if nlib == liblen {log.Fatalf("error -- liblen too small!\n")}
        libfil.Close()

        nidxbyt = append(nidxbyt,libbyt ...)
    }

    rem := []byte("\n</script>\n</html>\n")

    nidxbyt = append(nidxbyt, rem ...)

	return nidxbyt, nil
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
//  if scIdx<0 {return res, fmt.Errorf("no script")}
//  if scEnd<0 {return res, fmt.Errorf("no /script")}
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
