//go:build small

package question

import (
	"reflect"
	"testing"
)

func TestRecords(t *testing.T) {
	tests := []struct {
		name string
		data string
		want []Record
	}{
		{name: "when Records with questions and workflow issues then returns them in order", data: "{\"issue\":\"u1\",\"kind\":\"question\",\"task\":\"a\"}\n{\"issue\":\"u2\",\"kind\":\"workflow\",\"task\":\"main\"}\n", want: []Record{{Issue: "u1", Kind: Kind("question"), Task: "a"}, {Issue: "u2", Kind: Kind("workflow"), Task: "main"}}},
		{name: "when Records with an unknown kind or a broken line then returns the others", data: "{\"issue\":\"u1\",\"kind\":\"other\"}\n{\"issue\":\"u2\",\"kind\":\"question\"}\n{\"issue\":\"u3\",\"ki", want: []Record{{Issue: "u2", Kind: Kind("question")}}},
		{name: "when Records with no data then returns nothing", data: "", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Records([]byte(tt.data))

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestQuery(t *testing.T) {
	tests := []struct {
		name    string
		records []Record
		want    string
	}{
		{name: "when Query with records then returns one resource per record by its URL", records: []Record{{Issue: "https://github.example/issues/1", Kind: Kind("question")}, {Issue: "https://github.example/issues/2", Kind: Kind("workflow")}}, want: "query { i0: resource(url: \"https://github.example/issues/1\") { ... on Issue { state } } i1: resource(url: \"https://github.example/issues/2\") { ... on Issue { state } } }"},
		{name: "when Query with no records then returns a query asking nothing", records: nil, want: "query { __typename }"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Query(tt.records)

			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOpened(t *testing.T) {
	tests := []struct {
		name     string
		records  []Record
		response string
		want     Open
		wantErr  bool
	}{
		{name: "when Opened with open issues then returns them by kind", records: []Record{{Issue: "u1", Kind: Kind("question")}, {Issue: "u2", Kind: Kind("workflow")}}, response: `{"data": {"i0": {"state": "OPEN"}, "i1": {"state": "OPEN"}}}`, want: Open{Questions: []string{"u1"}, Workflows: []string{"u2"}}},
		{name: "when Opened with a closed issue then returns it not", records: []Record{{Issue: "u1", Kind: Kind("question")}, {Issue: "u2", Kind: Kind("question")}}, response: `{"data": {"i0": {"state": "CLOSED"}, "i1": {"state": "OPEN"}}}`, want: Open{Questions: []string{"u2"}}},
		{name: "when Opened with an issue not found then returns it not", records: []Record{{Issue: "u1", Kind: Kind("question")}}, response: `{"data": {"i0": null}}`, want: Open{}},
		{name: "when Opened with an issue recorded twice then returns it once", records: []Record{{Issue: "u1", Kind: Kind("question")}, {Issue: "u1", Kind: Kind("question")}}, response: `{"data": {"i0": {"state": "OPEN"}, "i1": {"state": "OPEN"}}}`, want: Open{Questions: []string{"u1"}}},
		{name: "when Opened with no records and the answer to an empty query then returns nothing", records: nil, response: `{"data": {"__typename": "Query"}}`, want: Open{}},
		{name: "when Opened with a response that is not JSON then returns an error", records: nil, response: `{`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Opened(tt.records, []byte(tt.response))

			if (err != nil) != tt.wantErr || !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, %v, want %#v, error %v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}
