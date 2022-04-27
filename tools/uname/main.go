package main

import (
	"bufio"
	"fmt"
	"github.com/akamensky/argparse"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

type CmdLineArgs struct {
	Machine *bool
	parser  *argparse.Parser
}

func (cmd *CmdLineArgs) InitFlags() {
	cmd.parser = argparse.NewParser("Dummy Uname utility", "Dummy Uname utility")
	cmd.Machine = cmd.parser.Flag("m", "machine",
		&argparse.Options{
			Required: false,
			Default:  false,
			Help:     "r",
		},
	)
}

func (cmd *CmdLineArgs) ParseArgs(args []string) error {
	err := cmd.parser.Parse(args)
	if err != nil {
		fmt.Print(cmd.parser.Usage(err))
		return err
	}
	return nil
}

type DataStruct struct {
	Machine string
}

func (data *DataStruct) ReadFromFile(filePath string) {
	var err error

	jsonFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	parseStruct := map[string]func(string){
		" +([^ ]+) [^ ]+$": func(s string) { data.Machine = s },
	}

	scanner := bufio.NewScanner(jsonFile)
	for k, callback := range parseStruct {
		if !scanner.Scan() {
			log.Fatal("cannot scan next line in the input file")
		}
		data := parseLine(scanner.Text(), k)
		callback(data)
	}
}

func parseLine(line string, regexpStr string) string {
	regex, err := regexp.CompilePOSIX(regexpStr)
	if err != nil {
		log.Fatalf("Cannot parse '%s' from '%s'", regexpStr, line)
	}
	subMatch := regex.FindStringSubmatch(line)
	if subMatch == nil {
		log.Fatalf("Cannot parse '%s' from '%s'", regexpStr, line)
	}
	return subMatch[1]
}

func main() {
	var err error
	var args CmdLineArgs

	args.InitFlags()
	err = args.ParseArgs(os.Args)
	if err != nil {
		return
	}

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	var unameData DataStruct
	unameData.ReadFromFile(path.Join(exPath, "uname.txt"))

	if *args.Machine {
		fmt.Printf("%s\n", unameData.Machine)
	}

}
