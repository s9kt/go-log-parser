package customParser

import (
	"bufio"
	"fmt"
	"regexp"
)

type Settings struct {
	File      string
	Format    string
	Filter    string
	Aggregate string
	Top       int
	Verbose   bool
}

type nginxFormattedLine struct {
	Raw           string
	RemoteAddr    string
	RemoteUser    string
	TimeLocal     string
	ReqMethod     string
	ReqPath       string
	HttpVer       string
	Status        string
	BytesSent     string
	HttpReferer   string
	HttpUserAgent string
	HttpForwarded string
}

type nginxErrorLine struct {
	Raw          string
	TimeLocal    string
	LogLevel     string
	Pid          string
	Tid          string
	ConnectionId string
	Message      string
	Client       string
	Server       string
	ReqMethod    string
	ReqPath      string
	HttpVer      string
	Host         string
}

type Filter struct {
	Field string
	Value string
}

var nginxMainFormatRegex = regexp.MustCompile(`^(\S+) - (\S+) \[([^\]]+)\] "(\S+) (\S+) (\S+)" (\d{3}) (\d+|-) "([^"]*)" "([^"]*)" "([^"]*)"$`)
var nginxCombinedFormatRegex = regexp.MustCompile(`^(\S+) - (\S+) \[([^\]]+)\] "(\S+) (\S+) (\S+)" (\d{3}) (\d+|-) "([^"]*)" "([^"]*)"$`)

var nginxErrorWithRequestRegex = regexp.MustCompile(`^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}) \[(\w+)\] (\d+)#(\d+): (?:\*(\d+) )?(.+), client: (\S+), server: (\S+), request: "(\S+) (\S+) (\S+)", host: "([^"]+)"$`)
var nginxErrorNoRequestRegex = regexp.MustCompile(`^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}) \[(\w+)\] (\d+)#(\d+): (.+)$`)

var filterRegex = regexp.MustCompile(`^(\w+)=(.+)$`)

func CreateFilter(config *Settings) (Filter, error) {
	if config.Filter == "" {
		return Filter{}, nil
	}
	matches := filterRegex.FindStringSubmatch(config.Filter)
	if matches == nil {
		return Filter{}, fmt.Errorf("invalid filter %q, must be nginx_field=value", config.Filter)
	}

	return Filter{Field: matches[1], Value: matches[2]}, nil
}

func nginxFieldValue(line nginxFormattedLine, field string) string {
	fields := map[string]string{
		"remote_addr":     line.RemoteAddr,
		"remote_user":     line.RemoteUser,
		"time_local":      line.TimeLocal,
		"req_method":      line.ReqMethod,
		"req_path":        line.ReqPath,
		"http_ver":        line.HttpVer,
		"status":          line.Status,
		"bytes_sent":      line.BytesSent,
		"http_referer":    line.HttpReferer,
		"http_user_agent": line.HttpUserAgent,
		"http_forwarded":  line.HttpForwarded,
	}
	return fields[field]
}

func nginxErrorFieldValue(line nginxErrorLine, field string) string {
	fields := map[string]string{
		"time_local":    line.TimeLocal,
		"log_level":     line.LogLevel,
		"pid":           line.Pid,
		"tid":           line.Tid,
		"connection_id": line.ConnectionId,
		"message":       line.Message,
		"client":        line.Client,
		"server":        line.Server,
		"req_method":    line.ReqMethod,
		"req_path":      line.ReqPath,
		"http_ver":      line.HttpVer,
		"host":          line.Host,
	}
	return fields[field]
}

func handleNginxMatch(parsedLine nginxFormattedLine, f Filter, aggregateMap map[string]int, config *Settings) {
	if config.Filter != "" && nginxFieldValue(parsedLine, f.Field) != f.Value {
		return
	}

	if config.Aggregate != "" {
		aggregateMap[nginxFieldValue(parsedLine, config.Aggregate)]++
	} else {
		fmt.Println(parsedLine.Raw)
	}
}

func handleNginxErrorMatch(parsedLine nginxErrorLine, f Filter, aggregateMap map[string]int, config *Settings) {
	if config.Filter != "" && nginxErrorFieldValue(parsedLine, f.Field) != f.Value {
		return
	}

	if config.Aggregate != "" {
		aggregateMap[nginxErrorFieldValue(parsedLine, config.Aggregate)]++
	} else {
		fmt.Println(parsedLine.Raw)
	}
}

func NginxFile(scanner *bufio.Scanner, f Filter, aggregateMap map[string]int, config *Settings) error {

	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		isMain := true

		matches := nginxMainFormatRegex.FindStringSubmatch(line)
		if matches == nil {
			matches = nginxCombinedFormatRegex.FindStringSubmatch(line)
			if matches == nil {
				if config.Verbose {
					fmt.Println("Couldn't parse line", line)
				}
				continue
			}
			isMain = false
		}

		parsedLine := nginxFormattedLine{
			Raw:           line,
			RemoteAddr:    matches[1],
			RemoteUser:    matches[2],
			TimeLocal:     matches[3],
			ReqMethod:     matches[4],
			ReqPath:       matches[5],
			HttpVer:       matches[6],
			Status:        matches[7],
			BytesSent:     matches[8],
			HttpReferer:   matches[9],
			HttpUserAgent: matches[10],
		}

		if isMain {
			parsedLine.HttpForwarded = matches[11]
		}

		handleNginxMatch(parsedLine, f, aggregateMap, config)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil

}

func NginxErrorFile(scanner *bufio.Scanner, f Filter, aggregateMap map[string]int, config *Settings) error {
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		hasRequest := true

		matches := nginxErrorWithRequestRegex.FindStringSubmatch(line)
		if matches == nil {
			matches = nginxErrorNoRequestRegex.FindStringSubmatch(line)
			if matches == nil {
				if config.Verbose {
					fmt.Println("Couldn't parse line", line)
				}
				continue
			}
			hasRequest = false
		}

		parsedLine := nginxErrorLine{
			Raw:       line,
			TimeLocal: matches[1],
			LogLevel:  matches[2],
			Pid:       matches[3],
			Tid:       matches[4],
			Message:   matches[5],
		}

		if hasRequest {
			parsedLine.ConnectionId = matches[5]
			parsedLine.Message = matches[6]
			parsedLine.Client = matches[7]
			parsedLine.Server = matches[8]
			parsedLine.ReqMethod = matches[9]
			parsedLine.ReqPath = matches[10]
			parsedLine.HttpVer = matches[11]
			parsedLine.Host = matches[12]
		}

		handleNginxErrorMatch(parsedLine, f, aggregateMap, config)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
