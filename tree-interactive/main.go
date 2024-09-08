package main

import (
    "fmt"
    "io/fs"
    "log"
    "os"
)

func main() {
    pwd := getCurrentPath()
    listFiles(pwd)
}

// 現在のパスを返す
func getCurrentPath() string {
    pwd, err := os.Getwd()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    fmt.Println(pwd)

    return pwd
}


// dir配下のファイル、ディレクトリを返す
func listFiles(dir string) []string {
    root := os.DirFS(dir)

    mdFiles, err := fs.Glob(root, "*")

    if err != nil {
       log.Fatal(err)
    }

    var files []string
    for _, v := range mdFiles {
       files = append(files,  v)
       fmt.Println(v)
    }
    return files
}
