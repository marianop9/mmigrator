package main

import (
	"fmt"
	"path"
)

func main() {
	s := make([]int, 2)

	s = append(s, 3)

	fmt.Println(path.Ext("a.sql"))
	fmt.Printf("%+v", s)

}
