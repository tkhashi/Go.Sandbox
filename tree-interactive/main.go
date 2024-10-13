package main

import (
	"fmt"
	"os"

	gotree "github.com/disiqueira/gotree"
)

type Folder struct {
	path string
}

type File struct {
	path string
}

func GetChildren(tree gotree.Tree) (gotree.Tree, error) {
	composite, err := os.ReadDir(tree.Text())
	if err != nil {
		return nil, err
	}

	for _, de := range composite {
		leaf := tree.Add(de.Name())

		_, err := GetChildren(leaf)
		if err != nil {
			continue
		}
	}

	return tree, nil
}

func main() {
	rootPath, err := os.Getwd()
	if err != nil {
		return
	}

	tree, err := GetChildren(gotree.New(rootPath))
	if err != nil {
		print("Invalid error ", err)
	}

	fmt.Println(tree.Print())
}
