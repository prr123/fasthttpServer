package pparse


import (


)

type Path struct {
	Ext []byte
	Fnam []byte
	Fold []byte
}

func Pparse(path []byte) (p Path){

	ext := -1
	pEnd := 0
	state := 0
	for i:=len(path)-1; i>0; i-- {
		let := path[i]
		switch state {
		case 0:
			switch let {
			case '.':
				state = 1
				ext = i
			case '/':
				state =2
				pEnd = i
			default:
			}
		case 1:
		// found extension
			switch let {
			case '.':
				state = 1
				ext = i
			case '/':
				state =2
				pEnd = i
			default:
			}

		case 2:
		// found folder path


		default:
		}

		if pEnd> 0 {break}
	}

	if ext > 0 {
		p.Ext = path[ext+1:]
		p.Fnam= path[pEnd+1:]
		if pEnd > 0 {
			p.Fold = path[0:pEnd+1]
		}
	} else {
		p.Fold= path[pEnd:]
		if pEnd > 0 {
			p.Fold = path[0:pEnd+1]
			p.Fnam= path[pEnd+1:]
		}

	}
	return p
}
