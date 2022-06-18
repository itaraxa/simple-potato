package main

import (
	"fmt"

	"github.com/itaraxa/simple-potato/internal/fileOperation"
)

func main() {
	fmt.Println("TESTS!!!!")

	t1 := new(fileOperation.MetaFile)
	t1.Init("test/tmp/1/CD.txt")
	t1.PrettyOut()

	t2 := new(fileOperation.MetaFile)
	t2.Init("test/tmp/2/щ.txt")
	t2.PrettyOut()
}
