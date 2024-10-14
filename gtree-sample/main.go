package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ddddddO/gtree"
)

func main() {
	var (
		root *gtree.Node
		node *gtree.Node
	)

	rootPath, err := os.Getwd()
	if err != nil {
		println(err.Error())
		return
	}

	filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		splited := strings.Split(path, "/")

		for i, s := range splited {
			if root == nil {
				root = gtree.NewRoot(s)
				node = root
				continue
			}
			if i == 0 {
				continue
			}

			tmp := node.Add(s)
			node = tmp
		}
		node = root

		return err
	})
	if err := gtree.OutputProgrammably(os.Stdout, root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	println("end")
}
