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
	"strconv"
)

// CmdLineArgs
// Represents Cmd line arguments passed to  cmd line of the target program.
// Program operates in two modes
// - build Docker images (Docker mode),
// - build package (package mode)
// Exactly one of these modes can be active in a time.
type CmdLineArgs struct {
	FlagR  *bool
	FlagS  *bool
	FlagI  *bool
	parser *argparse.Parser
}

func (cmd *CmdLineArgs) InitFlags() {
	cmd.parser = argparse.NewParser("Dummy LSB Release ", "My Test Parser")
	cmd.FlagR = cmd.parser.Flag("r", "r",
		&argparse.Options{
			Required: false,
			Default:  false,
			Help:     "r",
		},
	)
	cmd.FlagS = cmd.parser.Flag("s", "s",
		&argparse.Options{
			Required: false,
			Default:  false,
			Help:     "s",
		},
	)
	cmd.FlagI = cmd.parser.Flag("i", "i",
		&argparse.Options{
			Required: false,
			Default:  false,
			Help:     "i",
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
	ReleaseNumber int
	DistributorID string
}

func (data *DataStruct) ReadFromFile(filePath string) {
	var err error

	jsonFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	parseStruct := map[string]func(string){
		"^Distributor ID:\t([^\t]+)$": func(s string) { data.DistributorID = s },
		"^Release:\t([^\t]+)$":        func(s string) { data.ReleaseNumber, _ = strconv.Atoi(s) },
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
	filePath := path.Join(exPath, "lsb_release.txt")

	var lsbReleaseData DataStruct
	lsbReleaseData.ReadFromFile(filePath)

	if *args.FlagS == true {
		if *args.FlagI {
			fmt.Println(lsbReleaseData.DistributorID)
		}
		if *args.FlagR {
			fmt.Println(lsbReleaseData.ReleaseNumber)
		}
	} else {
		if *args.FlagI {
			fmt.Printf("Distributor ID:\t%s\n", lsbReleaseData.DistributorID)
		}
		if *args.FlagR {
			fmt.Printf("Release:\t%d\n", lsbReleaseData.ReleaseNumber)
		}
	}

}
