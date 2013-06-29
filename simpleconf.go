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

type Config map[string]Section
type Section map[string]string

func (c Config) addSection(section Section, name string) {
	if section != nil {
		c[name] = section
	}
}

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

var (
	NotFound = errors.New("Section or key not found.")
	NotBool  = errors.New("Could not interpret value as bool.")
)

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

func (c Config) GetStringDefault(d, section, key string) (string, error) {
	rv, err := c.GetString(section, key)
	if err == NotFound {
		return d, nil
	}
	return rv, err
}

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

func (c Config) GetFile(flag int, perm os.FileMode, section, key string) (*os.File, error) {
	s, err := c.GetString(section, key)
	if err != nil {
		return nil, err
	}

	return os.OpenFile(s, flag, perm)
}

func (c Config) GetFileReadonly(section, key string) (*os.File, error) {
	return c.GetFile(os.O_RDONLY, 0, section, key)
}
