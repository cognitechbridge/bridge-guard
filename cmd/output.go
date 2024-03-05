package cmd

import (
	"ctb-cli/core"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
)

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

// Set must have pointer receiver so it doesn't change the value of a copy
func (e *outputEnum) Set(v string) error {
	switch v {
	case "json", "text", "yaml", "xml":
		*e = outputEnum(v)
		return nil
	default:
		return errors.New(`must be one of "text", or "josn"`)
	}
}

// Type is only used in help text
func (e *outputEnum) Type() string {
	return "output"
}

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
