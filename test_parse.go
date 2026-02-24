package main

import (
	"encoding/json"
	"fmt"
	"zeno/pkg/engine"
)

func main() {
	node, _ := engine.ParseString(`__native_write "hello from php"`, "test")
	b, _ := json.MarshalIndent(node, "", "  ")
	fmt.Println(string(b))
}
