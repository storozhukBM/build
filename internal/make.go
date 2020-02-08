package main

import . "github.com/storozhukBM/build"

var b = NewBuild(BuildOptions{})
var commands = []Command{

	{`build`, b.RunCmd(
		Go, `build`, `./...`,
	)},

	{`verify`, func() {
		b.Run(Go, `vet`, `-composites=false`, `./...`)
		b.Run(Go, `build`, `./...`)
	}},
}

func main() {
	b.Register(commands)
	b.BuildFromOsArgs()
}
