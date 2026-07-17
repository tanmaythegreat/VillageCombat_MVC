package models

import (
	"reflect"
	"testing"
)

// ---------------------------------------------------------------------------
// pgQuote / pgBool
// ---------------------------------------------------------------------------

func TestPgQuote(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", `""`},
		{"simple", "abc", `"abc"`},
		{"with quote", `a"b`, `"a\"b"`},
		{"with backslash", `a\b`, `"a\\b"`},
		{"with quote and backslash", `a"\b`, `"a\"\\b"`},
		{"with comma and parens", `a,(b)`, `"a,(b)"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := pgQuote(c.in)
			if got != c.want {
				t.Errorf("pgQuote(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestPgBool(t *testing.T) {
	if pgBool(true) != "t" {
		t.Errorf("pgBool(true) = %q, want \"t\"", pgBool(true))
	}
	if pgBool(false) != "f" {
		t.Errorf("pgBool(false) = %q, want \"f\"", pgBool(false))
	}
}

// ---------------------------------------------------------------------------
// toScanString
// ---------------------------------------------------------------------------

func TestToScanString(t *testing.T) {
	s, err := toScanString("hello")
	if err != nil || s != "hello" {
		t.Errorf("toScanString(string) = (%q, %v), want (\"hello\", nil)", s, err)
	}

	s, err = toScanString([]byte("hello"))
	if err != nil || s != "hello" {
		t.Errorf("toScanString([]byte) = (%q, %v), want (\"hello\", nil)", s, err)
	}

	_, err = toScanString(42)
	if err == nil {
		t.Errorf("toScanString(int) expected error, got nil")
	}

	_, err = toScanString(nil)
	if err == nil {
		t.Errorf("toScanString(nil) expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// parsePgArrayElements
// ---------------------------------------------------------------------------

func strPtr(s string) *string { return &s }

func TestParsePgArrayElements_Empty(t *testing.T) {
	els, err := parsePgArrayElements("{}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(els) != 0 {
		t.Errorf("expected 0 elements, got %d", len(els))
	}
}

func TestParsePgArrayElements_Invalid(t *testing.T) {
	_, err := parsePgArrayElements("not-an-array")
	if err == nil {
		t.Errorf("expected error for malformed array literal, got nil")
	}
}

func TestParsePgArrayElements_SingleQuotedTuple(t *testing.T) {
	els, err := parsePgArrayElements(`{"(1,2,3)"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []*string{strPtr("(1,2,3)")}
	if !equalStringPtrSlices(els, want) {
		t.Errorf("got %v, want %v", derefAll(els), derefAll(want))
	}
}

func TestParsePgArrayElements_MultipleQuotedTuples(t *testing.T) {
	els, err := parsePgArrayElements(`{"(1,2,3)","(4,5,6)"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []*string{strPtr("(1,2,3)"), strPtr("(4,5,6)")}
	if !equalStringPtrSlices(els, want) {
		t.Errorf("got %v, want %v", derefAll(els), derefAll(want))
	}
}

func TestParsePgArrayElements_EscapedQuotesInsideTuple(t *testing.T) {
	// Composite literal containing an escaped inner double-quote and a comma,
	// e.g. a troop_id string that itself contains a quote character.
	els, err := parsePgArrayElements(`{"(\"a,b\",1,t,2,3,4.5)"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(els) != 1 {
		t.Fatalf("expected 1 element, got %d", len(els))
	}
	want := `("a,b",1,t,2,3,4.5)`
	if els[0] == nil || *els[0] != want {
		t.Errorf("got %v, want %q", derefAll(els), want)
	}
}

func TestParsePgArrayElements_NullElement(t *testing.T) {
	els, err := parsePgArrayElements(`{NULL,"(1,2,3)"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(els) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(els))
	}
	if els[0] != nil {
		t.Errorf("expected first element to be nil (SQL NULL), got %v", *els[0])
	}
	if els[1] == nil || *els[1] != "(1,2,3)" {
		t.Errorf("expected second element (1,2,3), got %v", derefAll(els))
	}
}

func equalStringPtrSlices(a, b []*string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if (a[i] == nil) != (b[i] == nil) {
			return false
		}
		if a[i] != nil && *a[i] != *b[i] {
			return false
		}
	}
	return true
}

func derefAll(s []*string) []string {
	out := make([]string, len(s))
	for i, p := range s {
		if p == nil {
			out[i] = "<nil>"
		} else {
			out[i] = *p
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// parseCompositeFields
// ---------------------------------------------------------------------------

func TestParseCompositeFields_Basic(t *testing.T) {
	fields, err := parseCompositeFields(`(1,2,t,3,4,5.5)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"1", "2", "t", "3", "4", "5.5"}
	if !reflect.DeepEqual(fields, want) {
		t.Errorf("got %v, want %v", fields, want)
	}
}

func TestParseCompositeFields_QuotedFieldWithComma(t *testing.T) {
	fields, err := parseCompositeFields(`("a,b",1,t,2,3,4.5)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"a,b", "1", "t", "2", "3", "4.5"}
	if !reflect.DeepEqual(fields, want) {
		t.Errorf("got %v, want %v", fields, want)
	}
}

func TestParseCompositeFields_EmptyTrailingFieldIsNull(t *testing.T) {
	// A trailing empty, unquoted field represents SQL NULL.
	fields, err := parseCompositeFields(`(1,2,)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"1", "2", ""}
	if !reflect.DeepEqual(fields, want) {
		t.Errorf("got %v, want %v", fields, want)
	}
}

func TestParseCompositeFields_Invalid(t *testing.T) {
	_, err := parseCompositeFields("1,2,3")
	if err == nil {
		t.Errorf("expected error for missing parens, got nil")
	}
}

func TestParseCompositeFields_EscapedBackslashAndQuote(t *testing.T) {
	fields, err := parseCompositeFields(`("a\\b\"c",1)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{`a\b"c`, "1"}
	if !reflect.DeepEqual(fields, want) {
		t.Errorf("got %v, want %v", fields, want)
	}
}

// ---------------------------------------------------------------------------
// TroopSpawnArray Value / Scan round-trip
// ---------------------------------------------------------------------------

func TestTroopSpawnArray_Value_Empty(t *testing.T) {
	var a TroopSpawnArray
	v, err := a.Value()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "{}" {
		t.Errorf("got %v, want \"{}\"", v)
	}
}

func TestTroopSpawnArray_ValueThenScan_RoundTrip(t *testing.T) {
	original := TroopSpawnArray{
		{
			TroopID:           "troop-1",
			TroopLevel:        3,
			SpawnedByAttacker: true,
			SpawnedAt_X:       10,
			SpawnedAt_Y:       -5,
			SpawnTime:         1.25,
		},
		{
			TroopID:           "troop,with\"special\\chars",
			TroopLevel:        1,
			SpawnedByAttacker: false,
			SpawnedAt_X:       0,
			SpawnedAt_Y:       0,
			SpawnTime:         0,
		},
	}

	v, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	strVal, ok := v.(string)
	if !ok {
		t.Fatalf("Value() returned non-string: %T", v)
	}

	var scanned TroopSpawnArray
	if err := scanned.Scan(strVal); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if !reflect.DeepEqual(original, scanned) {
		t.Errorf("round-trip mismatch.\n original: %+v\n scanned:  %+v", original, scanned)
	}
}

func TestTroopSpawnArray_Scan_Nil(t *testing.T) {
	a := TroopSpawnArray{{TroopID: "x"}}
	if err := a.Scan(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a != nil {
		t.Errorf("expected nil after scanning nil, got %v", a)
	}
}

func TestTroopSpawnArray_Scan_WrongFieldCount(t *testing.T) {
	var a TroopSpawnArray
	err := a.Scan(`{"(1,2,3)"}`)
	if err == nil {
		t.Errorf("expected error for wrong field count, got nil")
	}
}

func TestTroopSpawnArray_Scan_InvalidNumericField(t *testing.T) {
	var a TroopSpawnArray
	err := a.Scan(`{"(id1,notanint,t,1,2,3.0)"}`)
	if err == nil {
		t.Errorf("expected error for invalid troop_level, got nil")
	}
}

// ---------------------------------------------------------------------------
// InitialBuildingArray Value / Scan round-trip
// ---------------------------------------------------------------------------

func TestInitialBuildingArray_Value_Empty(t *testing.T) {
	var a InitialBuildingArray
	v, err := a.Value()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "{}" {
		t.Errorf("got %v, want \"{}\"", v)
	}
}

func TestInitialBuildingArray_ValueThenScan_RoundTrip(t *testing.T) {
	original := InitialBuildingArray{
		{BuildingID: "building-1", Grid_X: 1, Grid_Y: 2, Level: 5, IsBroken: true},
		{BuildingID: "building,\"2", Grid_X: -1, Grid_Y: 0, Level: 0, IsBroken: false},
	}

	v, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	strVal, ok := v.(string)
	if !ok {
		t.Fatalf("Value() returned non-string: %T", v)
	}

	var scanned InitialBuildingArray
	if err := scanned.Scan(strVal); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if !reflect.DeepEqual(original, scanned) {
		t.Errorf("round-trip mismatch.\n original: %+v\n scanned:  %+v", original, scanned)
	}
}

func TestInitialBuildingArray_Scan_Nil(t *testing.T) {
	a := InitialBuildingArray{{BuildingID: "x"}}
	if err := a.Scan(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a != nil {
		t.Errorf("expected nil after scanning nil, got %v", a)
	}
}

func TestInitialBuildingArray_Scan_WrongFieldCount(t *testing.T) {
	var a InitialBuildingArray
	err := a.Scan(`{"(1,2,3)"}`)
	if err == nil {
		t.Errorf("expected error for wrong field count, got nil")
	}
}

func TestInitialBuildingArray_Scan_InvalidIntField(t *testing.T) {
	var a InitialBuildingArray
	err := a.Scan(`{"(id1,notanint,2,3,t)"}`)
	if err == nil {
		t.Errorf("expected error for invalid grid_x, got nil")
	}
}

func TestInitialBuildingArray_Scan_NullElementSkipped(t *testing.T) {
	var a InitialBuildingArray
	err := a.Scan(`{NULL,"(building-1,1,2,3,f)"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a) != 1 {
		t.Fatalf("expected 1 element (NULL skipped), got %d", len(a))
	}
	if a[0].BuildingID != "building-1" || a[0].Grid_X != 1 || a[0].Grid_Y != 2 || a[0].Level != 3 || a[0].IsBroken != false {
		t.Errorf("unexpected parsed value: %+v", a[0])
	}
}

func TestInitialBuildingArray_Scan_BytesInput(t *testing.T) {
	// Simulate a driver returning []byte instead of string.
	var a InitialBuildingArray
	err := a.Scan([]byte(`{"(building-1,1,2,3,t)"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a) != 1 || a[0].BuildingID != "building-1" {
		t.Errorf("unexpected result: %+v", a)
	}
}
