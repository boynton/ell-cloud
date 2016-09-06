package main

import (
	"github.com/boynton/ell"
	cloud "github.com/boynton/ell-cloud"
)

func main() {
	ell.Main(new(cloud.Extension))
}
