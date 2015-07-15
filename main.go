package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
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

func formatTmp(cc string) *os.File {
	fmt.Printf("=> Writing formatted cloud-config to a temporary file...\n")
	file, err := os.Open(cc)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	tmp, err := ioutil.TempFile("/tmp", "cloudcfg")
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tmp.WriteString("\"" + strings.Replace(scanner.Text(), "\"", "\\\"", -1) + "\\n\",\n")
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	return tmp
}

func inject(toInject, toReceive string, start, end int64) {
	fmt.Printf("=> Removing lines [%d:%d] in %s...\n", start, end, toReceive)

	// read in the lines to be received
	receiving, err := os.Open(toReceive)
	if err != nil {
		log.Fatal(err)
	}

	receivingScanner := bufio.NewScanner(receiving)
	preLines := ""
	postLines := ""
	var numLines int64 = 0
	for receivingScanner.Scan() {
		numLines++
		if numLines < start {
			preLines = preLines + receivingScanner.Text() + "\n"
		} else if numLines > end {
			postLines = postLines + receivingScanner.Text() + "\n"
		}
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
	injectLines := ""
	for injectingScanner.Scan() {
		injectLines = injectLines + injectingScanner.Text() + "\n"
	}
	if err := injectingScanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	// make sure that the last line gets special treatment
	k := strings.LastIndex(injectLines, "\\n\",")
	if k != -1 {
		injectLines = injectLines[:k] + "\""
	}

	// combine all parts of the file into one
	fmt.Printf("=> Injecting %s into %s...\n", toInject, toReceive)
	injectedLines := preLines + injectLines + postLines

	// write to the receiving file
	receiving.Close()
	receivingWrite, err := os.Create(toReceive)
	if err != nil {
		log.Fatal(err)
	}
	defer receivingWrite.Close()
	receivingWrite.WriteString(injectedLines)
	receivingWrite.Sync()
}

func injectAndFormat(toInject, toReceive string, start, end int64) {
	tmpFile := formatTmp(toInject).Name()
	inject(tmpFile, toReceive, start, end)
	defer os.Remove(tmpFile)
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
	app.Version = "1.0"
	app.Usage = "inject cloud-configs into cloud-formation templates"
	app.Commands = []cli.Command{
		{
			Name:  "format",
			Usage: "format a cloud-config file into an acceptable JSON structure",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "tmp",
					Usage: "write the cloud-config to a tmp file",
				},
				cli.StringFlag{
					Name:  "out",
					Usage: "write the cloud-config to this file",
				},
			},
			Action: func(c *cli.Context) {
				if c.String("out") != "" {
					formatOut(c.Args().First(), c.String("out"))
				} else if c.Bool("tmp") {
					formatTmp(c.Args().First())
				} else {
					format(c.Args().First())
				}

			},
		},
		{
			Name:  "inject",
			Usage: "inject a file into another, first removing from line1 to line2",
			Flags: []cli.Flag{
				cli.BoolTFlag{
					Name:  "format",
					Usage: "format the cloud-config before injecting it",
				},
			},
			Action: func(c *cli.Context) {
				start, err := strconv.ParseInt(c.Args().Get(2), 0, 0)
				if err != nil {
					log.Fatal("start must be an integer")
				}
				end, err := strconv.ParseInt(c.Args().Get(3), 0, 0)
				if err != nil {
					log.Fatal("start must be an integer")
				}
				if c.Bool("format") == true {
					injectAndFormat(c.Args().First(), c.Args().Get(1), start, end)
				} else {
					inject(c.Args().First(), c.Args().Get(1), start, end)
				}
			},
		},
	}
	app.Run(os.Args)
}
