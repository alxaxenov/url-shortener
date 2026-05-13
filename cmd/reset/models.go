package main

type PackageData struct {
	Name    string
	Path    string
	Structs []StructData
}

type StructData struct {
	Name string
	Body string
}
