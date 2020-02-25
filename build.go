package build

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const magenta = "\u001b[35m"
const cyan = "\u001b[36m"
const yellow = "\u001b[33m"
const green = "\u001b[32m"
const red = "\u001b[31m"
const reset = "\u001b[0m"

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

	targets []string

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
		fmt.Printf(magenta+"[cmd] %s %s\n"+reset, cmd, strings.Join(args, " "))
	}
	runErr := c.Run()
	if runErr != nil {
		b.AddError(runErr)
		b.printAllErrorsAndExit()
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
		b.AddError(fmt.Errorf("can't register command `%v`. Already has command with such name", subCommand))
		return
	}
	if body == nil {
		b.AddError(fmt.Errorf("can't register command `%v`. Command body can't be nil", subCommand))
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
			b.AddError(fmt.Errorf("can't find such command as: `%v`", cmd))
			b.printAllErrorsAndExit()
		}
	}

	for _, cmd := range args {
		b.targets = []string{cmd}
		b.printCurrentCommand()
		b.commands[cmd]()
		if len(b.buildErrors) > 0 {
			b.printAllErrorsAndExit()
		}
	}

	fmt.Println()
	fmt.Println(green + "Successful build" + reset)
}

func (b *Build) AddError(err error) {
	if err == nil {
		return
	}
	b.buildErrors = append(b.buildErrors, err)
}

func (b *Build) AddTarget(newTarget string) func() {
	b.targets = append(b.targets, newTarget)
	b.printCurrentCommand()
	return func() {
		b.targets = b.targets[:len(b.targets)-1]
	}
}

func (b *Build) Info(message string) {
	if !b.verbose {
		return
	}
	fmt.Println(green + "[info] " + message + reset)
}

func (b *Build) Warn(message string) {
	fmt.Println(yellow + "[warn] " + message + reset)
}

func (b *Build) printCurrentCommand() {
	fmt.Println(cyan + b.targetsToString() + reset)
}

func (b *Build) printAvailableTargets() {
	fmt.Printf("Available targets:\n")
	for _, cmd := range b.commandsRegistrationOrd {
		fmt.Printf("    - "+cyan+"%+v\n"+reset, cmd)
	}
}

func (b *Build) targetsToString() string {
	if len(b.targets) == 0 {
		return ""
	}
	buf := bytes.NewBufferString("[" + b.targets[0])
	for _, target := range b.targets[1:] {
		_, _ = buf.WriteString(" | ")
		_, _ = buf.WriteString(target)
	}
	buf.WriteString("]")
	return buf.String()
}

func (b *Build) printAllErrorsAndExit() {
	fmt.Println()
	for _, err := range b.buildErrors {
		fmt.Printf(red+"%v\n"+reset, err)
	}
	fmt.Println(red + b.targetsToString() + " Build failed" + reset)
	os.Exit(-1)
}
