package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
	"regexp"
	"strings"
)

type Field struct {
	Name string
	Pg   string
}

func main() {
	if len(os.Args) < 2 {
		logrus.Fatal("missing first arg struct field path")
	}
	structFile := os.Args[1]
	b, err := os.ReadFile(structFile)
	if err != nil {
		logrus.Fatalf("%+v", err)
	}
	r, err := parseStruct(string(b))
	if err != nil {
		logrus.Fatalf("%+v", err)
	}
	resultFile := structFile + ".changedPgFields"
	if err := os.WriteFile(resultFile, []byte(r), 0666); err != nil {
		logrus.Fatalf("%+v", err)
	}
	logrus.Infof("result is written in file %+v", resultFile)
	logrus.Infof("pbcopy < %+v", resultFile)
}

var ignoreFields = []string{"ID", "tableName"}

func parseStruct(s string) (string, error) {
	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")
	if len(lines) < 3 {
		return "", errors.Errorf("not enough lines")
	}
	first := lines[0]
	structName := getStructName(first)
	if len(structName) == 0 {
		return "", errors.Errorf("missing structName")
	}
	lines = lines[1 : len(lines)-1]
	var fields []Field
cycle1:
	for _, l := range lines {
		f, err := parseField(l)
		if err != nil {
			return "", errors.WithStack(err)
		}
		if f == nil {
			return "", errors.Errorf("nil field")
		}
		for _, ig := range ignoreFields {
			if ig == f.Name {
				continue cycle1
			}
		}
		fields = append(fields, *f)
	}
	var resultLines []string
	resultLines = append(resultLines,
		fmt.Sprintf("func Get%+vChangedFields(a *%+v, b %+v) (pgColumns []string) {",
			structName, structName, structName))
	for _, f := range fields {
		resultLines = append(resultLines, fmt.Sprintf("    if a.%+v != b.%+v && !utils.IsEmpty(b.%+v) {", f.Name, f.Name, f.Name))
		resultLines = append(resultLines, fmt.Sprintf(`        a.%+v = b.%+v`, f.Name, f.Name))
		resultLines = append(resultLines, fmt.Sprintf(`        pgColumns = append(pgColumns, "%+v")`, f.Pg))
		resultLines = append(resultLines, fmt.Sprintf(`    }`))
	}
	resultLines = append(resultLines, "    return pgColumns")
	resultLines = append(resultLines, "}")
	return strings.Join(resultLines, "\n"), nil
}

func getStructName(line string) string {
	mm := structNameRegexp.FindAllStringSubmatch(line, -1)
	if len(mm) == 0 {
		return ""
	}
	if len(mm[0]) < 2 {
		return ""
	}
	return mm[0][1]
}

var structNameRegexp = regexp.MustCompile(`type ([^\s]+) struct {`)
var fieldRegexp = regexp.MustCompile(`([^\s]+).+pg:"([^\s]+)"`)

func parseField(line string) (*Field, error) {
	if !fieldRegexp.MatchString(line) {
		return nil, errors.Errorf("couldn't parse line")
	}
	mm := fieldRegexp.FindStringSubmatch(line)
	if len(mm) < 3 {
		return nil, errors.Errorf("not enough submatches")
	}
	fieldName := mm[1]
	pg := strings.TrimSpace(strings.Split(mm[2], ",")[0])
	return &Field{Name: fieldName, Pg: pg}, nil
}

//type IncomeStatement struct {
//	tableName struct{} `json:"-" pg:"income_statement,discard_unknown_columns"`
//	ID        int      `json:"id" pg:"id"`
//	Date      string   `json:"date" pg:"date"`
//}
