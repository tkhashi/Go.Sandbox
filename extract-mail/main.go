package main

import (
	"fmt"
	"io"
	"os"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"github.com/joho/godotenv"
)


func main() {

    err := godotenv.Load()
    if err != nil {
    	return 
    }
    
    apikey := os.Getenv("apikey")
    endpoint := os.Getenv("endpoint")
    bytes := useReadFile("raw-mail.txt")

    analyze(endpoint, apikey, string(bytes))
}

func useReadFile(fileName string) []byte {
    bytes, err := os.ReadFile(fileName)
    if err != nil {
        panic(err)
    }

    return bytes
}

func analyze(endpoint string, apikey string ,mailMsg string) {
    r := strings.NewReader(mailMsg)
    m, _ := mail.ReadMessage(r)

    header := m.Header
    
    body, err := io.ReadAll(m.Body)
    fmt.Println("宛先", header.Get("Delivered-To"))
    fmt.Println("経由サーバー", header.Get("Received"))
    //fmt.Println("メールボディ", body)
    if err != nil {
	return
    }
    // fmt.Printf("%s", body)
    fmt.Println("Body: ", string(body)[:90])

    summarize(endpoint, apikey, string(body)[:90])
}

func summarize(endpoint string, apikey string ,msg string) {


    client := http.Client{
        Timeout: 10 * time.Second,
    }

    req, err := http.NewRequest(http.MethodPost, endpoint, nil)
    q := req.URL.Query()
    q.Add("apikey", apikey)
    q.Add("sentences", msg)
    req.URL.RawQuery = q.Encode()

    res, err := client.Do(req)

    if err != nil {
	fmt.Println(res.Body)
    } else {
	fmt.Println("Error request")
    }

    defer res.Body.Close()
}

