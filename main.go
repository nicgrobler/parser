package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

/*
	This parser takes a basic input (what will eventually come from Helpline request) in json, and uses this, along with some basic
    logic (based on standards) to produce two yaml files:

    1. New project yaml
	2. roleBinding yaml for Active Directory group to this new project.
	3. resource limit yaml

    The AD group's name can be calculated using the logic used to create the group within active directory.
*/

const (
	subDir string = "files"
)

/*
	composable minimal types used to create new yamlfiles
*/

type metaData struct {
	Name      string `yaml:"name"`                // binding name
	NameSpace string `yaml:"namespace,omitempty"` // projectname
}

type roleRef struct {
	Kind     string `yaml:"kind"`     // Project
	APIGroup string `yaml:"apiGroup"` // rbac.authorization.k8s.io
	Name     string `yaml:"name"`     // group name

}

type spec struct {
	Hard struct {
		CPU    int    `yaml:"limits.cpu,omitempty"`
		Memory string `yaml:"limits.memory,omitempty"`
		PVC    int    `yaml:"persistentvolumeclaims,omitempty"`
	} `yaml:"hard,omitempty"`
	Soft struct {
		CPU    int    `yaml:"limits.cpu,omitempty"`
		Memory string `yaml:"limits.memory,omitempty"`
		PVC    int    `yaml:"persistentvolumeclaims,omitempty"`
	} `yaml:"soft,omitempty"`
}

type subject struct {
	Kind     string `yaml:"kind"`     // Project
	APIGroup string `yaml:"apiGroup"` // rbac.authorization.k8s.io
	Name     string `yaml:"name"`     // group name

}

type subjects []subject

type baseObject struct {
	Kind       string   `yaml:"kind"`       // Project
	APIVersion string   `yaml:"apiVersion"` // project.openshift.io/v1
	Metadata   metaData `yaml:"metadata"`
}

type roleBinding struct {
	Kind       string   `yaml:"kind"`       // RoleBinding
	APIVersion string   `yaml:"apiVersion"` // rbac.authorization.k8s.io/v1
	Metadata   metaData `yaml:"metadata"`
	Subjects   subjects `yaml:"subjects"`
	RoleRef    roleRef  `yaml:"roleRef"`
}

type quota struct {
	Kind       string   `yaml:"kind"`       // RoleBinding
	APIVersion string   `yaml:"apiVersion"` // rbac.authorization.k8s.io/v1
	Metadata   metaData `yaml:"metadata"`
	Spec       spec     `yaml:"spec"`
}

/*
	expected input accepted by this tool. This is effectively the API.
*/

type expectedInput struct {
	ProjectName string           `json:"projectname"`
	Role        string           `json:"role"`
	Environment string           `json:"environment"`
	Optionals   *optionalObjects `json:",omitempty"`
}

type optionalObject struct {
	ProjectName string `json:"projectname"`
	Count       int    `json:"count"`
	Unit        string `json:"unit,omitempty"`
}

type optionalObjects []optionalObject

/*
	Main functions for creating our serialized yaml objects
*/

func createNewProjectFile(data *expectedInput) (string, []byte) {
	// create our object
	y := baseObject{
		Kind:       "Project",
		APIVersion: "project.openshift.io/v1",
	}
	y.Metadata.Name = data.ProjectName

	// serialize it into a slice of bytes
	d, err := yaml.Marshal(&y)
	if err != nil {
		exitLog("serialization error: " + err.Error())
	}
	name := strings.ToLower(y.Metadata.Name) + "-new-project.yaml"
	return name, d
}

func createNewRoleBindingFile(data *expectedInput) (string, []byte) {
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
	d, err := yaml.Marshal(&y)
	if err != nil {
		exitLog("serialization error: " + err.Error())
	}
	name := strings.ToLower(y.Metadata.NameSpace) + "-new-rolebinding.yaml"
	return name, d
}

func createNewLimitsFile(data *expectedInput) (string, []byte) {
	// create our object
	y := quota{
		Kind:       "ResourceQuota",
		APIVersion: "v1",
	}
	y.Metadata.Name = "default-quotas"
	y.Metadata.NameSpace = data.ProjectName
	y.Spec.Hard.CPU = 1
	y.Spec.Hard.Memory = "2Gi"
	y.Spec.Hard.PVC = 3

	// serialize it into a slice of bytes
	d, err := yaml.Marshal(&y)
	if err != nil {
		exitLog("serialization error: " + err.Error())
	}
	name := strings.ToLower(y.Metadata.NameSpace) + "-new-quota.yaml"
	return name, d
}

func process(data *expectedInput) {
	/*
		input will ALWAYS include projectname, role, and environment members - but MAY include optionals as well.
		so, we need to take this into account here
	*/

	dumpToFile(createNewProjectFile(data))
	dumpToFile(createNewRoleBindingFile(data))
	if data.Optionals != nil {
		dumpToFile(createNewLimitsFile(data))
	}

}

/*
	Helpers
*/

func dumpToFile(name string, data []byte) {
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

func logFunction(format string) {
	fmt.Println(format)
	os.Exit(1)
}

var exitLog = logFunction

func main() {

	data, err := ioutil.ReadFile("prereqs.json")
	if err != nil {
		exitLog("program exited due to error reading input file: " + err.Error())
	}

	var inputData expectedInput
	err = json.Unmarshal(data, &inputData)
	if err != nil {
		exitLog("program exited due to error in input: " + err.Error())
	}

	process(&inputData)
}
