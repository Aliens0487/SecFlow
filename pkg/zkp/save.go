package zkp

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/Aliens0487/secFlow/pkg/model"
	"github.com/Aliens0487/secFlow/pkg/model/bpmn"
)

func loadFile(path string) (*model.BPMNGraph, error) {
	xmlBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var definitions bpmn.Definitions
	err = xml.Unmarshal(xmlBytes, &definitions)
	if err != nil {
		return nil, fmt.Errorf("error parsing XML: %w", err)
	}

	return model.BuildGraph(&definitions), nil
}

func loadInputs(path string) (*Inputs, error) {
	jsonBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var inputs Inputs
	err = json.Unmarshal(jsonBytes, &inputs)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return &inputs, nil
}
