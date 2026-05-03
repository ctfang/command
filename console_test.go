package command

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilterPositionalArgs_legacyDoesNotEatNextToken(t *testing.T) {
	// 原规则：-b 不吞 bar，bar 仍为位置参数
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

func TestFilterPositionalArgs_doubleDash(t *testing.T) {
	got := filterPositionalArgs([]string{"-x=1", "a", "--", "-y", "b"})
	want := []string{"a", "-y", "b"}
	if len(got) != len(want) {
		t.Fatalf("len got %d want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d]=%q want %q", i, got[i], want[i])
		}
	}
}

func TestParsed_dashAgeSpaceDoesNotBindValue(t *testing.T) {
	c := testConsole()
	inp := Input{
		console:  c,
		Has:      map[string]bool{},
		Argument: map[string]string{},
		Option:   map[string][]string{},
	}
	cfg := Argument{
		Option: []ArgParam{{Name: "age", Default: "0", Description: ""}},
	}
	err := inp.Parsed(cfg, []string{"--age", "20"})
	if err != nil {
		t.Fatal(err)
	}
	if inp.GetOption("age") != "" {
		t.Fatalf("want empty age, got %q", inp.GetOption("age"))
	}
}

func TestParsed_longFormEqSetsOption(t *testing.T) {
	c := testConsole()
	inp := Input{
		console:  c,
		Has:      map[string]bool{},
		Argument: map[string]string{},
		Option:   map[string][]string{},
	}
	cfg := Argument{
		Option: []ArgParam{{Name: "age", Default: "0", Description: ""}},
	}
	err := inp.Parsed(cfg, []string{"--age=99"})
	if err != nil {
		t.Fatal(err)
	}
	if inp.GetOption("age") != "99" {
		t.Fatalf("got %q", inp.GetOption("age"))
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

// demoIniCmd 用于 Console + INI 集成测试
type demoIniCmd struct{}

func (demoIniCmd) Configure() Configure {
	return Configure{
		Name:        "demo",
		Description: "ini default",
		Input: Argument{
			Option: []ArgParam{
				{Name: "url", Description: "", Default: "fallback"},
				{Name: "port", Description: "", Default: "9000"},
			},
		},
	}
}

func (demoIniCmd) Execute(input Input) {}

func TestAddCommand_iniOverridesOptionDefault(t *testing.T) {
	dir := t.TempDir()
	iniPath := filepath.Join(dir, "app.ini")
	if err := os.WriteFile(iniPath, []byte("url=\"from-ini\"\nport=\"6000\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	c := New()
	c.SetConfig(iniPath)
	if err := c.IniConfig(); err != nil {
		t.Fatal(err)
	}
	c.AddCommand(demoIniCmd{})
	mc := c.MapCommand["demo"]
	opts := mc.CommandConfig.Input.Option
	if len(opts) < 2 {
		t.Fatalf("options: %d", len(opts))
	}
	if opts[0].Default != "from-ini" {
		t.Fatalf("url default got %q", opts[0].Default)
	}
	if opts[1].Default != "6000" {
		t.Fatalf("port default got %q", opts[1].Default)
	}
}

func TestParsed_missingRequiredArgument(t *testing.T) {
	c := testConsole()
	inp := Input{console: c, Has: map[string]bool{}, Argument: map[string]string{}, Option: map[string][]string{}}
	cfg := Argument{
		Argument: []ArgParam{{Name: "a", Description: ""}, {Name: "b", Description: ""}},
	}
	err := inp.Parsed(cfg, []string{"only-one"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParsed_commandLineOverridesIniDefault(t *testing.T) {
	c := testConsole()
	// 模拟 AddCommand 后的 Default 已由 ini 写成 from-ini
	inp := Input{console: c, Has: map[string]bool{}, Argument: map[string]string{}, Option: map[string][]string{}}
	cfg := Argument{
		Option: []ArgParam{{Name: "url", Default: "from-ini", Description: ""}},
	}
	if err := inp.Parsed(cfg, []string{"-url=cli"}); err != nil {
		t.Fatal(err)
	}
	if inp.GetOption("url") != "cli" {
		t.Fatalf("got %q", inp.GetOption("url"))
	}
}

func TestParsed_repeatOptionAppendsSlice(t *testing.T) {
	c := testConsole()
	inp := Input{console: c, Has: map[string]bool{}, Argument: map[string]string{}, Option: map[string][]string{}}
	cfg := Argument{
		Option: []ArgParam{{Name: "tag", Default: "", Description: ""}},
	}
	if err := inp.Parsed(cfg, []string{"-tag=a", "-tag=b"}); err != nil {
		t.Fatal(err)
	}
	got := inp.GetOptions("tag")
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("%#v", got)
	}
}

func TestParsed_IsDaemonDoubleHyphen(t *testing.T) {
	c := &Console{
		MapCommand: map[string]MapCommand{},
		baseOption: []ArgParam{},
		baseHas: []ArgParam{
			{Name: "-d", Description: ""},
			{Name: "--d", Description: ""},
		},
	}
	inp := Input{console: c, Has: map[string]bool{}, Argument: map[string]string{}, Option: map[string][]string{}}
	if err := inp.Parsed(Argument{}, []string{"--d"}); err != nil {
		t.Fatal(err)
	}
	if !inp.IsDaemon() {
		t.Fatal("IsDaemon should be true for --d")
	}
}

func TestParsed_mixedOptionAndPositionalOrder(t *testing.T) {
	c := testConsole()
	inp := Input{console: c, Has: map[string]bool{}, Argument: map[string]string{}, Option: map[string][]string{}}
	cfg := Argument{
		Argument: []ArgParam{{Name: "x", Description: ""}, {Name: "y", Description: ""}},
		Option:   []ArgParam{{Name: "n", Default: "0", Description: ""}},
	}
	if err := inp.Parsed(cfg, []string{"-n=5", "a", "b"}); err != nil {
		t.Fatal(err)
	}
	if inp.GetArgument("x") != "a" || inp.GetArgument("y") != "b" {
		t.Fatalf("%v %v", inp.GetArgument("x"), inp.GetArgument("y"))
	}
	if inp.GetOption("n") != "5" {
		t.Fatalf("n=%q", inp.GetOption("n"))
	}
}

func TestInput_GetFilePath(t *testing.T) {
	inp := Input{FilePath: "/tmp/prog"}
	if inp.GetFilePath() != "/tmp/prog" {
		t.Fatal(inp.GetFilePath())
	}
}

func TestGetOptions_unknownKeyEmptySlice(t *testing.T) {
	inp := Input{Option: map[string][]string{}}
	if len(inp.GetOptions("nope")) != 0 {
		t.Fatal()
	}
}

func TestGetArgument_unknownEmpty(t *testing.T) {
	inp := Input{Argument: map[string]string{}}
	if inp.GetArgument("nope") != "" {
		t.Fatal()
	}
}
