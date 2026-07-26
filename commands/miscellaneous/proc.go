package miscellaneous

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// To whomever is stupid enough, do not try to run this on a Windows server :)

func processRSSMB() (float64, error) {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			kb, err := strconv.ParseFloat(fields[1], 64)
			if err != nil {
				return 0, err
			}
			return kb / 1024, nil
		}
	}
	return 0, os.ErrNotExist
}

func processCPUPercent() (float64, error) {
	const window = 200 * time.Millisecond

	t1, err := processCPUTicks()
	if err != nil {
		return 0, err
	}
	time.Sleep(window)
	t2, err := processCPUTicks()
	if err != nil {
		return 0, err
	}

	const clkTck = 100.0 
	cpuSeconds := (t2 - t1) / clkTck
	return cpuSeconds / window.Seconds() * 100, nil
}

func processCPUTicks() (float64, error) {
	data, err := os.ReadFile("/proc/self/stat")
	if err != nil {
		return 0, err
	}
	s := string(data)
	end := strings.LastIndex(s, ")")
	if end == -1 || end+2 >= len(s) {
		return 0, os.ErrInvalid
	}
	fields := strings.Fields(s[end+2:])
	if len(fields) < 13 {
		return 0, os.ErrInvalid
	}
	utime, err := strconv.ParseFloat(fields[11], 64) 
	if err != nil {
		return 0, err
	}
	stime, err := strconv.ParseFloat(fields[12], 64) 
	if err != nil {
		return 0, err
	}
	return utime + stime, nil
}
