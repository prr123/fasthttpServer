// embed
// program that reads an index file and embeds the src files
//
// Author: prr, azul software
// Date 23 July 2024
// copyright (c) 2024 prr, azul software
//


package main


import (
    "os"
    "log"
    "fmt"
    "bytes"

    util "github.com/prr123/utility/utilLib"
)

type scrIns struct {
    filnam string
    st int
    end int
}

func main() {

    wwwBase := "/home/peter/www/azuldist/"

    numarg := len(os.Args)
    flags:=[]string{"dbg", "index", "out"}

    useStr := "[/index=idxfil] [/out=outstr] [/dbg]"
    helpStr := "embed"

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

    idxBase := "indexjsV5"
    ival, ok := flagMap["index"]
    if ok {
        if ival.(string) == "none" {log.Fatalf("error: no index file name provided!\n")}
        idxBase = ival.(string)
    }

    outBase := idxBase + "_embed"
    outval, ok := flagMap["out"]
    if ok {
        if outval.(string) == "none" {log.Fatalf("error: no out value provided!\n")}
        outBase = outval.(string)
    }

    idxFilnam := wwwBase + "html/" + idxBase + ".html"
	outFilnam := wwwBase + "html/" + outBase + ".html"

	if dbg {
		fmt.Println("*********** cli ***********")
		fmt.Printf("index File:  %s->%s\n", idxBase, idxFilnam)
		fmt.Printf("output File: %s->%s\n", outBase, outFilnam)
		fmt.Println("********* end cli *********")
	}

//  preload
    idxfil, err := os.Open(idxFilnam)
    if err != nil {log.Fatalf("error -- cannot open index file: %v\n", err)}


	idxlen := 1024*200
    idxbyt := make([]byte, idxlen)
    n, err:= idxfil.Read(idxbyt)
    if err != nil {log.Fatalf("error -- reading idx file: %v\n",err)}
    if n == idxlen {log.Fatalf("error -- idxlen too small!\n")}

    if dbg {fmt.Printf("dbg -- idx [%d]: %s\n%s\n", n, idxFilnam, string(idxbyt))}

    res, err := parseScript(idxbyt)
    if err != nil {log.Fatalf("error -- parsing idx file: %v\n",err)}

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
//    lidx := pt
	// save file

	if dbg {fmt.Printf("out [%d]:\n%s\n", len(nidxbyt), string(nidxbyt))}

	err = os.WriteFile(outFilnam, nidxbyt, 0644)
	if err != nil {log.Fatalf("error -- could not write file: %v\n", err)}

	log.Println("success")
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
