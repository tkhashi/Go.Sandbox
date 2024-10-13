package main

import (
	"io/fs"
	"os"
	"path/filepath"

	gotree "github.com/disiqueira/gotree"
)

type Folder struct {
	path string
}

type File struct {
	path string
}

func GetChildren(root string, tree gotree.Tree) (gotree.Tree, error) {
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		tree.Add(path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return tree, nil
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		return
	}

	tree := gotree.New(root)

	children, err := GetChildren(root, tree)

	println(children.Print())
}
