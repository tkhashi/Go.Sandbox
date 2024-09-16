package main

type Noder interface {
	GetName() string
}

type File struct {
	name string
}

type Directory struct {
	name     string
	children []Noder
}

func (f File) GetName() string {
	return f.name
}

func (f Directory) GetName() string {
	return f.name
}

func (d Directory) GetChildren() []Noder {
	return d.children
}
