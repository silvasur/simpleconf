// Package simpleconf implements a parser for a very simple configuration file format.
//
// The format is inspired by INI files. A file consists of lines. A line is one of:
//
// Comment: The Line starts with an '#', or an ';'. Comments can only be on their own line.
//
// Section header: A section header names the current section. The Section named is wrapped in '[' and ']'.
// Section names must not be empty.
//
// Key-Value Pair: A value assigned to a key. The pair must belong to a section.
// Key and value are separated by '='. Everything after the '=' is the value.
// Keys must not be empty.
// Leading and trailing whitespace in key and value will be deleted.
//
// Example:
//
// 	[foo]
// 	test = Hello, World!
// 	answer = 42
//
// 	; I am a comment.
// 	# I am also a comment
// 	[bar]
// 	trololo = bla.. ; I am NOT A comment, I belong to value bar.trololo!
package simpleconf

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Config contains a loaded config file. You can access the values by simply using the map or by using the `Get...` functions.
type Config map[string]Section
type Section map[string]string

func (c Config) addSection(section Section, name string) {
	if section != nil {
		c[name] = section
	}
}

// Load loads a config file. See package description for the file format.
//
// If outerr != nil, either an I/O error occurred, or the file was not a valid config file.
// In both cases, the error will describe what went wrong.
func Load(r io.Reader) (config Config, outerr error) {
	config = make(Config)
	scanner := bufio.NewScanner(r)

	var section Section
	var sectName string

	for l := 1; scanner.Scan(); l++ {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}

		switch line[0] {
		case ';', '#':
			continue
		case '[':
			parts := strings.SplitN(line, "[", 2)
			parts = strings.SplitN(parts[1], "]", 2)

			if len(parts) != 2 {
				return nil, fmt.Errorf("Missing closing ']' in section header at line %d", l)
			}
			if len(parts[1]) != 0 {
				return nil, fmt.Errorf("More data after closing ']' at line %d", l)
			}
			if len(parts[0]) == 0 {
				return nil, fmt.Errorf("Empty section name at line %d", l)
			}

			config.addSection(section, sectName)
			section = make(Section)
			sectName = parts[0]
		default:
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("Couldn't neither find a comment, a section header nor a key-value pair at line %d", l)
			}

			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if len(key) == 0 {
				return nil, fmt.Errorf("Empty key at line %d", l)
			}

			if section == nil {
				return nil, fmt.Errorf("Found key-value pair, but no section at line %d", l)
			}

			section[key] = val
		}
	}

	config.addSection(section, sectName)

	outerr = scanner.Err()
	return
}

// Errors of the `Get...` functions.
var (
	NotFound = errors.New("Section or key not found.")
	NotBool  = errors.New("Could not interpret value as bool.")
)

// GetString gets the string assigned to [section] key. It will return NotFound, if no such key exists.
func (c Config) GetString(section, key string) (string, error) {
	s, ok := c[section]
	if !ok {
		return "", NotFound
	}
	rv, ok := s[key]
	if !ok {
		return "", NotFound
	}
	return rv, nil
}

// GetStringDefault is like GetString, but will return d, if the key was not found.
func (c Config) GetStringDefault(d, section, key string) (string, error) {
	rv, err := c.GetString(section, key)
	if err == NotFound {
		return d, nil
	}
	return rv, err
}

// GetInt is like GetString, but will additionally parse the value as an integer. See strconv.ParseInt for possible errors.
func (c Config) GetInt(section, key string) (int64, error) {
	s, err := c.GetString(section, key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(s, 10, 64)
}

func (c Config) GetIntDefault(d int64, section, key string) (int64, error) {
	rv, err := c.GetInt(section, key)
	if err == NotFound {
		return d, nil
	}
	return rv, err
}

// GetFloat is like GetString, but will additionally parse the value as a float. See strconv.ParseFloat for possible errors.
func (c Config) GetFloat(section, key string) (float64, error) {
	s, err := c.GetString(section, key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(s, 64)
}

func (c Config) GetFloatDefault(d float64, section, key string) (float64, error) {
	rv, err := c.GetFloat(section, key)
	if err == NotFound {
		return d, nil
	}
	return rv, err
}

// GetBool is like GetString, but will additionally parse the value as a boolean value.
//
// true, on, yes, y and 1 are all true, false, off, no, n, 0 are all false. Other values will result in a NotBool error.
func (c Config) GetBool(section, key string) (bool, error) {
	s, err := c.GetString(section, key)
	if err != nil {
		return false, err
	}
	switch strings.ToLower(s) {
	case "true", "on", "yes", "y", "1":
		return true, nil
	case "false", "off", "no", "n", "0":
		return false, nil
	default:
		return false, NotBool
	}
}

func (c Config) GetBoolDefault(d bool, section, key string) (bool, error) {
	rv, err := c.GetBool(section, key)
	if err == NotFound {
		return d, nil
	}
	return rv, err
}

// GetFile is like GetString, but will additionally open the file with that name.
// See os.OpenFile for flag and perm values and additional error values.
func (c Config) GetFile(flag int, perm os.FileMode, section, key string) (*os.File, error) {
	s, err := c.GetString(section, key)
	if err != nil {
		return nil, err
	}

	return os.OpenFile(s, flag, perm)
}

// GetFileReadonly is like GetFile, but with default values for flag and perm that will open the file in read-only mode.
func (c Config) GetFileReadonly(section, key string) (*os.File, error) {
	return c.GetFile(os.O_RDONLY, 0, section, key)
}
