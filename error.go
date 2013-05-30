package main

import (
	"fmt"
	"encoding/json"
)

func whatever (msg string) {
	mm := make ([]byte, 1024)
	err := json.Unmarshal ([]byte(msg), mm)
	if err != nil {
		switch (Error{err, "context"}.Handle()) {
		case 1:
			//Do something
		}
		
	}

}
func (e Error) Handle() int {
	fmt.Println(e.err.Error())
	fmt.Println(e.context)
	return 1
}
