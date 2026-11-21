//go:build small

package main

import (
	"reflect"
	"testing"
)

func TestAnnotate(t *testing.T) {
	tests := []struct {
		name string
		path string
		text string
		want Annotation
	}{
		{name: "when Annotate with a matching sha then returns its hash only", path: "a.feature", text: "Feature: F\n\n  @sha:75928eaccc55\n  Scenario: A\n    Given x\n", want: Annotation{Hashes: []string{"a.feature:4: @sha:75928eaccc55"}}},
		{name: "when Annotate with no sha then returns it missing and annotated", path: "a.feature", text: "Feature: F\n\n  Scenario: A\n    Given x\n", want: Annotation{Missing: []string{"a.feature:3: Scenario: A"}, Hashes: []string{"a.feature:3: @sha:75928eaccc55"}, Changed: true, Annotated: "Feature: F\n\n  @sha:75928eaccc55\n  Scenario: A\n    Given x\n"}},
		{name: "when Annotate with a skipped scenario then returns it annotated but not missing", path: "a.feature", text: "Feature: F\n\n  @skip\n  Scenario: A\n    Given x\n", want: Annotation{Hashes: []string{"a.feature:4: @sha:75928eaccc55"}, Changed: true, Annotated: "Feature: F\n\n  @skip\n  @sha:75928eaccc55\n  Scenario: A\n    Given x\n"}},
		{name: "when Annotate with a stale sha then returns it replaced", path: "a.feature", text: "Feature: F\n\n  @sha:000000000000 @slow\n  Scenario: A\n    Given x\n", want: Annotation{Missing: []string{"a.feature:4: Scenario: A"}, Hashes: []string{"a.feature:4: @sha:75928eaccc55"}, Changed: true, Annotated: "Feature: F\n\n  @slow\n  @sha:75928eaccc55\n  Scenario: A\n    Given x\n"}},
		{name: "when Annotate with comments and blank lines in a scenario then returns the same hash", path: "a.feature", text: "Feature: F\n\n  @sha:75928eaccc55\n  Scenario: A\n    # note\n\n    Given x\n", want: Annotation{Hashes: []string{"a.feature:4: @sha:75928eaccc55"}}},
		{name: "when Annotate with a tag before the background then returns it not taken by the scenario", path: "a.feature", text: "Feature: F\n  @sha:75928eaccc55\n  Background:\n    Given y\n  Scenario: A\n    Given x\n", want: Annotation{Missing: []string{"a.feature:5: Scenario: A"}, Hashes: []string{"a.feature:5: @sha:75928eaccc55"}, Changed: true, Annotated: "Feature: F\n  @sha:75928eaccc55\n  Background:\n    Given y\n  @sha:75928eaccc55\n  Scenario: A\n    Given x\n"}},
		{name: "when Annotate with no scenario then returns nothing", path: "a.feature", text: "Feature: F\n", want: Annotation{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Annotate(tt.path, tt.text)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
