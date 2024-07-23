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

	embed "server/http/fasthttp/embedLib"
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

    idxBase := "indexV6"
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

	nidxbyt, err := embed.EmbedScript(idxbyt[:n], wwwBase)
	if err != nil {log.Fatalf("error -- EmbedScipt: %v",err)}

	if dbg {fmt.Printf("out [%d]:\n%s\n", len(nidxbyt), string(nidxbyt))}

	err = os.WriteFile(outFilnam, nidxbyt, 0644)
	if err != nil {log.Fatalf("error -- could not write file: %v\n", err)}

	log.Println("success")
}

