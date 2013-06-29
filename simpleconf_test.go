package simpleconf

import (
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	r := strings.NewReader(`; I am just a comment.
    # Me too!
[foo]
 a = Hello, World!    
b=1337
c= on
[bar]
trololo			= 1.5
file =     /dev/zero`)

	conf, err := Load(r)
	if err != nil {
		t.Fatalf("Could not read config: %s", err)
	}

	if s, err := conf.GetString("foo", "a"); err != nil || s != "Hello, World!" {
		t.Errorf("Unexpected return for [foo] a: `%s`, %s", s, err)
	}

	if i, err := conf.GetInt("foo", "b"); err != nil || i != 1337 {
		t.Errorf("Unexpected return for [foo] b: %d, %s", i, err)
	}

	if b, err := conf.GetBool("foo", "c"); err != nil || b != true {
		t.Errorf("Unexpected return for [foo] c: %s, %s", b, err)
	}

	if f, err := conf.GetFloat("bar", "trololo"); err != nil || f != 1.5 {
		t.Errorf("Unexpected return for [bar] trololo: %f, %s", f, err)
	}

	f, err := conf.GetFileReadonly("bar", "file")
	if err != nil {
		t.Fatalf("Error while reading [bar] trololo: %s", err)
	}
	defer f.Close()

	if f.Name() != "/dev/zero" {
		t.Errorf("Unexpected file opened: %s", f.Name())
	}

	if s, err := conf.GetStringDefault("?", "baz", "bla"); err != nil || s != "?" {
		t.Errorf("Unexpected return for [baz] bla: `%s`, %s", s, err)
	}
}
