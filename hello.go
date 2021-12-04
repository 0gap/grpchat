package main

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"learningGo/morestrings"
	"os"
)

func main() {
	fmt.Println(morestrings.ReverseRunes("Hello from " + os.Getenv("USER")))
	fmt.Println(cmp.Diff("Hello world", "Hello Go"))
}
