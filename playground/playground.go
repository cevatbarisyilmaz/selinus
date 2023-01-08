package main

import "fmt"

func main() {
	res1 := []string{"a", "b"}
	res2 := res1
	fmt.Println(res1, res2)
	res1 = res1[:1]
	fmt.Println(res1, res2)
	res2[1] = "z"
	fmt.Println(res1, res2)
	res1 = append(res1, "c")
	fmt.Println(res1, res2)
	res2 = append(res2, "d")
	fmt.Println(res1, res2)
	res1 = append(res1, "e")
	fmt.Println(res1, res2)
	res1 = append(res1, "f")
	fmt.Println(res1, res2)
}
