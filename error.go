package main

import (
	"fmt"
	"encoding/json"
	"os"
)

// 0 continue, 1 try again, 2 kill all
func whatever (msg string) {
	mm := make ([]byte, 1024)
	for {
		err := json.Unmarshal ([]byte(msg), mm)
		//Could either check if the error != nil before wrapping, or after, same effect with negligible overhead, cleaner looking code for the latter, but the overhead exists.
		
		if err != nil {
			switch (Error{err, "context"}.Handle()) {
			case 0:
				break
			case 1:
				continue
			case 2:
				os.Exit(1)
			}
			
		}
	}
}
func (e Error) Handle() int {
	fmt.Println(e.err.Error())
	fmt.Println(e.context)
	return 1
}
