package _main

import (
	"fmt"
	"os"

	gotree "github.com/disiqueira/gotree"
)

func GetChildren(tree gotree.Tree) (gotree.Tree, error) {
	composite, err := os.ReadDir(tree.Text())
	if err != nil {
		println("err GetChildren ", err.Error(), tree.Text())
		return nil, err
	}

	for _, de := range composite {
		leaf := tree.Add(de.Name())
		childTree, err := GetChildren(leaf)
		if err != nil {
			continue
		}
		tree.AddTree(childTree)
	}

	return tree, nil
}

func main() {
	rootPath, err := os.Getwd()
	if err != nil {
		println(err.Error())
		return
	}

	tree, err := GetChildren(gotree.New(rootPath))
	if err != nil {
		println(err.Error())
		return
	}

	fmt.Println(tree.Print())
}
