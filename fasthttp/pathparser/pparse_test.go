package pparse


import (
	"testing"
	"bytes"
	"fmt"
)


func TestPpath(t *testing.T) {

	tstpath := []byte("/img/test.png")
	fmt.Printf("testing path: %s\n", string(tstpath))

	p := Pparse(tstpath)
	fmt.Printf("Fold: \"%s\" Fnam: \"%s\" Ext: \"%s\"\n", p.Fold, p.Fnam, p.Ext)

	if !bytes.Equal(p.Fold, []byte("/img/")) {t.Errorf("p.Fold not \"/img/\"\n")}
	if !bytes.Equal(p.Fnam, []byte("test")) {t.Errorf("p.Fnam not \"test\"\n")}
	if !bytes.Equal(p.Ext, []byte("png")) {t.Errorf("p.Ext not \"png\"\n")}


	tstpath = []byte("/foo")
	fmt.Printf("testing path: %s\n", string(tstpath))

	p = Pparse(tstpath)
	fmt.Printf("Fold: \"%s\" Fnam: \"%s\" Ext: \"%s\"\n", p.Fold, p.Fnam, p.Ext)

	if !bytes.Equal(p.Fold, []byte("/foo")) {t.Errorf("p.Fold not \"/foo\"\n")}
	if !bytes.Equal(p.Fnam, []byte("")) {t.Errorf("p.Fnam not \"\"\n")}
	if !bytes.Equal(p.Ext, []byte("")) {t.Errorf("p.Ext not \"\"\n")}

	tstpath = []byte("/foo/bar")
	fmt.Printf("testing path: %s\n", string(tstpath))

	p = Pparse(tstpath)
	fmt.Printf("Fold: \"%s\" Fnam: \"%s\" Ext: \"%s\"\n", p.Fold, p.Fnam, p.Ext)

	if !bytes.Equal(p.Fold, []byte("/foo/")) {t.Errorf("p.Fold not \"/foo/\"\n")}
	if !bytes.Equal(p.Fnam, []byte("bar")) {t.Errorf("p.Fnam not \"bar\"\n")}
	if !bytes.Equal(p.Ext, []byte("")) {t.Errorf("p.Ext not \"\"\n")}




}
