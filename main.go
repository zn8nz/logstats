// logstats processes logs and and prints the number of lines that contain a given regexp.
// Pass as arguments the options -o to define the order of number fields inside the timestamp and
// -t to set a time interval for grouping, ranging from 10 minutes to 1 day.
// Use -k <regexp> instead of -t to filter lines by regexp; for those lines count the main regexp matches.
package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"
)

const version = "1.4"

var (
	// command line parameters
	settings = struct {
		order    string
		groupBy  int
		rxCount  *regexp.Regexp
		rxKey    *regexp.Regexp
		progress bool
		duration bool
		version  bool
		offset   time.Duration
		split    string
		cutoff   int
	}{
		order:   "ymdhi",
		groupBy: 1,
	}

	counter = make(map[string]int)

	rx *regexp.Regexp
)

func main() {
	// parse options
	var key, offset string

	flag.StringVar(&settings.order, "o", "ymdhi", "order of the timestamp fields: y=year, m=month, d=day, h=hour, i=min, seconds ignored")
	flag.IntVar(&settings.groupBy, "t", 24, "valid intervals: 5, 10, 15, 20, 30 = minutes; 1, 2, 3, 6, 12, 24 = hours; 31 = month; 365 = year")
	flag.StringVar(&key, "k", "", "regexp that defines the key to group by; cannot use with -o and -t")
	flag.BoolVar(&settings.progress, "p", false, "print number of matches per file name")
	flag.BoolVar(&settings.duration, "d", false, "print duration and number of files")
	flag.BoolVar(&settings.version, "version", false, "print version number only")
	flag.StringVar(&offset, "ofs", "", "timestamp offset in a format like -1.5h +13h45.5m")
	flag.StringVar(&settings.split, "s", "", "split timestamp at position indicated by 'x', e.g. '....x..x..x' to split a continuous date '20160304' for parsing")
	flag.IntVar(&settings.cutoff, "cof", 25, "only look for timestamp in the beginning of the line, upto this number of characters")
	flag.Parse()

	if settings.version {
		fmt.Println(version)
		return
	}

	if offset != "" {
		var err error
		settings.offset, err = time.ParseDuration(offset)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invalid offset format, use h, m, e.g. -10h30.5m")
			return
		}
	}

	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "Usage: <options> <regexp> <glob>")
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
	t0 := time.Now()
	n, err := loopFiles(flag.Arg(1))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if settings.progress {
		fmt.Println("")
	}
	elapsed := time.Since(t0)
	if settings.duration {
		fmt.Printf("%v for %d files\n\n", elapsed, n)
	}

	// find max key length
	klen := 0
	for k := range counter {
		if len(k) > klen {
			klen = len(k)
		}
	}

	// fill array with output for sorting
	arr := make([]string, len(counter))
	i := 0
	for k, v := range counter {
		arr[i] = fmt.Sprintf("%-*s,%9d", klen, k, v)
		i++
	}
	// sort and then print
	sort.Strings(arr)
	for _, s := range arr {
		fmt.Println(s)
	}
}

func loopFiles(glob string) (n int, err error) {
	files, err := filepath.Glob(glob)
	if err != nil {
		return 0, err
	}
	for _, fn := range files {
		err = oneFile(fn)
		n++
	}
	return n, nil
}

func oneFile(fn string) error {
	file, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer file.Close()

	nmatch, err := loopLines(file)
	if err != nil {
		return err
	}

	if settings.progress {
		fmt.Printf("%9d in %s\n", nmatch, fn)
	}
	return nil
}

func loopLines(reader io.Reader) (int, error) {
	var key0 string
	var ts0 *time.Time
	var fresh bool
	var nmatch int
	cutoff := settings.cutoff

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if settings.rxKey != nil {
			match := settings.rxKey.FindString(line)
			if match == "" {
				continue
			}
			if settings.rxCount.MatchString(line) {
				nmatch++
				counter[match]++
			}
		} else {
			var key string
			c := len(line)
			if cutoff < c {
				c = cutoff
			}

			ts, err := parseTimestamp(line[0:c])
			if err != nil {
				ts = ts0
				key = key0
			} else {
				ts0 = ts
				fresh = true
			}
			if fresh && settings.rxCount.MatchString(line) {
				min := 0
				hour := ts.Hour()
				day := ts.Day()
				month := ts.Month()
				format := "2006-01-02 15:04"

				div := settings.groupBy
				switch div {
				case 5, 10, 15, 20, 30:
					min = div * int(ts.Minute()/div)
				case 1:
					// noop
				case 2, 3, 6, 12:
					hour = div * int(ts.Hour()/div)
				case 24:
					format = "2006-01-02"
					hour = 0
				case 31:
					format = "2006-01"
					day, hour = 1, 0
				case 365:
					format = "2006"
					month, day, hour = 0, 1, 0
				default:
					return 0, fmt.Errorf("Invalid value interval %d ", div)
				}
				tt := time.Date(ts.Year(), month, day, hour, min, 0, 0, ts.Location())
				key = tt.Format(format)

				nmatch++
				counter[key]++
				key0 = key
				fresh = false
				//fmt.Printf("%s|%s\n", tt.Format("2006-01-02 15:04:05"), line)
			}
		}
	}
	return nmatch, nil
}

// timestamp analyses a timestamp string and returns it as a *time.Time.
func parseTimestamp(ts string) (*time.Time, error) {
	if settings.split != "" {
		ts = split(ts, settings.split)
		if ts == "" {
			return nil, errors.New("no timestamp(1)")
		}
	}
	arr := rx.FindAllStringSubmatch(ts, len(settings.order))
	if len(arr) < len(settings.order) {
		return nil, errors.New("no timestamp(2)")
	}
	var year, month, day, hour, min int
	for i, s := range arr {
		c := settings.order[i]
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
		default:
			return nil, fmt.Errorf("Invalid timestamp characters in layout: [%s] use [-ymdhi]", string(c))
		}
	}
	if year < 100 {
		year += 2000
	}
	var t = time.Date(year, time.Month(month), day, hour, min, 0, 0, time.UTC).Add(settings.offset)
	return &t, nil
}

// split inserts a space in string 's' whereever a 'x' occurs in the string 'split'
// and returns the result. E.g. split("  20160406T225401|error", "......x..x.....x..x") ->
// "  2016 04 06T22 54 01|error". Any other character than a 'x' means: take corresponding
// character from 's'.
func split(s string, split string) string {
	if len(split) > len(s) {
		return ""
	}
	var buf bytes.Buffer
	var j int
	for i := 0; i < len(split); i++ {
		if split[i] == 'x' {
			buf.WriteByte(' ')
		} else {
			buf.WriteByte(s[j])
			j++
		}
	}
	buf.WriteString(s[j:])
	return buf.String()
}
