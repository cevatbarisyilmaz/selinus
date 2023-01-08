package util

import (
	"encoding/json"
	"fmt"
)

func PrettyPrint(obj interface{}) {
	if obj == nil {
		fmt.Println("<nil>")
		return
	}
	s, _ := json.MarshalIndent(obj, "", "\t")
	fmt.Println(string(s))
}
