package main

import (
	"encoding/json"
	"errors"
	"strings"
)

/*
	Expected input accepted by this tool. This is effectively the API.
	By defininng our own custom decoders, we are able to apply any "checkin logic" at the time of decoding, as well as
	extending it as needed without poluting the rest of the codebase (decoupling)


		Example of expected input supplied at runtime via "prereqs.json" file:
		{
			"projectname": "nic-test-backbase-reference",
			"environment": "dev",
			"optionals":[
						{
							"name":"cpu",
							"count": 1
						},
						{
							"name":"memory",
							"count":1,
							"unit":"Gi"
						},
						{
							"name":"volumes",
							"count":2
						},
						}
							"name":"storage",
							"count":10,
							"unit":"Gi"
						}
			]
		}


*/

type expectedInput struct {
	ProjectName string           `json:"projectname"`
	Environment string           `json:"environment"`
	Optionals   []optionalObject `json:",omitempty"`
}

type optionalObject struct {
	Name  oName  `json:"name"`
	Count oCount `json:"count"`
	Unit  oUnit  `json:"unit,omitempty"`
}

type optionalObjects []optionalObject

type oName struct {
	string
}

type oCount struct {
	int
}

type oUnit struct {
	string
}

func (o *oName) UnmarshalJSON(data []byte) error {
	var c string
	if err := json.Unmarshal(data, &c); err != nil {
		return err
	}
	// right type, now verify that the value is valid
	lowerCaseName := strings.ToLower(c)
	if !validName(lowerCaseName) {
		return errors.New("optional name entry is invalid: " + lowerCaseName)
	}
	o.string = lowerCaseName
	return nil
}

func (o *oCount) UnmarshalJSON(data []byte) error {
	var c int
	if err := json.Unmarshal(data, &c); err != nil {
		return err
	}
	o.int = c
	return nil
}

func (o *oUnit) UnmarshalJSON(data []byte) error {
	var c string
	if err := json.Unmarshal(data, &c); err != nil {
		return err
	}
	// right type, now verify that the value is valid
	if !validUnit(c) {
		return errors.New("optional unit entry is invalid: " + c)
	}
	o.string = c
	return nil

}

func (input *expectedInput) getOptional(name string) *optionalObject {
	// simple helper that looks for, and then returns an optionalObject with a name that matches name
	for _, object := range input.Optionals {
		if object.Name.string == name {
			return &object
		}
	}
	return nil
}

func validName(name string) bool {
	/*
	  returns true if objects are all contained in:
	  "cpu","memory","volumes"
	*/
	validList := []string{"cpu", "memory", "volumes", "storage"}
	inList := false

	for _, valid := range validList {
		if valid == name {
			inList = true
		}
	}
	return inList

}

func validUnit(unit string) bool {
	/*
	  returns true if objects are all contained in:
	  "Mi", "Gi", "Ti", "Ki", "K", "M", "G", "T", "m"

	  "Mi", "Gi", "Ti", "Ki", "K", "M", "G", "T" are valid units for Memory, and "m" is valid for CPU
	*/
	validList := []string{"Mi", "Gi", "Ti", "Ki", "K", "M", "G", "T", "m"}
	inList := false

	for _, valid := range validList {
		if valid == unit {
			inList = true
		}
	}
	return inList

}

func validUnitDependency(optional optionalObject) bool {
	/*

		certain objects are invalid without both a count and a unit. Since count is compulsory, we use
		the combination of "name" and "unit" to check that a unit has been specified.

	*/
	if optional.Name.string == "memory" || optional.Name.string == "storage" {
		if optional.Unit.string == "" {
			return false
		}
	}
	return true
}

func checkOptionals(opts []optionalObject) error {
	for _, optional := range opts {
		if !validUnitDependency(optional) {
			return errors.New("invalid or missing unit for: " + optional.Name.string)
		}
	}
	return nil
}

func (input *expectedInput) UnmarshalJSON(data []byte) error {
	/*

		type expectedInput struct {
			ProjectName string           `json:"projectname"`
			Environment string           `json:"environment"`
			Optionals   *optionalObjects `json:",omitempty"`
		}

	*/
	type exctract struct {
		ProjectName string           `json:"projectname"`
		Environment string           `json:"environment"`
		Optionals   []optionalObject `json:",omitempty"`
	}

	ex := exctract{}

	err := json.Unmarshal(data, &ex)
	if err != nil {
		return err
	}

	if ex.Environment == "" || ex.ProjectName == "" {
		return errors.New("missing data")
	}
	if strings.Contains(ex.Environment, " ") || strings.Contains(ex.ProjectName, " ") {
		return errors.New("data contains illegal spaces")
	}
	if strings.Contains(ex.Environment, "_") || strings.Contains(ex.ProjectName, "_") {
		return errors.New("data contains illegal underscores")
	}

	// make all lowercase
	input.ProjectName = strings.ToLower(ex.ProjectName)
	input.Environment = strings.ToLower(ex.Environment)
	if ex.Optionals != nil {
		input.Optionals = ex.Optionals
	}
	// check optionals for dependencies
	err = checkOptionals(input.Optionals)
	if err != nil {
		return err
	}
	return nil
}
