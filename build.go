package build

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const Go = "go"

type Command struct {
	Name string
	Body func()
}

type BuildOptions struct {
	Env    map[string]string
	Stdout io.Writer
	Stderr io.Writer
}

type Build struct {
	verbose     bool
	env         map[string]string
	stdout      io.Writer
	stderr      io.Writer
	buildErrors []error

	currentTarget string

	commands                map[string]func()
	commandsRegistrationOrd []string

	onceRuns map[string]struct{}
}

func NewBuild(o BuildOptions) *Build {
	result := &Build{
		env:    nil,
		stdout: os.Stdout,
		stderr: os.Stderr,

		commands: make(map[string]func()),
		onceRuns: make(map[string]struct{}),
	}
	if o.Env != nil {
		result.env = make(map[string]string, len(o.Env))
		for k, v := range o.Env {
			result.env[k] = v
		}
	}
	if o.Stdout != nil {
		result.stdout = o.Stdout
	}
	if o.Stderr != nil {
		result.stderr = o.Stderr
	}
	return result
}

func (b *Build) Run(cmd string, args ...string) {
	c := exec.Command(cmd, args...)
	c.Env = os.Environ()
	for k, v := range b.env {
		c.Env = append(c.Env, k+"="+v)
	}
	c.Stderr = b.stdout
	c.Stdout = b.stderr
	c.Stdin = os.Stdin
	if b.verbose {
		if b.currentTarget != "" {
			fmt.Printf("[%s] ", b.currentTarget)
		}
		fmt.Printf("%s %s\n", cmd, strings.Join(args, " "))
	}
	runErr := c.Run()
	if runErr != nil {
		b.buildErrors = append(b.buildErrors, runErr)
	}
}

func (b *Build) ForceRun(cmd string, args ...string) {
	b.Run(cmd, args...)
	b.buildErrors = nil
}

func (b *Build) RunCmd(cmd string, args ...string) func() {
	return func() {
		b.Run(cmd, args...)
	}
}

func (b *Build) RunForceCmd(cmd string, args ...string) func() {
	return func() {
		b.ForceRun(cmd, args...)
	}
}

func (b *Build) ForceShRun(cmd string, args ...string) {
	b.ShRun(cmd, args...)
	b.buildErrors = nil
}

func (b *Build) ShRunCmd(cmd string, args ...string) func() {
	return func() {
		b.ShRun(cmd, args...)
	}
}

func (b *Build) ShRun(cmd string, args ...string) {
	fullCmd := []string{cmd}
	fullCmd = append(fullCmd, args...)
	b.Run("/bin/sh", "-c", strings.Join(fullCmd, " "))
}

func (b *Build) Cmd(subCommand string, body func()) {
	_, ok := b.commands[subCommand]
	if ok {
		b.buildErrors = append(
			b.buildErrors, fmt.Errorf("can't register command `%v`. Already has command with such name", subCommand),
		)
		return
	}
	if body == nil {
		b.buildErrors = append(
			b.buildErrors, fmt.Errorf("can't register command `%v`. Command body can't be nil", subCommand),
		)
		return
	}
	b.commands[subCommand] = body
	b.commandsRegistrationOrd = append(b.commandsRegistrationOrd, subCommand)
}

func (b *Build) Register(commands []Command) {
	for _, cmd := range commands {
		b.Cmd(cmd.Name, cmd.Body)
	}
}

func (b *Build) Once(name string, body func()) {
	if _, ok := b.onceRuns[name]; ok {
		return
	}
	b.onceRuns[name] = struct{}{}
	body()
}

func (b *Build) BuildFromOsArgs() {
	b.Build(os.Args[1:])
}

func (b *Build) Build(args []string) {
	if len(b.buildErrors) > 0 {
		b.printAllErrorsAndExit()
		return
	}

	if len(args) == 0 || args[0] == "-h" {
		b.printAvailableTargets()
		return
	}

	if args[0] == "-v" {
		b.verbose = true
		args = args[1:]
	}

	for _, cmd := range args {
		if _, ok := b.commands[cmd]; !ok {
			b.printAvailableTargets()
			b.buildErrors = append(b.buildErrors, fmt.Errorf("can't find such command as: `%v`", cmd))
			b.printAllErrorsAndExit()
		}
	}

	for _, cmd := range args {
		b.currentTarget = cmd
		b.printCurrentCommand()
		b.commands[cmd]()
		if len(b.buildErrors) > 0 {
			b.printAllErrorsAndExit()
		}
	}
}

const blue = "\u001b[36m"
const red = "\u001b[31m"
const reset = "\u001b[0m"

func (b *Build) printCurrentCommand() {
	fmt.Println(blue + "[" + b.currentTarget + "]" + reset)
}

func (b *Build) printAvailableTargets() {
	fmt.Printf("Available targets:\n")
	for _, cmd := range b.commandsRegistrationOrd {
		fmt.Printf("    - "+blue+"%+v\n"+reset, cmd)
	}
}

func (b *Build) printAllErrorsAndExit() {
	for _, err := range b.buildErrors {
		fmt.Printf(red+"%v\n"+reset, err)
	}
	fmt.Println("Can't execute build")
	os.Exit(-1)
}
