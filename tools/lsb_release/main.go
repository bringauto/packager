package main

import (
	"bufio"
	"fmt"
	"github.com/akamensky/argparse"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
)

// CmdLineArgs
// Represents Cmd line arguments passed to  cmd line of the target program.
// Program operates in two modes
// - build Docker images (Docker mode),
// - build package (package mode)
// Exactly one of these modes can be active in a time.
type CmdLineArgs struct {
	FlagR        *bool
	FlagS        *bool
	FlagI        *bool
	FlagValidate *bool
	parser       *argparse.Parser
}

func (cmd *CmdLineArgs) InitFlags() {
	cmd.parser = argparse.NewParser("Dummy LSB Release ", "Dummy LSB Release")
	cmd.FlagR = cmd.parser.Flag("r", "r",
		&argparse.Options{
			Required: false,
			Help:     "r",
		},
	)
	cmd.FlagS = cmd.parser.Flag("s", "s",
		&argparse.Options{
			Required: false,
			Help:     "s",
		},
	)
	cmd.FlagI = cmd.parser.Flag("i", "i",
		&argparse.Options{
			Required: false,
			Help:     "i",
		},
	)
	cmd.FlagValidate = cmd.parser.Flag("", "validate",
		&argparse.Options{
			Required: false,
			Help:     "Validate the input file",
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
	ReleaseNumber string
	DistributorID string
}

func (data *DataStruct) ReadFromFile(filePath string, validate bool) {
	var err error

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	parseStruct := map[string]func(string){
		"^Distributor ID:\t([^\t]+)$": func(s string) { data.DistributorID = s },
		"^Release:\t([^\t]+)$":        func(s string) { data.ReleaseNumber = s },
	}

	scanner := bufio.NewScanner(file)
	keys := reflect.ValueOf(parseStruct).MapKeys()

	handled := 0
	for scanner.Scan() {
		line := scanner.Text()
		for _, key := range keys {
			keyString := key.String()
			data := parseLine(line, keyString)
			if data != "" {
				parseStruct[keyString](data)
				handled++
				break
			}
		}
	}
	if validate && len(keys) != handled {
		log.Panicf("Not all needed values were extracted!")
	}
}

func parseLine(line string, regexpStr string) string {
	regex, err := regexp.Compile(regexpStr)
	if err != nil {
		log.Panicf("Cannot compile regex '%s'", regexpStr)
	}
	subMatch := regex.FindStringSubmatch(line)
	if subMatch == nil {
		return ""
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
	lsbReleaseData.ReadFromFile(filePath, *args.FlagValidate)

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
			fmt.Printf("Release:\t%s\n", lsbReleaseData.ReleaseNumber)
		}
	}
}
