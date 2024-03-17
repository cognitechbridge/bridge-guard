package cmd

import (
	"ctb-cli/core"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Output format enum for the CLI
type outputEnum string

const (
	outputEnumJson outputEnum = "json"
	outputEnumText outputEnum = "text"
	outputEnumYaml outputEnum = "yaml"
	outputEnumXml  outputEnum = "xml"
)

// String is used both by fmt.Print and by Cobra in help text
func (e *outputEnum) String() string {
	return string(*e)
}

// Set is used by Cobra to parse the CLI flags
func (e *outputEnum) Set(v string) error {
	switch v {
	case "json", "text", "yaml", "xml":
		*e = outputEnum(v)
		return nil
	default:
		return errors.New(`must be one of "text", "josn", "yaml", or "xml"`)
	}
}

// Type is only used in help text
func (e *outputEnum) Type() string {
	return "output"
}

// MarshalOutput marshals the given AppResult into a byte slice and prints the result.
// The output format is determined by the value of the `output` variable that is set by the CLI flags.
// The formatted result is then printed to the console.
func MarshalOutput(result core.AppResult) {
	res := make([]byte, 0)
	switch output {
	case outputEnumJson:
		var err error
		res, err = json.Marshal(result)
		if err != nil {
			panic(err)
		}
	case outputEnumYaml:
		var err error
		res, err = yaml.Marshal(result)
		if err != nil {
			panic(err)
		}
	case outputEnumXml:
		var err error
		res, err = xml.Marshal(result)
		if err != nil {
			panic(err)
		}
	case outputEnumText:
		if result.Ok {
			res = fmt.Appendf(res, "Ok\n")
		} else {
			res = fmt.Appendf(res, "Error\n%v", result.Err)
		}
		if result.Result != nil {
			res = fmt.Appendf(res, "%v", result.Result)
		}
	}
	fmt.Println(string(res))
}
