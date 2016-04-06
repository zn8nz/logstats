# logstats
Search through logs with timestamped entries, creating summary of occurrences of a regexp per line, per time interval.

## version
Version 1.0, 2016-03-16

## parameters

`./logstats <options> <regexp> <glob>`

`-t`: _time unit_: grouping interval; valid values: 10, 15, 30 = minutes; 1, 2, 3, 6, 12, 24 = hours; 31 = month; 365 = year.

`-o` _order_: the order of the fields in the timestamps, by default "ymdhisf": i=minutes, f=fraction of seconds.

`-k` _regexp_: regexp to filter and group by.

`-p`: print all the file names and the number of matches per file name.

`-d`: print the duration of execution and number of files

`-ofs`: timestamp pos/neg offset: e.g. `-ofs -5h45m10s`. Any *Go* duration format is accepted.

`-s`: use to split continuous timestamp formats by inserting a space at position indicated with `x`, e.g. `20161204T125901.999` can be split with `-s ....x..x.....x..x`

`-version`: print version number

## examples with -t

Count occurrences of "error" in all files that match data/log*.txt, group by 30 min intervals, 
output for values > 0.

`./logstats -t 30 "error" data/log*.txt`

output e.g.

```
2016-02-28 12:00:00,        1
2016-02-28 12:30:00,        2
2016-02-29 05:30:00,        1
2016-02-29 14:00:00,        1
2016-02-29 15:00:00,        1
```

Count occurrences of "error", ignore case, in all files *.txt that that have a line layout like:

`000034|03|03/27/2016 10:20:59.114|error#983 stacktrace..qworwor woriweorroie rwoi ruo`

E.g. with some line number and thread number before the timestamp. We can skip these numbers by using a one "-" per number.
in the `-o` option. If the date format is month/day/year, we can indicate that with "mdy" as part of the `-o` option.

`./logstat -t 1 -o "--mdyhisf" "(?i:error)" *.txt`


## examples with -k

Count occurrences of "error" in all files in current folder that match *.log. Only consider lines
that start with "2016-03-21", "2016-03-22", "2016-03-23". Group by each unique match of -k regexp.
Output values > 0

`./logstats -k "^2016-03-2[123]" "error" *.log`

output e.g.
```
2016-02-21,        4 
2016-02-23,        2
```

Count all lines that contain `|warn|`, `|error|`, `|fatal|` and group by these three. The regexp
"." matches anything, but another regexp could filter.

`./logstats -k "\|warn\||\|error\||\|fatal\|" "." *.log`

output e.g.
```
warn ,       23 
error,        2
fatal,        1
```

The effect of -k can be described in a pseudo SQL as:
```sql
SELECT k, count(*) 
FROM all_lines a
WHERE a matches k AND a matches regexp
GROUP BY k
```

## examples with -p and -d
`-p` prints the files and how many occurrences of the search pattern
`-d` prints the duration
```
> ./logstats -p -d -t 3 "house" ../data/log000?.txt
        0 in ../data/log0001.txt
        0 in ../data/log0002.txt
        0 in ../data/log0003.txt
        1 in ../data/log0004.txt
        0 in ../data/log0005.txt
        0 in ../data/log0006.txt
        1 in ../data/log0007.txt
        2 in ../data/log0008.txt
        0 in ../data/log0009.txt

552.0397ms for 9 files

2016-03-03 15:00,        1
2016-03-03 18:00,        2
2016-03-03 21:00,        1
```

# known bugs / to do
1. Log entries that span multiple lines may not be handled correctly, as logstats consideres each individual line.
Lines that show up with dates like 1999 and 2000 in the output are a symptom of this bug.

