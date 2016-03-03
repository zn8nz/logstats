# logstats
Search through logs with timestamped entries, creating summary of occurrences of a regexp per line, per time interval.

##example

`./logstats -p 30 "error" data/log*.txt`

```
2016-02-28 12:00:00,        1
2016-02-28 12:30:00,        2
2016-02-29 05:30:00,        1
2016-02-29 14:00:00,        1
2016-02-29 15:00:00,        1
```

##parameters

`./logstats <options> <regexp> <glob>`

-p: 10 = ten minutes, 15 = 15 minutes, 30 = half hour, 1 = one hour, 2 = two hour, 3 = 3 hour, 6 = 6 hour, 12 = 12 hour, 0 = 1 day intervals.

-o: the order of the fields in the timestamps, by default "ymdhisf": i=minutes, f=fraction of seconds.

