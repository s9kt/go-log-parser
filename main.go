package main

import (
	"bufio"
	"flag"
	"fmt"
	customParser "go-log-parser/parser"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	config := customParser.Settings{Format: "nginx"}

	flag.StringVar(&config.File, "file", config.File, "path to the log file (default stdin)")
	flag.StringVar(&config.Format, "format", config.Format, "The format of the log file (nginx | json | nginx-error)")
	flag.StringVar(&config.Filter, "filter", config.Filter, "Filter the log by field. field=value . When value has spaces, enclose in quotes.")
	flag.StringVar(&config.Aggregate, "aggregate", config.Aggregate, "Specify a category to count")
	flag.IntVar(&config.Top, "top", config.Top, "Show only n entries. Only works with aggregate. n = -1 will print all aggregate values.")
	flag.BoolVar(&config.Verbose, "verbose", config.Verbose, "Show full output")
	flag.Parse()

	if config.Top != 0 && config.Aggregate == "" {
		fmt.Println("Top can only be set when aggregate is also set.")
		return
	}

	var scanner *bufio.Scanner

	if config.File != "" {
		logPath, err := filepath.Abs(config.File)
		if err != nil {
			fmt.Println(err)
			return
		}

		file, err := os.Open(logPath)
		if err != nil {
			fmt.Println(err)
			return
		}

		defer file.Close()

		scanner = bufio.NewScanner(file)

	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			fmt.Println("You must either define a file with -file or pipe from stdin")
			return
		}

		scanner = bufio.NewScanner(os.Stdin)
	}

	f, err := customParser.CreateFilter(&config)
	if err != nil {
		fmt.Println(err)
		return
	}

	aggregateMap := make(map[string]int)

	switch config.Format {
	case "nginx":
		err = customParser.NginxFile(scanner, f, aggregateMap, &config)
	case "nginx-error":
		err = customParser.NginxErrorFile(scanner, f, aggregateMap, &config)
	case "json":
		fmt.Println("JSON is not yet supported.")
		return
	}

	if err != nil {
		fmt.Println(err)
		return
	}

	if config.Aggregate != "" && len(aggregateMap) > 0 {

		keys := make([]string, 0, len(aggregateMap))
		for k := range aggregateMap {
			keys = append(keys, k)
		}

		sort.Slice(keys, func(i, j int) bool {
			return aggregateMap[keys[i]] > aggregateMap[keys[j]]
		})
		// Set default
		if config.Top == 0 {
			config.Top = 10
		}

		limit := config.Top
		if limit > len(keys) || limit == -1 {
			limit = len(keys)
		}

		fmt.Println(config.Aggregate, "amount")
		for i := 0; i < limit; i++ {
			fmt.Println(keys[i], aggregateMap[keys[i]])
		}
	} else if config.Aggregate != "" && len(aggregateMap) == 0 {
		fmt.Println("No matching values found")
	}

}
