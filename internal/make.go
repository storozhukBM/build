package main

import (
	"fmt"
	. "github.com/storozhukBM/build"
)

var b = NewBuild(BuildOptions{})
var commands = []Command{

	{`build`, b.RunCmd(
		Go, `build`, `./...`,
	)},

	{`verify`, func() {
		b.Run(Go, `vet`, `-composites=false`, `./...`)
		b.Run(Go, `build`, `./...`)
	}},

	{`test`, func() {
		b.Info(fmt.Sprintf("hello %v!!!", "sailor"))
		b.Warn(fmt.Sprintf("hello %v!!!", "bananas"))
	}},
}

func main() {
	b.Register(commands)
	b.BuildFromOsArgs()
}
