package command

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilterPositionalArgs(t *testing.T) {
	got := filterPositionalArgs([]string{"-a=1", "foo", "-b", "bar", "baz"})
	want := []string{"foo", "bar", "baz"}
	if len(got) != len(want) {
		t.Fatalf("len got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d]=%q want %q", i, got[i], want[i])
		}
	}
}

func TestArgTokenSet(t *testing.T) {
	m := argTokenSet([]string{"a", "b", "a"})
	if _, ok := m["a"]; !ok {
		t.Fatal("expected a")
	}
	if _, ok := m["b"]; !ok {
		t.Fatal("expected b")
	}
	if len(m) != 2 {
		t.Fatalf("unique tokens: got %d want 2", len(m))
	}
}

// testConsole 无 -h 回调，避免单测中触发 os.Exit
func testConsole() *Console {
	return &Console{
		MapCommand: map[string]MapCommand{},
		baseOption: []ArgParam{},
		baseHas: []ArgParam{
			{Name: "-d", Description: "daemon"},
		},
	}
}

func TestParsed_baseHasAndHas(t *testing.T) {
	c := testConsole()
	inp := Input{
		console:  c,
		Has:      map[string]bool{},
		Argument: map[string]string{},
		Option:   map[string][]string{},
	}
	cfg := Argument{
		Has: []ArgParam{{Name: "one", Description: ""}},
	}
	err := inp.Parsed(cfg, []string{"-d", "one"})
	if err != nil {
		t.Fatal(err)
	}
	if !inp.GetHas("-d") || !inp.GetHas("one") {
		t.Fatalf("Has: %#v", inp.Has)
	}
}

func TestParsedOptions_multiDefaultSameName(t *testing.T) {
	c := testConsole()
	inp := Input{
		console:  c,
		Has:      map[string]bool{},
		Argument: map[string]string{},
		Option:   map[string][]string{},
	}
	cfg := Argument{
		Option: []ArgParam{
			{Name: "age", Default: "18"},
			{Name: "age", Default: "24"},
		},
	}
	inp.ParsedOptions(cfg, []string{})
	got := inp.GetOptions("age")
	if len(got) != 2 || got[0] != "18" || got[1] != "24" {
		t.Fatalf("age options: %#v", got)
	}
}

func TestIniLoad_typesAndRawStringFallback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.ini")
	content := "; c\n[sec]\n" +
		"k=42\n" +
		"f=3.14\n" +
		"b=true\n" +
		"raw=notanumber\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	var i ini
	if err := i.Load(path); err != nil {
		t.Fatal(err)
	}
	if i.GetInt("sec.k", 0) != 42 {
		t.Fatalf("int got %v", i.GetInt("sec.k", 0))
	}
	if i.GetString("sec.raw", "") != "notanumber" {
		t.Fatalf("raw string got %q", i.GetString("sec.raw", ""))
	}
	if !i.GetBool("sec.b", false) {
		t.Fatal("expected bool true")
	}
	f := i.config["sec.f"].(float64)
	if f < 3.13 || f > 3.15 {
		t.Fatalf("float got %v", f)
	}
}

func TestIniLoad_missingFileIsOK(t *testing.T) {
	var i ini
	err := i.Load(filepath.Join(t.TempDir(), "none.ini"))
	if err != nil {
		t.Fatal(err)
	}
}
