package util

import (
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"time"
)

// DigitsOnly checks if sitrng s is a number
func DigitsOnly(s string) bool {
	re, _ := regexp.Compile("^\\d+$")
	return re.MatchString(s)
}

// Exists checks if string x can be found in array a
func Exists(x string, a []string) bool {
	for _, s := range a {
		if s == x {
			return true
		}
	}
	return false
}

// DaysSince returns elapsed time since date as string
func DaysSince(d string) time.Duration {
	if t, err := time.Parse(time.RubyDate, d); err == nil {
		return time.Since(t)
	}
	return 0
}

// DotNucollPath returns the location of the configuration file
// with CGO_ENABLED=0 golang throws user: Current not implemented on linux/amd64
// so we look for alternatives including $HOME and current directory
func DotNucollPath() (string, error) {
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir, nil
	}
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(ex), nil
}
