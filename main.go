package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
)

func format(cc string) {
	fmt.Printf("=> Writing formatted cloud-config to STDOUT...\n")
	file, err := os.Open(cc)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println("\"" + strings.Replace(scanner.Text(), "\"", "\\\"", -1) + "\\n\",")
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func formatOut(cc, outPath string) {
	fmt.Printf("=> Writing formatted cloud-config to %s...\n", cc)
	file, err := os.Open(cc)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	out, err := os.Create(outPath)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		out.WriteString("\"" + strings.Replace(scanner.Text(), "\"", "\\\"", -1) + "\\n\",\n")
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func inject(toInject, toReceive string, start, end int64) {
	fmt.Printf("=> Removing lines [%d:%d] in %s...\n", start, end, toReceive)
	// read in the lines to be received
	receiving, err := os.Open(toReceive)
	if err != nil {
		log.Fatal(err)
	}

	receivingScanner := bufio.NewScanner(receiving)
	lines := []string{}
	for receivingScanner.Scan() {
		lines = append(lines, receivingScanner.Text())
	}
	if err := receivingScanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	// read in the lines to be injected
	injecting, err := os.Open(toInject)
	if err != nil {
		log.Fatal(err)
	}
	defer injecting.Close()

	injectingScanner := bufio.NewScanner(injecting)
	linesToInject := []string{}
	for injectingScanner.Scan() {
		linesToInject = append(linesToInject, injectingScanner.Text())
	}
	if err := injectingScanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	// combine all parts of the file into one
	fmt.Printf("=> Injecting %s into %s...", toInject, toReceive)
	pre := lines[0 : start-1]
	post := lines[end:]
	injectedLines := append(pre, linesToInject...)
	injectedLines = append(injectedLines, post...)

	// write to the receiving file
	receiving.Close()
	receivingWrite, err := os.Create(toReceive)
	if err != nil {
		log.Fatal(err)
	}
	defer receivingWrite.Close()

	for _, s := range injectedLines {
		receivingWrite.WriteString(s + "\n")
	}
	receivingWrite.Sync()
}

func main() {
	app := cli.NewApp()
	app.Name = "cumulus"
	app.Authors = []cli.Author{
		{
			Name:  "Ben Bailey",
			Email: "bennettbailey@gmail.com",
		},
	}
	app.Usage = "inject cloud-configs into cloud-formation templates"
	app.Commands = []cli.Command{
		{
			Name:  "format",
			Usage: "format a cloud-config file into an acceptable JSON structure",
			Action: func(c *cli.Context) {
				if len(c.Args()) == 1 {
					format(c.Args().First())
				} else if len(c.Args()) == 2 {
					formatOut(c.Args().First(), c.Args().Get(1))
				} else {
					log.Fatal("Format takes one or two arguments.")
				}

			},
		},
		{
			Name:  "inject",
			Usage: "inject a file into another, first removing from line1 to line2",
			Action: func(c *cli.Context) {
				start, _ := strconv.ParseInt(c.Args().Get(2), 0, 0)
				end, _ := strconv.ParseInt(c.Args().Get(3), 0, 0)
				inject(c.Args().First(), c.Args().Get(1), start, end)
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "format",
					Usage: "format the cloud-config before injecting it",
				},
			},
		},
	}
	app.Run(os.Args)
}
