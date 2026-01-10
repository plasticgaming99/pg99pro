package main

import (
	"os"

	"github.com/plasticgaming99/pg99pro/cmd"
)

func main() {
	cmd.Execute(os.Args[1:])
}
