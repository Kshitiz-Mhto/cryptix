package utility

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type byLen []string

func (a byLen) Len() int {
	return len(a)
}
func (a byLen) Less(i, j int) bool {
	return len(a[i]) > len(a[j])
}
func (a byLen) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

type OutputWriter struct {
	Keys       []string
	Labels     []string
	Values     [][]string
	TempValues []string
}

// NewOutputWriter builds a new OutputWriter
func NewOutputWriter() *OutputWriter {
	ret := OutputWriter{}
	return &ret
}

// NewOutputWriterWithMap builds a new OutputWriter and automatically
// inserts the supplied map as a single line
func NewOutputWriterWithMap(data map[string]string) *OutputWriter {
	ow := OutputWriter{}
	ow.StartLine()

	for k, v := range data {
		ow.AppendData(k, v)
	}

	return &ow
}

// ToJSON is a function to show the output in json format
func (ow *OutputWriter) ToJSON(v interface{}, pretty bool) {
	value, _ := json.Marshal(v)

	switch pretty {
	case true:
		result, _ := prettyprint(value)
		fmt.Println(string(result))
	default:
		fmt.Println(string(value))
	}
}

// StartLine starts a new line of output
func (ow *OutputWriter) StartLine() {
	ow.finishExistingLine()
	ow.TempValues = make([]string, len(ow.Keys))
}

func (ow *OutputWriter) finishExistingLine() {
	if len(ow.TempValues) > 0 {
		ow.Values = append(ow.Values, ow.TempValues)
		ow.TempValues = nil
	}
}

// AppendDataWithLabel adds a line of data to the output writer
func (ow *OutputWriter) AppendDataWithLabel(key, value, label string) {
	found := -1
	for i, v := range ow.Keys {
		if v == key {
			found = i
		}
	}

	if found == -1 {
		ow.Keys = append(ow.Keys, key)
		ow.Labels = append(ow.Labels, label)
		ow.TempValues = append(ow.TempValues, value)
	} else {
		ow.TempValues[found] = value
	}
}

// AppendData adds a line of data to the output writer
func (ow *OutputWriter) AppendData(key, value string) {
	ow.AppendDataWithLabel(key, value, key)
}

// WriteKeyValues prints a single object stored in the OutputWriter
// in key: value format
func (ow *OutputWriter) WriteKeyValues() {
	ow.finishExistingLine()

	longestLabelLength := 0
	for _, label := range ow.Labels {
		if len(label) > longestLabelLength {
			longestLabelLength = len(label)
		}
	}

	for i := range ow.Keys {
		value := ow.Values[0][i]
		label := ow.Labels[i]
		fmt.Printf("%"+strconv.Itoa(longestLabelLength)+"s : %s\n", label, value)
	}
}

// WriteTable prints multiple objects stored in the OutputWriter
// in tabular format
func (ow *OutputWriter) WriteTable() {
	ow.finishExistingLine()

	table := tablewriter.NewWriter(os.Stdout)
	if len(ow.Keys) > 0 {
		table.SetHeader(ow.Labels)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(false)
		table.SetRowLine(true)
	} else {
		table.SetBorder(false)
	}

	table.AppendBulk(ow.Values)
	table.Render()
}

// WriteSubheader writes a centred heading line in to output
func (ow *OutputWriter) WriteSubheader(label string) {
	count := (72 - len(label) + 2) / 2
	fmt.Println(strings.Repeat("-", count) + " " + label + " " + strings.Repeat("-", count))
}

// WriteHeader WriteSubheader writes a centred heading line in to output
func (ow *OutputWriter) WriteHeader(label string) {
	fmt.Printf("%s:\n", label)
}

func prettyprint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

func (ow *OutputWriter) FinishAndPrintOutput() {
	ow.finishExistingLine()

	ow.WriteTable()
}
