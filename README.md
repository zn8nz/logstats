# logstats
Search through logs with timestamped entries, creating summary of occurrences of a regexp per line, per time interval.

##examples with -t

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
in the -o option. If the date format is month/day/year, we can indicate that with "mdy" as part of the -o option.

`./logstat -t 1 -o "--mdyhisf" "(?i:error)" *.txt`


##examples with -k

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
warn,       23 
error,        2
fatal,        1
```

##parameters

`./logstats <options> <regexp> <glob>`

`-p`: 10 = ten minutes, 15 = 15 minutes, 30 = half hour, 1 = one hour, 2 = two hour, 3 = 3 hour, 6 = 6 hour, 12 = 12 hour, 0 = 1 day intervals.

`-o`: the order of the fields in the timestamps, by default "ymdhisf": i=minutes, f=fraction of seconds.

`-k`: regexp to filter and group by.

The effect of -k can be described in a pseudo SQL as:
```sql
SELECT k, count(*) 
FROM all_lines a
WHERE a matches k AND a matches regexp
GROUP BY k
```

# known bugs / to do
1. Log entries that span multiple lines may not be handled correctly, as logstats consideres each individual line.
Lines that show up with dates like 1999 and 2000 in the output are a symptom of this bug.
2. Output not aligned with `-k`, if the key regexp matches different length strings.
3. Timestamps without separators between the fields, e.g. `20160306-125959.3` cannot be parsed correctly as such with the `-o` option. As a workaround for now use the `-k` with a string match. I think I will solve this by an alternative to `-o`: `-f "yyyymmddhhiissf"`
