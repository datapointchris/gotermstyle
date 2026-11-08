package tuitheme

import (
	"encoding/json"
	"maps"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

type schemaGroup struct {
	Properties    map[string]json.RawMessage `json:"properties"`
	PropertyNames struct {
		Pattern string `json:"pattern"`
	} `json:"propertyNames"`
}

type schemaFile struct {
	Properties  map[string]schemaGroup `json:"properties"`
	Definitions struct {
		Color struct {
			Pattern string `json:"pattern"`
		} `json:"color"`
		Style struct {
			Properties map[string]json.RawMessage `json:"properties"`
		} `json:"style"`
	} `json:"definitions"`
}

// A role added to the schema, such as a ui.cursor, needs a value in default.json
// along its dots, and a value form the schema allows needs a reader that takes it.
func TestTheSchemaAndTheReaderAgreeOnEveryGroupRoleAndColor(t *testing.T) {
	data, err := os.ReadFile("schema/v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema schemaFile
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}

	schemaGroups := slices.Sorted(maps.Keys(schema.Properties))
	schemaGroups = slices.DeleteFunc(schemaGroups, func(g string) bool { return g == "$schema" })
	if got := slices.Sorted(slices.Values(groups)); !slices.Equal(got, schemaGroups) {
		t.Errorf("the reader reads groups %v, the schema declares %v", got, schemaGroups)
	}

	for _, group := range schemaGroups {
		for key := range schema.Properties[group].Properties {
			if _, ok := lookup(builtIn, group+"."+key); !ok {
				t.Errorf("%s.%s has no value in default.json, not even along its dots", group, key)
			}
		}
	}

	for role := range builtIn {
		group, key, _ := strings.Cut(role, ".")
		declared := schema.Properties[group]
		if _, ok := declared.Properties[key]; ok {
			continue
		}
		if p := declared.PropertyNames.Pattern; p != "" && regexp.MustCompile(p).MatchString(key) {
			continue
		}
		t.Errorf("default.json sets %s, which the schema does not allow", role)
	}

	var fields []string
	for _, f := range reflect.VisibleFields(reflect.TypeFor[styleObject]()) {
		fields = append(fields, f.Tag.Get("json"))
	}
	slices.Sort(fields)
	if want := slices.Sorted(maps.Keys(schema.Definitions.Style.Properties)); !slices.Equal(fields, want) {
		t.Errorf("the reader reads style fields %v, the schema declares %v", fields, want)
	}

	color := regexp.MustCompile(schema.Definitions.Color.Pattern)
	probes := []string{"#2d4f67", "#ABCDEF", "#fff", "#2d4f6", "#2d4f677", "#gggggg", "2d4f67", "16", "07", "-1", "+5", "", "red"}
	for n := range 16 {
		probes = append(probes, strconv.Itoa(n))
	}
	for _, s := range probes {
		_, err := parseColor(s)
		if accepted, allowed := err == nil, color.MatchString(s); accepted != allowed {
			t.Errorf("parseColor(%q) accepts it: %v; the schema allows it: %v", s, accepted, allowed)
		}
	}
}
