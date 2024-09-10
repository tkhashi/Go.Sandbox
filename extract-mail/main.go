package main

import (
	"fmt"
	"io/ioutil"
)


func main() {
  useIoutilReadFile("row-mail.txt")
}

func useIoutilReadFile(fileName string) {
    bytes, err := ioutil.ReadFile(fileName)
    if err != nil {
        panic(err)
    }

    fmt.Println(string(bytes))
}
