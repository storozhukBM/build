package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"

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
		additionalBuildFunc()
		b.Warn(fmt.Sprintf("hello %v!!!", "bananas"))
	}},

	{`fail`, func() {
		b.Info(fmt.Sprintf("going to fail"))
		b.AddTarget("targetThatWillFail")
		b.AddError(fmt.Errorf("This thing supose to fail"))
	}},

	{`failFromRun`, func() {
		b.Info(fmt.Sprintf("going to fail"))
		b.Run(strconv.Itoa(rand.Int()))
	}},
}

func additionalBuildFunc() {
	defer b.AddTarget("additionalStep")()
	b.Info(fmt.Sprintf("hey %v!!!", "brother"))
}

func main() {
	b.Register(commands)
	buildErr := b.BuildFromOsArgs()
	if buildErr != nil {
		fmt.Println(buildErr)
		os.Exit(-1)
	}
}
