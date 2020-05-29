package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

/*
	This parser takes a basic input (what will eventually come from Helpline request / elsewhere) in json, and uses this, along with some basic
    logic (based on standards) to produce two json files:
    1. New project json
	2. roleBinding json for Active Directory group to this new project.
	3. resource limit json
    The AD group's name can be calculated using the logic used to create the group within active directory.
*/

const (
	subDir     string = "files"
	priority   string = "1"
	nopriority string = "10"
)

/*
	Composable minimal types used to create new json files.
*/

type metaData struct {
	Name      string `json:"name"`                // binding name
	NameSpace string `json:"namespace,omitempty"` // projectname
}

type roleRef struct {
	Kind     string `json:"kind"`     // Project
	APIGroup string `json:"apiGroup"` // rbac.authorization.k8s.io
	Name     string `json:"name"`     // group name

}

type specQuota struct {
	Hard struct {
		CPU    int    `json:"limits.cpu,omitempty"`
		Memory string `json:"limits.memory,omitempty"`
		PVC    int    `json:"persistentvolumeclaims,omitempty"`
	} `json:"hard,omitempty"`
}

type specNetwork struct {
	PodSelector struct {
		Todo string `json:"not-implemented-yet,omitempty"`
	} `json:"podSelector,omitempty"`
	PolicyTypes []string `json:"policyTypes,omitempty"`
}

type subject struct {
	Kind      string `json:"kind"`                // Project
	APIGroup  string `json:"apiGroup,omitempty"`  // rbac.authorization.k8s.io
	Name      string `json:"name"`                // group name
	Namespace string `json:"namespace,omitempty"` // name of project

}

type subjects []subject

type baseObject struct {
	Kind       string   `json:"kind"`       // Project
	APIVersion string   `json:"apiVersion"` // project.openshift.io/v1
	Metadata   metaData `json:"metadata"`
}

type roleBinding struct {
	Kind       string   `json:"kind"`       // RoleBinding
	APIVersion string   `json:"apiVersion"` // rbac.authorization.k8s.io/v1
	Metadata   metaData `json:"metadata"`
	Subjects   subjects `json:"subjects"`
	RoleRef    roleRef  `json:"roleRef"`
}

type quota struct {
	Kind       string    `json:"kind"`       // RoleBinding
	APIVersion string    `json:"apiVersion"` // rbac.authorization.k8s.io/v1
	Metadata   metaData  `json:"metadata"`
	Spec       specQuota `json:"spec"`
}

type network struct {
	Kind       string      `json:"kind"`       // NetworkPolicy
	APIVersion string      `json:"apiVersion"` // networking.k8s.io/v1
	Metadata   metaData    `json:"metadata"`
	Spec       specNetwork `json:"spec"`
}

/*
	Expected input accepted by this tool. This is effectively the API.
*/

type expectedInput struct {
	ProjectName string           `json:"projectname"`
	Role        string           `json:"role"`
	Environment string           `json:"environment"`
	Optionals   *optionalObjects `json:",omitempty"`
}

type optionalObject struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Unit  string `json:"unit,omitempty"`
}

type optionalObjects []optionalObject

/*
	Main functions for creating our serialized json objects
*/

func createNewProjectFile(data *expectedInput) (string, []byte) {
	// create our object
	y := baseObject{
		Kind:       "Project",
		APIVersion: "project.openshift.io/v1",
	}
	y.Metadata.Name = data.ProjectName

	// serialize it into a slice of bytes
	d, err := json.MarshalIndent(&y, "", "  ")
	if err != nil {
		exitLog("serialization error: " + err.Error())
	}
	name := priority + "-" + strings.ToLower(y.Metadata.Name) + "-new-project.json"
	return name, d
}

func createNewNetworkPolicyFile(data *expectedInput) (string, []byte) {
	/*

		kind: NetworkPolicy
		apiVersion: networking.k8s.io/v1
		metadata:
		  name: deny-by-default
		  namespace: aaaa
		spec:
		  podSelector: {}
		  policyTypes:
		    - Ingress
			- Egress

	*/
	// create our object
	y := network{
		Kind:       "NetworkPolicy",
		APIVersion: "networking.k8s.io/v1",
	}
	y.Metadata.Name = "deny-by-default"
	y.Metadata.NameSpace = data.ProjectName
	y.Spec.PolicyTypes = []string{
		"Ingress",
		"Egress",
	}

	// serialize it into a slice of bytes
	d, err := json.MarshalIndent(&y, "", "  ")
	if err != nil {
		exitLog("serialization error: " + err.Error())
	}
	name := nopriority + "-" + strings.ToLower(y.Metadata.NameSpace) + "-new-networkpolicy.json"

	return name, d

}

func createNewRoleBindingFiles(data *expectedInput) ([]string, [][]byte) {
	var names []string
	var bytes [][]byte

	// first process the
	// compose our strings
	adGroupName := inferADGroupName(data)
	roleName := lookupRole(data.Role)
	roleBindingName := strings.ToLower(data.ProjectName + "-" + roleName + "-" + "binding")
	// create our object
	y := roleBinding{
		Kind:       "RoleBinding",
		APIVersion: "rbac.authorization.k8s.io/v1",
	}
	y.Metadata.Name = roleBindingName
	y.Metadata.NameSpace = data.ProjectName
	y.Subjects = subjects{
		subject{
			Kind:     "Group",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     adGroupName,
		},
	}
	y.RoleRef.APIGroup = "rbac.authorization.k8s.io"
	y.RoleRef.Kind = "ClusterRole"
	y.RoleRef.Name = roleName

	// serialize it into a slice of bytes
	d, err := json.MarshalIndent(&y, "", "  ")
	if err != nil {
		exitLog("serialization error: " + err.Error())
	}
	name := nopriority + "-" + strings.ToLower(y.Metadata.NameSpace) + "-new-rolebinding.json"

	// add to results
	names = append(names, name)
	bytes = append(bytes, d)

	// now do the second roleBinding
	roleName = "admin-relman"
	roleBindingName = strings.ToLower(data.ProjectName + "-" + roleName + "-" + "binding")
	// create our object
	y = roleBinding{
		Kind:       "RoleBinding",
		APIVersion: "rbac.authorization.k8s.io/v1",
	}
	y.Metadata.Name = roleBindingName
	y.Metadata.NameSpace = data.ProjectName
	y.Subjects = subjects{
		subject{
			Kind:      "ServiceAccount",
			Name:      "relman",
			Namespace: "relman",
		},
	}
	y.RoleRef.APIGroup = "rbac.authorization.k8s.io"
	y.RoleRef.Kind = "ClusterRole"
	y.RoleRef.Name = "admin"

	// serialize it into a slice of bytes
	d, err = json.MarshalIndent(&y, "", "  ")
	if err != nil {
		exitLog("serialization error: " + err.Error())
	}
	name = nopriority + "-" + strings.ToLower(y.Metadata.NameSpace) + "-new-default-rolebinding.json"

	// add to results
	names = append(names, name)
	bytes = append(bytes, d)

	return names, bytes
}

func createNewLimitsFile(data *expectedInput) (string, []byte) {
	if data.Optionals == nil {
		// should never happen, but if so, handle it
		return "", nil
	}
	// create our object
	y := quota{
		Kind:       "ResourceQuota",
		APIVersion: "v1",
	}
	y.Metadata.Name = "default-quotas"
	y.Metadata.NameSpace = data.ProjectName

	// now get the optionals
	if o := data.Optionals.get("cpu"); o != nil {
		y.Spec.Hard.CPU = o.Count
	}

	if o := data.Optionals.get("memory"); o != nil {
		y.Spec.Hard.Memory = concat(o.Count, o.Unit)
	}

	if o := data.Optionals.get("volumes"); o != nil {
		y.Spec.Hard.PVC = o.Count
	}

	// serialize it into a slice of bytes
	d, err := json.MarshalIndent(&y, "", "  ")
	if err != nil {
		exitLog("serialization error: " + err.Error())
	}
	name := nopriority + "-" + strings.ToLower(y.Metadata.NameSpace) + "-new-quota.json"
	return name, d
}

func concat(i int, s string) string {
	return strconv.Itoa(i) + s
}

func process(data *expectedInput) {
	/*
		input will ALWAYS include projectname, role, and environment members - but MAY include optionals as well.
		so, we need to take this into account here
	*/

	dumpToFile(createNewProjectFile(data))
	dumpToFile(createNewRoleBindingFiles(data))
	if data.Optionals != nil {
		dumpToFile(createNewLimitsFile(data))
	}
	dumpToFile(createNewNetworkPolicyFile(data))

	// create the touchfile used for CICD
	createEnvTouchFile(data)

}

/*
	Helpers
*/

func (objects optionalObjects) get(name string) *optionalObject {
	// simple helper that looks for, and then returns an optionalObject with a name that matches name
	for _, object := range objects {
		if object.Name == name {
			return &object
		}
	}
	return nil
}

func dumpToFile(f interface{}, d interface{}) {
	/*
		helper that will write out supplied data given the followinng inputs:
		1. string, []byte -> filename and data
		2. []string, [][]byte -> list of file names, and data for each

		anything else is invalid
	*/
	switch name := f.(type) {
	case string:
		// assert that data is of the expected type
		data, ok := d.([]byte)
		if !ok {
			exitLog("invalid type for data: expected []byte")
		}
		file, err := os.Create(subDir + "/" + name)
		defer file.Close()
		// file create
		if err != nil {
			exitLog("failed to create output file due to error: " + err.Error())
		}
		// write data
		_, err = file.Write(data)
		if err != nil {
			exitLog("failed to write to output file due to error: " + err.Error())
		}
	case []string:
		// assert that data is of the expected type
		data, ok := d.([][]byte)
		if !ok {
			exitLog("invalid type for data: expected []byte")
		}
		// now work on each file in turn
		for i, n := range name {
			file, err := os.Create(subDir + "/" + n)
			defer file.Close()
			// file create
			if err != nil {
				exitLog("failed to create output file due to error: " + err.Error())
			}
			// write data
			_, err = file.Write(data[i])
			if err != nil {
				exitLog("failed to write to output file due to error: " + err.Error())
			}
		}
	default:
		exitLog("invalid datatypes passed")
	}
}

/*
	AD groups names will be formatted as:
	"RES" + "-" + environment + "-" + "OPSH" + "-" + role + "-" + project_name
	values within the input data are used in to infer what the AD group name will be within the RoleBinding
*/

func inferADGroupName(data *expectedInput) string {
	s := "RES" + "-" + data.Environment + "-" + "OPSH" + "-" + data.Role + "-" + strings.ReplaceAll(data.ProjectName, "-", "_")
	return strings.ToUpper(s)
}

func createTouchfileName(data *expectedInput) string {
	return "OPSH_ENV." + strings.ToUpper(data.Environment)
}
func createEnvTouchFile(data *expectedInput) {
	fileName := createTouchfileName(data)
	_, err := os.Create(fileName)
	if err != nil {
		exitLog("failed to create touchfile due to error: " + err.Error())
	}
}

func lookupRole(ut string) string {
	userType := strings.ToLower(ut)
	switch userType {
	case "developer":
		return "edit"
	case "admin":
		return "admin"
	case "readonly":
		return "view"
	default:
		exitLog("invalid user type specified")
	}
	return ""
}

func checkInputValid(data expectedInput) error {
	/*

		type expectedInput struct {
			ProjectName string           `json:"projectname"`
			Role        string           `json:"role"`
			Environment string           `json:"environment"`
			Optionals   *optionalObjects `json:",omitempty"`
		}

	*/
	if data.Environment == "" || data.Role == "" || data.ProjectName == "" {
		return errors.New("missing data")
	}
	if strings.Contains(data.Environment, " ") || strings.Contains(data.Role, " ") || strings.Contains(data.ProjectName, " ") {
		return errors.New("data contains illegal spaces")
	}
	if strings.Contains(data.Environment, "_") || strings.Contains(data.Role, "_") || strings.Contains(data.ProjectName, "_") {
		return errors.New("data contains illegal underscores")
	}

	// make all lowercase
	data.ProjectName = strings.ToLower(data.ProjectName)
	data.Environment = strings.ToLower(data.Environment)
	data.Role = strings.ToLower(data.Role)

	return nil
}

func formatInput(data *expectedInput) {

	// make all lowercase
	data.ProjectName = strings.ToLower(data.ProjectName)
	data.Environment = strings.ToLower(data.Environment)
	data.Role = strings.ToLower(data.Role)

}

func logFunction(format string) {
	fmt.Println(format)
	os.Exit(1)
}

var exitLog = logFunction

func main() {

	/*
		Example of expected input supplied at runtime via "prereqs.json" file:
		{
			"projectname": "nic-test-backbase-reference",
			"role": "developer",
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
						}
			]
		}
	*/
	data, err := ioutil.ReadFile("prereqs.json")
	if err != nil {
		exitLog("program exited due to error reading input file: " + err.Error())
	}

	var inputData expectedInput
	err = json.Unmarshal(data, &inputData)
	if err != nil {
		exitLog("program exited due to error in parsing input: " + err.Error())
	}

	// run input through basic checks
	if err := checkInputValid(inputData); err != nil {
		exitLog("program exited due to error in input value(s): " + err.Error())
	}
	formatInput(&inputData)
	// lets go
	process(&inputData)
}
