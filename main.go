package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

/*
	This parser takes a basic input (what will eventually come from Helpline request / elsewhere) in json, and uses this, along with some basic
    logic (based on standards) to produce two json files:
    1. New project json
	2. roleBinding json for EDIT Active Directory group to this new project*
	3. roleBinding json for VIEW Active Directory group to this new project*
	4. roleBinding json for relman service account to this new project
	5. resource limit json
	6. networkPolicy for the project
	7. egress networkPolicy for the project

    *The AD group names are generated using the logic used to create the groups within active directory.
*/

const (
	quotaFilename               string = "10-quotas.yaml"
	projectFilename             string = "1-project.yaml"
	defaultRolebindingFilename  string = "10-default-rolebinding.yaml"
	jenkinsRolebindinngFilename string = "10-jenkins-rolebinding.yaml"
	editRolebindingFilename     string = "10-edit-group-rolebinding.yaml"
	viewRolebindingFilename     string = "10-view-group-rolebinding.yaml"
	networkPolicyFilename       string = "10-networkpolicy.yaml"
	egressNetworkPolicyFilename string = "10-egress-networkpolicy.yaml"
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
		CPU     interface{} `json:"limits.cpu,omitempty"`
		Memory  string      `json:"limits.memory,omitempty"`
		PVC     int         `json:"persistentvolumeclaims,omitempty"`
		Storage string      `json:"requests.storage,omitempty"`
	} `json:"hard,omitempty"`
}

type specNetwork struct {
	PodSelector struct {
		Todo string `json:"not-implemented-yet,omitempty"`
	} `json:"podSelector,omitempty"`
	PolicyTypes []string `json:"policyTypes,omitempty"`
}

type specEgressNetwork struct {
	Egress []egressRules `json:"egress"`
}

type egressRules struct {
	EgressType string `json:"type"`
	To         struct {
		Cidr string `json:"cidrSelector,omitempty"`
		URL  string `json:"dnsName,omitempty"`
	} `json:"to"`
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

type egressNetwork struct {
	Kind       string            `json:"kind"`       // EgressNetworkPolicy
	APIVersion string            `json:"apiVersion"` // network.openshift.io/v1
	Metadata   metaData          `json:"metadata"`
	Spec       specEgressNetwork `json:"spec"`
}

/*
	Results object: a simple nested json, where each key is the name of the destinationn file,
	and the data, is the associated value
*/

type resultEntry struct {
	Name    string      `json:"filename"`
	Content interface{} `json:"content"`
}

type resultsObject []resultEntry

/*
	Main functions for creating our serialized json objects
*/

func createProjectObject(data *expectedInput) (string, baseObject) {
	// create our object
	y := baseObject{
		Kind:       "Project",
		APIVersion: "project.openshift.io/v1",
	}
	y.Metadata.Name = data.ProjectName

	name := projectFilename
	return name, y
}

func createNetworkPolicyObject(data *expectedInput) (string, network) {

	// create our NetworkPolicy object
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

	name := networkPolicyFilename

	return name, y

}

func createEgressNetworkPolicyObject(data *expectedInput) (string, egressNetwork) {

	// create our EgressNetworkPolicy object
	e := egressNetwork{
		Kind:       "EgressNetworkPolicy",
		APIVersion: "network.openshift.io/v1",
	}
	e.Metadata.Name = "default-egress"
	e.Metadata.NameSpace = data.ProjectName
	e.Spec.Egress = []egressRules{egressRules{EgressType: "Deny"}}
	e.Spec.Egress[0].To.Cidr = "0.0.0.0/0"

	name := egressNetworkPolicyFilename

	return name, e

}

func createRoleBindingObjects(data *expectedInput) ([]string, []roleBinding) {
	/*
		This function will produce the data for 3 files:

		1. The generated AD groupname that has the EDIT role
		2. The generated AD groupname that has the VIEW role
		3. The static service account name (relman) that has admin role for deployments
	*/
	var names []string
	var bytes []roleBinding

	// first generate data for 1 & 2 above
	adRolesAndGroupNames := generateADGroupNames(data)
	for roleName, adGroupName := range adRolesAndGroupNames {
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
		y.RoleRef.Name = strings.ToLower(roleName)

		name := ""
		if y.RoleRef.Name == "edit" {
			name = editRolebindingFilename
		} else {
			name = viewRolebindingFilename
		}
		// add to results
		names = append(names, name)
		bytes = append(bytes, y)
	}
	// now do 3
	roleName := "admin-relman"
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
			Kind:      "ServiceAccount",
			Name:      "relman",
			Namespace: "relman",
		},
	}
	y.RoleRef.APIGroup = "rbac.authorization.k8s.io"
	y.RoleRef.Kind = "ClusterRole"
	y.RoleRef.Name = "admin"

	name := jenkinsRolebindinngFilename

	// add to results
	names = append(names, name)
	bytes = append(bytes, y)

	return names, bytes
}

func createLimitsObject(data *expectedInput) (string, quota) {
	if data.Optionals == nil {
		// should never happen, but if so, handle it
		return "", quota{}
	}
	// create our object
	y := quota{
		Kind:       "ResourceQuota",
		APIVersion: "v1",
	}
	y.Metadata.Name = "default-quotas"
	y.Metadata.NameSpace = data.ProjectName

	// now get the optionals
	if o := data.getOptional("cpu"); o != nil {
		// CPU can be specified with, and without a suffix - handle both
		if o.Unit.string != "" {
			y.Spec.Hard.CPU = concat(o.Count.int, o.Unit.string)
		} else {
			y.Spec.Hard.CPU = o.Count.int
		}
	}

	if o := data.getOptional("memory"); o != nil {
		y.Spec.Hard.Memory = concat(o.Count.int, o.Unit.string)
	}

	if o := data.getOptional("volumes"); o != nil {
		y.Spec.Hard.PVC = o.Count.int
	}

	if o := data.getOptional("storage"); o != nil {
		y.Spec.Hard.Storage = concat(o.Count.int, o.Unit.string)
	}

	name := quotaFilename

	return name, y
}

func concat(i int, s string) string {
	return strconv.Itoa(i) + s
}

func process(data *expectedInput) *resultsObject {
	/*
		Populate our resultsObject here with each file and its contents
	*/

	results := &resultsObject{}
	results.addPayload(createProjectObject(data))
	results.addPayload(createRoleBindingObjects(data))
	results.addPayload(createLimitsObject(data))
	results.addPayload(createNetworkPolicyObject(data))
	results.addPayload(createEgressNetworkPolicyObject(data))

	return results
}

/*
	Helpers
*/

func (results *resultsObject) addBindingsType(name []string, d interface{}) bool {
	switch data := d.(type) {
	case []roleBinding:
		// now work on each set of data in turn
		for i := range name {
			r := resultEntry{}
			r.Name = name[i]
			r.Content = data[i]
			*results = append(*results, r)
		}
		return true
	default:
		return false
	}
}

func isEmptyObject(d interface{}) bool {
	// quotas may be empty if no limits were supplied - if so, we want to avoid adding it
	switch object := d.(type) {
	case quota:
		if (quota{}) == object {
			return true
		}
	}
	return false
}

func (results *resultsObject) addPayload(f interface{}, d interface{}) {
	/*
		f can be a string, or []string - d can be an object, or a slice of them
	*/
	if isEmptyObject(d) {
		return
	}

	switch name := f.(type) {
	case string:
		r := resultEntry{}
		r.Name = name
		r.Content = d
		*results = append(*results, r)

	case []string:
		// assert that data is of the expected type, and add correct type
		ok := results.addBindingsType(name, d)
		if !ok {
			exitLog("invalid datatype passed, expected []roleBinding")
		}

	default:
		exitLog("invalid datatype passed")
	}
}

func generateADGroupNames(data *expectedInput) map[string]string {
	/*
		AD groups names will be gererated as:

		"RES" + "-" + environment + "-" + "OPSH" + "-" + role + "-" + project_name

		returns a map of "OPENSHIFT ROLE" : "AD GROUP NAME"
	*/
	s := make(map[string]string)
	s["EDIT"] = strings.ToUpper("RES" + "-" + data.Environment + "-" + "OPSH" + "-" + "DEVELOPER" + "-" + strings.ReplaceAll(data.ProjectName, "-", "_"))
	s["VIEW"] = strings.ToUpper("RES" + "-" + data.Environment + "-" + "OPSH" + "-" + "VIEWER" + "-" + strings.ReplaceAll(data.ProjectName, "-", "_"))
	return s
}

func logFunction(format string) {
	fmt.Println(format)
	os.Exit(1)
}

var exitLog = logFunction

func main() {

	var incomingJSON *string
	incomingJSON = flag.String("data", "", "the json payload used to generate the OpenShift json")
	flag.Parse()

	if *incomingJSON == "" {
		exitLog("program exited due to missing input")
	}

	var inputData expectedInput
	// unmarshal will call our custom decoders which do input verification
	err := json.Unmarshal([]byte(*incomingJSON), &inputData)
	if err != nil {
		exitLog("program exited due to error in parsing input: " + err.Error())
	}

	// lets go
	rawResults := process(&inputData)

	// serialize data to JSON
	data, err := json.MarshalIndent(rawResults, "", "  ")
	if err != nil {
		exitLog("serialization error: " + err.Error())
	}

	// dump result to STDOUT
	fmt.Println(string(data))

}
