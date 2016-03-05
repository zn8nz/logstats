// logstats processes logs and and prints the number of lines that contain a given regexp.
// Pass as arguments the options -o to define the order of number fields inside the timestamp and
// -t to set a time interval for grouping, ranging from 10 minutes to 1 day.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"
)

const (
	_10min   = 10
	_15min   = 15
	_30min   = 30
	_1hour   = 1
	_2hour   = 2
	_3hours  = 3
	_6hours  = 6
	_12hours = 12
	_1day    = 0
)

var (
	factor = [...]int{1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9}

	settings = struct {
		order   string
		groupBy int
		rxCount *regexp.Regexp
		rxKey   *regexp.Regexp
	}{
		order:   "ymdhisf",
		groupBy: _1hour,
	}

	counter = make(map[string]int)

	rx *regexp.Regexp
)

func main() {
	// parse options
	var key string

	flag.StringVar(&settings.order, "o", "ymdhisf", "order of the number fields: y=year, m=month, d=day, h=hour, i=min, s=sec, f=fraction")
	flag.IntVar(&settings.groupBy, "t", 0, "group by: 10=ten minutes, 15=¼ hour, 30=½ hour, 1=hour, 2=two hr, 3=three hr, 6=six hr, 12=½ day, 0=day")
	flag.StringVar(&key, "k", "", "regexp that defines the key to group by; cannot use with -o and -t")
	flag.Parse()
	if flag.NArg() < 2 {
		fmt.Println("Usage: <options> <regexp> <glob>")
		return
	}

	if key != "" {
		settings.rxKey = regexp.MustCompile(key)
	} else {
		// find timestamp numbers
		rx = regexp.MustCompile(`\d+`)
	}
	pattern := flag.Arg(0)
	settings.rxCount = regexp.MustCompile(pattern)

	// go through the files
	err := processFiles(flag.Arg(1))
	if err != nil {
		fmt.Println(err)
	}
	// sort
	arr := make([]string, len(counter))
	i := 0
	for k, v := range counter {
		arr[i] = fmt.Sprintf("%s,%9d", k, v)
		i++
	}
	sort.Strings(arr)
	for _, s := range arr {
		fmt.Println(s)
	}
}

func processFiles(glob string) error {
	files, err := filepath.Glob(glob)
	if err != nil {
		return err
	}
	for _, file := range files {
		file, err := os.Open(file)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if settings.rxKey != nil {
				match := settings.rxKey.FindString(line)
				if match == "" {
					continue
				}
				if settings.rxCount.MatchString(line) {
					counter[match]++
				}
			} else if settings.rxCount.MatchString(line) {
				ts, _ := parseTimestamp(settings.order, line)
				min := ts.Minute()
				hour := ts.Hour()
				div := settings.groupBy
				switch div {
				case _1day:
					min = 0
					hour = 0
				case _1hour:
					min = 0
				case _2hour, _3hours, _6hours, _12hours:
					min = 0
					hour = div * int(ts.Hour()/div)
				case _10min, _15min, _30min:
					min = div * int(ts.Minute()/div)
				default:
					return errors.New("Invalid value interval: " + strconv.Itoa(div))
				}
				tt := time.Date(ts.Year(), ts.Month(), ts.Day(), hour, min, 0, 0, ts.Location())
				key := tt.Format("2006-01-02 15:04:05")
				counter[key]++
				//fmt.Printf("%s|%s\n", tt.Format("2006-01-02 15:04:05"), line)
			}
		}
	}
	return nil
}

// timestamp analyses a timestamp string and returns it as a *time.Time.
// layout string is a combination of y, m, d, h, i, s, f, -
// where -=skip the next number and i=minutes, f=fraction of seconds
func parseTimestamp(layout, ts string) (*time.Time, error) {
	arr := rx.FindAllStringSubmatch(ts, len(layout))
	var year, month, day, hour, min, sec, nsec int
	for i, s := range arr {
		c := layout[i]
		if c == '-' {
			continue
		}
		number, _ := strconv.Atoi(s[0])

		switch c {
		case 'y':
			year = number
		case 'm':
			month = number
		case 'd':
			day = number
		case 'h':
			hour = number
		case 'i':
			min = number
		case 's':
			sec = number
		case 'f':
			nsec = number * factor[len(s[0])]
		default:
			return nil, errors.New("Invalid timestamp characters in layout; use [-ymdhisf]")
		}
	}
	if year < 100 {
		year += 2000
	}
	var t = time.Date(year, time.Month(month), day, hour, min, sec, nsec, time.UTC)
	return &t, nil
}
