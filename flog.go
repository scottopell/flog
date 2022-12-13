package flog

import (
	"compress/gzip"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var id = time.Now().UnixNano()

// Generate generates the logs with given options
func Generate(option *Option) error {
	var (
		splitCount = 1
		created    = time.Now()

		interval time.Duration
		delay    time.Duration
	)

	buildCache(option.Bytes)

	if option.Sleep > 0 {
		interval = option.Sleep
	}

	rand.Seed(time.Now().UnixNano())

	logFileName := option.Output
	writer, err := NewWriter(option.Type, logFileName)
	if err != nil {
		return err
	}

	var counter uint64 = 0

	if option.Forever {
		rate := option.Rate
		for {
			rate += option.Increment
			start := time.Now()
			for i := 0; i < rate; i++ {
				log := NewLog(option.Format, created, option.Bytes)
				counter++
				if option.Seq {
					log = writeSeq(counter, log)
				}
				_, _ = writer.Write([]byte(log + "\n"))
				created = created.Add(interval)
				if option.Rotate > 0 && counter%uint64(option.Rotate) == 0 {
					writer, _ = RotateFile(writer, logFileName)
				}
			}
			elapsed := time.Since(start)
			time.Sleep(time.Second - elapsed)
		}
	}

	// TODO : fix below
	if option.Bytes == 0 {
		// Generates the logs until the certain number of lines is reached
		for line := 0; line < option.Number; line++ {
			time.Sleep(delay)
			log := NewLog(option.Format, created, option.Bytes)
			_, _ = writer.Write([]byte(log + "\n"))

			if (option.Type != "stdout") && (option.SplitBy > 0) && (line > option.SplitBy*splitCount) {
				_ = writer.Close()
				fmt.Println(logFileName, "is created.")

				logFileName = NewSplitFileName(option.Output, splitCount)
				writer, _ = NewWriter(option.Type, logFileName)

				splitCount++
			}
			created = created.Add(interval)
		}
	} else {
		// Generates the logs until the certain size in bytes is reached
		bytes := 0
		for bytes < option.Bytes {
			time.Sleep(delay)
			log := NewLog(option.Format, created, option.Bytes)
			_, _ = writer.Write([]byte(log + "\n"))

			bytes += len(log)
			if (option.Type != "stdout") && (option.SplitBy > 0) && (bytes > option.SplitBy*splitCount+1) {
				_ = writer.Close()
				fmt.Println(logFileName, "is created.")

				logFileName = NewSplitFileName(option.Output, splitCount)
				writer, _ = NewWriter(option.Type, logFileName)

				splitCount++
			}
			created = created.Add(interval)
		}
	}

	if option.Type != "stdout" {
		_ = writer.Close()
		fmt.Println(logFileName, "is created.")
	}
	return nil
}

func RotateFile(currentWriter io.WriteCloser, logFileName string) (io.WriteCloser, error) {
	err := currentWriter.Close()
	if err != nil {
		return nil, err
	}
	os.Remove(logFileName + ".1")
	os.Rename(logFileName, logFileName+".1")
	return NewWriter("log", logFileName)
}

// NewWriter returns a closeable writer corresponding to given log type
func NewWriter(logType string, logFileName string) (io.WriteCloser, error) {
	switch logType {
	case "stdout":
		return os.Stdout, nil
	case "log":
		logFile, err := os.Create(logFileName)
		if err != nil {
			return nil, err
		}
		return logFile, nil
	case "gz":
		logFile, err := os.Create(logFileName)
		if err != nil {
			return nil, err
		}
		return gzip.NewWriter(logFile), nil
	default:
		return nil, nil
	}
}

// NewLog creates a log for given format
func NewLog(format string, t time.Time, length int) string {
	switch format {
	case "app_log":
		return NewAppLog(t, length)
	case "apache_common":
		return NewApacheCommonLog(t)
	case "apache_combined":
		return NewApacheCombinedLog(t)
	case "apache_error":
		return NewApacheErrorLog(t, length)
	case "rfc3164":
		return NewRFC3164Log(t, length)
	case "rfc5424":
		return NewRFC5424Log(t, length)
	case "common_log":
		return NewCommonLogFormat(t)
	case "json":
		return NewJSONLogFormat(t, length)
	default:
		return ""
	}
}

// NewSplitFileName creates a new file path with split count
func NewSplitFileName(path string, count int) string {
	logFileNameExt := filepath.Ext(path)
	pathWithoutExt := strings.TrimSuffix(path, logFileNameExt)
	return pathWithoutExt + strconv.Itoa(count) + logFileNameExt
}

func writeSeq(counter uint64, log string) string {
	seq := fmt.Sprintf(" log_seq:%d:%d:log_seq", id, counter)
	return log[:len(log)-len(seq)] + seq
}
