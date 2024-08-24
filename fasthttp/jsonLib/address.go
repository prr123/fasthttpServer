package address

import (
	"fmt"
	"os"
	"github.com/goccy/go-json"
)

type Name struct {
	NamId int
	First string
	Mid string
	Last string
	Short string
	Phone string
	Email string
}

type Adress struct {
	AdrId int
	Street string
	StNum string
	AptNum string
	City string
	Zip string
	Country string
}

type NamAdr struct {
	First string
	Mid string
	Last string
	Short string
	Street string
	StNum string
	AptNum string
	City string
	Zip string
	Country string
}

func GetNameAddress (filnam string) (namadr []NamAdr, err error) {

	bSlice, err := os.ReadFile(filnam)
	if err != nil {return namadr, fmt.Errorf("GetNameAddress read file: %v", err)}

	err = json.Unmarshal(bSlice, &namadr)
	if err != nil {return namadr, fmt.Errorf("GetNameAddress unmarshal: %v", err)}

	return namadr, nil
}

func PrintNamAdr(namadr []NamAdr) {

	fmt.Printf("*********** Name Address: %d Entries *************\n", len(namadr))
	for i:=0; i< len(namadr); i++ {
		na := namadr[i]
		fmt.Printf("  First:   %s\n", na.First)
		fmt.Printf("  Last:    %s\n", na.Last)
	}
	fmt.Printf("**************** End Name Address *****************\n")
}


