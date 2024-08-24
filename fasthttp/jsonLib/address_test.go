package address

import (
    "testing"
//    "fmt"
)

func TestGetNameAddress(t *testing.T) {

    jsonpath := "/home/peter/www/azuldist/json/adresses.json"

	namadr, err := GetNameAddress(jsonpath)
	if err != nil {t.Errorf("error -- GenNameAddress: %v\n", err)}
	PrintNamAdr(namadr)
}
