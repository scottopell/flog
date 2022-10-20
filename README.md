# Flog

## About this Fork

This [Flog fork](https://github.com/mingrammer/flog) is for testing and benchmarking the Datadog agent's logs agent. Any modifications made to this fork are to better fit our testing and performance scenarios. 

This fork is not compatible with the upstream version and adds several breaking configuration changes and features. Below is a non-exhaustive list of notable differences: 

- `./flog` with no arguments will:
  - emit a single line
  - emit a `golang` style log

- `-b --bytes` will control how many bytes each log line contains instead of the total bytes emitted. 

- `-r --rate` will control how many logs per second are emitted. 
  - Can be combined with `-b` to control exactly how many bytes per second are emitted
  - Log rate is controlled in batches. so `-r 1000 -b 1024` will emit `1MiB/sec` of logs in chunks of `1000` logs at a time. 

- `-q --seq` embeds a sequence number in the log
  - When combined with `-b` it will overwrite the end of the log to maintain the exact byte size. 
  - Is in the format: `log_seq:<ID>:<LOG_COUNT>:log_seq`
  - Currently only works with the `-l` flag. 

- `-r --rotate` emulates file rotations.
  - Only when `-l log` is set.

CPU footprint is greatly reduced by pre-computing all the logs ahead of time. (this may cause a startup delay for very large amount of logs). 

As a result maximum throughput is significantly higher using less resources: 

(Mac M1 Max)
```
go run . -l -b 1024 -r 1000000 | pv > /dev/null
9.42GiB 0:00:20 [ 500MiB/s]
```
## Usage

There are useful options. (`flog --help`)

```console
Options:
  -f, --format string      log format. available formats:
						   - app_log (default)
                           - apache_common 
                           - apache_combine
                           - apache_error
                           - rfc3164
                           - rfc5424
                           - json
  -o, --output string      output filename. Path-like is allowed. (default "generated.log")
  -t, --type string        log output type. available types:
                           - stdout (default)
                           - log
                           - gz
  -q  --seq integer        add sequence number to logs (only when using -l)
  -n, --number integer     number of lines to generate.
  -b, --bytes integer      length of each log line in bytes (default 512)
  -s, --sleep duration     fix creation time interval for each log (default unit "seconds"). It does not actually sleep.
                           examples: 10, 20ms, 5s, 1m
  -r, --rate rate          # of logs per second
                           examples: 10, 20ms, 5s, 1m
  -p, --split-by integer   set the maximum number of lines or maximum size in bytes of a log file.
                           with "number" option, the logs will be split whenever the maximum number of lines is reached.
                           with "byte" option, the logs will be split whenever the maximum size in bytes is reached.
  -w, --overwrite          overwrite the existing log files.
  -l, --loop               loop output forever until killed.
  -a  --rotate             rotate log after x logs (only in log mode)
```
