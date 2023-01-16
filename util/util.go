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

func PrettyString(obj interface{}) string {
	if obj == nil {
		return "<nil>"
	}
	s, _ := json.MarshalIndent(obj, "", "\t")
	return string(s)
}
