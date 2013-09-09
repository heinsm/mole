package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/calmh/mole/configuration"
	"github.com/calmh/mole/tmpfileset"
	"github.com/jessevdk/go-flags"
)

type cmdDig struct {
	Local bool `short:"l" long:"local" description:"Local file, not remote tunnel definition"`
}

var digParser *flags.Parser

func init() {
	cmd := cmdDig{}
	digParser = globalParser.AddCommand("dig", "Dig a tunnel", "Dig connects to a remote destination and sets up configured local TCP tunnels", &cmd)
}

func (c *cmdDig) Execute(args []string) error {
	setup()

	if len(args) != 1 {
		writeParser.WriteHelp(os.Stdout)
		fmt.Println()
		return fmt.Errorf("dig: missing required options <tunnelname>\n")
	}

	cfg, err := configuration.Load(args[0])
	if err != nil {
		log.Fatal(err)
	}

	var fs tmpfileset.FileSet
	sshConfig(cfg, &fs)
	expectConfig(cfg, &fs)
	fs.Save(homeDir)

	cmd := exec.Command("expect", "-f", fs.PathFor("expect-config"))
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()

	return nil
}
