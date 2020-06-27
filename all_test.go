package main

import (
	"encoding/json"
	"testing"
)

func findObjectIndex(name string, files []string) (int, bool) {
	// returns the index of name
	for i, f := range files {
		if f == name {
			return i, true
		}
	}
	return 0, false
}

func TestGenerateADGroupName(t *testing.T) {
	i := expectedInput{Environment: "boogie", ProjectName: "extra-good"}
	got := generateADGroupNames(&i)
	want := "RES-BOOGIE-OPSH-DEVELOPER-EXTRA_GOOD"
	if want != got["EDIT"] {
		t.Errorf("wanted %s, but got %s: \n", want, got)
	}
	want = "RES-BOOGIE-OPSH-VIEWER-EXTRA_GOOD"
	if want != got["VIEW"] {
		t.Errorf("wanted %s, but got %s: \n", want, got)
	}
}

func TestCheckInputValid(t *testing.T) {
	data := []byte(`{
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
	}`)
	d := expectedInput{}
	err := json.Unmarshal(data, &d)
	if err != nil {
		t.Errorf("wanted %s, but got %s: \n", "nil", err.Error())
	}

	if d.Environment != "dev" {
		t.Errorf("wanted %s, but got %s: \n", "dev", d.Environment)
	}

	if d.Optionals == nil {
		t.Errorf("wanted %s, but got %s: \n", "optionals", "nil")
	}

	// complain about spaces
	badData := []byte(`{
		"projectname": "nic-test backbase-reference",
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
	}`)
	d = expectedInput{}
	err = json.Unmarshal(badData, &d)
	if err == nil {
		t.Errorf("wanted %s, but got %s: \n", "an error", "nil")
	}
	want := "data contains illegal spaces"
	got := err.Error()
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}

	// complain about underscores
	badData = []byte(`{
		"projectname": "nic_test-backbase-reference",
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
	}`)
	d = expectedInput{}
	err = json.Unmarshal(badData, &d)
	if err == nil {
		t.Errorf("wanted %s, but got %s: \n", "an error", "nil")
	}
	want = "data contains illegal underscores"
	got = err.Error()
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}

	// should autoformat the data
	badData = []byte(`{
		"projectname": "NIC-test-backbase-reference",
		"environment": "DEV",
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
	}`)
	d = expectedInput{}
	err = json.Unmarshal(badData, &d)
	if err != nil {
		t.Errorf("wanted %s, but got %s: \n", "nil", err.Error())
	}

	if d.ProjectName != "nic-test-backbase-reference" {
		t.Errorf("wanted %v, but got %v: \n", "nic-test-backbase-reference", d.ProjectName)
	}

	if d.Environment != "dev" {
		t.Errorf("wanted %v, but got %v: \n", "dev", d.Environment)
	}

	// should complain about invalid name in optionals
	badData = []byte(`{
		"projectname": "NIC-test-backbase-reference",
		"environment": "DEV",
		"optionals":[
					{
						"name":"cpu",
						"count": 1
					},
					{
						"name":"memooory",
						"count":1,
						"unit":"Gi"
					},
					{
						"name":"volumes",
						"count":2
					}
		]
	}`)
	d = expectedInput{}
	err = json.Unmarshal(badData, &d)
	if err == nil {
		t.Errorf("wanted %s, but got %s: \n", "an error", "nil")
	}
	if err.Error() != "optional name entry is invalid: memooory" {
		t.Errorf("wanted %s, but got %s: \n", "an error", err.Error())
	}

	// should complain about invalid unit in optionals
	badData = []byte(`{
		"projectname": "NIC-test-backbase-reference",
		"environment": "DEV",
		"optionals":[
					{
						"name":"cpu",
						"count": 1
					},
					{
						"name":"memory",
						"count":1,
						"unit":"Giz"
					},
					{
						"name":"volumes",
						"count":2
					}
		]
	}`)
	d = expectedInput{}
	err = json.Unmarshal(badData, &d)
	if err == nil {
		t.Errorf("wanted %s, but got %s: \n", "an error", "nil")
	}
	if err.Error() != "optional unit entry is invalid: Giz" {
		t.Errorf("wanted %s, but got %s: \n", "optional unit entry is invalid: Giz", err.Error())
	}

	// should complain about missing unit in optionals
	badData = []byte(`{
		"projectname": "NIC-test-backbase-reference",
		"environment": "DEV",
		"optionals":[
					{
						"name":"cpu",
						"count": 1
					},
					{
						"name":"storage",
						"count":1
					},
					{
						"name":"volumes",
						"count":2
					}
		]
	}`)
	d = expectedInput{}
	err = json.Unmarshal(badData, &d)
	if err == nil {
		t.Errorf("wanted %s, but got %s: \n", "an error", "nil")
	}
	if err.Error() != "invalid or missing unit for: storage" {
		t.Errorf("wanted %s, but got %s: \n", "invalid or missing unit for: storage", err.Error())
	}

	// should complain about invalid count in optionals with type error
	badData = []byte(`{
		"projectname": "NIC-test-backbase-reference",
		"environment": "DEV",
		"optionals":[
					{
						"name":"cpu",
						"count": 1
					},
					{
						"name":"memory",
						"count":1.1,
						"unit":"Gi"
					},
					{
						"name":"volumes",
						"count":2
					}
		]
	}`)
	d = expectedInput{}
	err = json.Unmarshal(badData, &d)
	if err == nil {
		t.Errorf("wanted %s, but got %s: \n", "an error", "nil")
	}
	if err.Error() != "json: cannot unmarshal number 1.1 into Go struct field optionalObject.Optionals.count of type int" {
		t.Errorf("wanted %s, but got %s: \n", "json: cannot unmarshal number 1.1 into Go struct field optionalObject.Optionals.count of type int", err.Error())
	}

	// should complain about invalid count in optionals with type error
	badData = []byte(`{
		"projectname": "NIC-test-backbase-reference",
		"environment": "DEV",
		"optionals":[
					{
						"name":"cpu",
						"count": 1
					},
					{
						"name":"memory",
						"count":"1",
						"unit":"Gi"
					},
					{
						"name":"volumes",
						"count":2
					}
		]
	}`)
	d = expectedInput{}
	err = json.Unmarshal(badData, &d)
	if err == nil {
		t.Errorf("wanted %s, but got %s: \n", "an error", "nil")
	}
	if err.Error() != "json: cannot unmarshal string into Go struct field optionalObject.Optionals.count of type int" {
		t.Errorf("wanted %s, but got %s: \n", "json: cannot unmarshal string into Go struct field optionalObject.Optionals.count of type int", err.Error())
	}

	// should complain about invalid count in optionals with type error
	data = []byte(`{
		"projectname": "NIC-test-backbase-reference",
		"environment": "DEV",
		"optionals":[
					{
						"name":"cpu",
						"count": 1000,
						"unit": "m"
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
	}`)
	d = expectedInput{}
	err = json.Unmarshal(data, &d)
	if err != nil {
		t.Errorf("wanted %s, but got %s: \n", "nil", err.Error())
	}
	if d.Optionals[0].Count.int != 1000 || d.Optionals[0].Unit.string != "m" {
		t.Errorf("wanted %s, but got %s: \n", "should be equal", "are not equal")
	}

}

func TestValidUnit(t *testing.T) {

	want := false
	got := validUnit("gb")
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}

	want = true
	got = validUnit("Mi")
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}
}

func TestValidName(t *testing.T) {
	want := false
	got := validName("cpus")
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}

	want = true
	got = validName("cpu")
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}

	want = false
	got = validName("Memory")
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}

	want = true
	got = validName("volumes")
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}

	want = true
	got = validName("storage")
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}

	want = false
	got = validName("disk")
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}
}

func TestCreateNewProjectObject(t *testing.T) {

	expectedBytes := []byte(`{"kind":"Project","apiVersion":"project.openshift.io/v1","metadata":{"name":"boogie-test"}}`)

	i := expectedInput{ProjectName: "boogie-test"}

	fileName, baseObject := createProjectObject(&i)
	gotBytes, err := json.Marshal(baseObject)
	if err != nil {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
	}
	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}
	expectedObjectName := projectFilename
	if expectedObjectName != fileName {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedObjectName, fileName)
	}

}

func TestCreateNewNetworkPolicyObject(t *testing.T) {

	expectedBytes := []byte(`{"kind":"NetworkPolicy","apiVersion":"networking.k8s.io/v1","metadata":{"name":"deny-by-default","namespace":"boogie-test"},"spec":{"podSelector":{},"policyTypes":["Ingress","Egress"]}}`)

	i := expectedInput{ProjectName: "boogie-test"}
	fileName, baseObject := createNetworkPolicyObject(&i)
	gotBytes, err := json.Marshal(baseObject)
	if err != nil {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
	}
	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}
	expectedObjectName := networkPolicyFilename
	if expectedObjectName != fileName {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedObjectName, fileName)
	}

}

func TestCreateNewEgressNetworkPolicyObject(t *testing.T) {

	expectedBytes := []byte(`{"kind":"EgressNetworkPolicy","apiVersion":"network.openshift.io/v1","metadata":{"name":"default-egress","namespace":"boogie-test"},"spec":{"egress":[{"type":"Deny","to":{"cidrSelector":"0.0.0.0/0"}}]}}`)

	i := expectedInput{ProjectName: "boogie-test"}

	fileName, baseObject := createEgressNetworkPolicyObject(&i)
	gotBytes, err := json.Marshal(baseObject)
	if err != nil {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
	}
	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}
	expectedObjectName := egressNetworkPolicyFilename
	if expectedObjectName != fileName {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedObjectName, fileName)
	}

}

func TestCreateNewRoleBindingObject(t *testing.T) {

	expectedBytes := make(map[string][]byte)

	expectedBytes[editRolebindingFilename] = []byte(`{"kind":"RoleBinding","apiVersion":"rbac.authorization.k8s.io/v1","metadata":{"name":"boogie-test-edit-binding","namespace":"boogie-test"},"subjects":[{"kind":"Group","apiGroup":"rbac.authorization.k8s.io","name":"RES-DEV-OPSH-DEVELOPER-BOOGIE_TEST"}],"roleRef":{"kind":"ClusterRole","apiGroup":"rbac.authorization.k8s.io","name":"edit"}}`)
	expectedBytes[viewRolebindingFilename] = []byte(`{"kind":"RoleBinding","apiVersion":"rbac.authorization.k8s.io/v1","metadata":{"name":"boogie-test-view-binding","namespace":"boogie-test"},"subjects":[{"kind":"Group","apiGroup":"rbac.authorization.k8s.io","name":"RES-DEV-OPSH-VIEWER-BOOGIE_TEST"}],"roleRef":{"kind":"ClusterRole","apiGroup":"rbac.authorization.k8s.io","name":"view"}}`)
	expectedBytes[jenkinsRolebindinngFilename] = []byte(`{"kind":"RoleBinding","apiVersion":"rbac.authorization.k8s.io/v1","metadata":{"name":"boogie-test-admin-relman-binding","namespace":"boogie-test"},"subjects":[{"kind":"ServiceAccount","name":"relman","namespace":"relman"}],"roleRef":{"kind":"ClusterRole","apiGroup":"rbac.authorization.k8s.io","name":"admin"}}`)

	i := expectedInput{ProjectName: "boogie-test", Environment: "dev"}

	fileNames, baseObject := createRoleBindingObjects(&i)
	expectedObjectName := editRolebindingFilename

	index, found := findObjectIndex(expectedObjectName, fileNames)
	// verify file is in list
	if !found {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "true", "false")
	}
	// verify contents are correct
	gotBytes, err := json.Marshal(baseObject[index])
	if err != nil {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
	}
	if string(expectedBytes[expectedObjectName]) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes[expectedObjectName], gotBytes)
	}

	expectedObjectName = viewRolebindingFilename
	index, found = findObjectIndex(expectedObjectName, fileNames)
	// verify file is in list
	if !found {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "true", "false")
	}

	gotBytes, err = json.Marshal(baseObject[index])
	if err != nil {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
	}
	if string(expectedBytes[expectedObjectName]) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes[expectedObjectName], gotBytes)
	}

	expectedObjectName = jenkinsRolebindinngFilename
	index, found = findObjectIndex(expectedObjectName, fileNames)
	// verify file is in list
	if !found {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "true", "false")
	}
	gotBytes, err = json.Marshal(baseObject[index])
	if err != nil {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
	}
	if string(expectedBytes[expectedObjectName]) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes[expectedObjectName], gotBytes)
	}

}

func TestCreateNewLimitsObject(t *testing.T) {
	expectedBytes := []byte(`{"kind":"ResourceQuota","apiVersion":"v1","metadata":{"name":"default-quotas","namespace":"boogie-test"},"spec":{"hard":{"limits.cpu":2,"limits.memory":"1Gi","persistentvolumeclaims":3,"requests.storage":"100Gi"}}}`)

	o := []optionalObject{
		optionalObject{
			Name:  oName{"cpu"},
			Count: oCount{2},
		},
		optionalObject{
			Name:  oName{"memory"},
			Count: oCount{1},
			Unit:  oUnit{"Gi"},
		},
		optionalObject{
			Name:  oName{"volumes"},
			Count: oCount{3},
		},
		optionalObject{
			Name:  oName{"storage"},
			Count: oCount{100},
			Unit:  oUnit{"Gi"},
		},
	}

	i := expectedInput{ProjectName: "boogie-test", Environment: "dev", Optionals: o}

	fileName, baseObject := createLimitsObject(&i)
	// verify contents are correct
	gotBytes, err := json.Marshal(baseObject)
	if err != nil {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
	}
	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}
	expectedObjectName := quotaFilename
	if expectedObjectName != fileName {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedObjectName, fileName)
	}

	expectedBytes = []byte(`{"kind":"ResourceQuota","apiVersion":"v1","metadata":{"name":"default-quotas","namespace":"boogie-test"},"spec":{"hard":{"limits.cpu":1,"limits.memory":"5Gi","requests.storage":"5Gi"}}}`)

	o = []optionalObject{
		optionalObject{
			Name:  oName{"cpu"},
			Count: oCount{1},
		},
		optionalObject{
			Name:  oName{"memory"},
			Count: oCount{5},
			Unit:  oUnit{"Gi"},
		},
		optionalObject{
			Name:  oName{"storage"},
			Count: oCount{5},
			Unit:  oUnit{"Gi"},
		}}

	i = expectedInput{ProjectName: "boogie-test", Environment: "dev", Optionals: o}

	fileName, baseObject = createLimitsObject(&i)
	// verify contents are correct
	gotBytes, err = json.Marshal(baseObject)
	if err != nil {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
	}
	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}

	expectedObjectName = quotaFilename
	if expectedObjectName != fileName {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedObjectName, fileName)
	}

	expectedBytes = []byte(`{"kind":"","apiVersion":"","metadata":{"name":""},"spec":{"hard":{}}}`)

	i = expectedInput{ProjectName: "boogie-test", Environment: "dev"}

	fileName, baseObject = createLimitsObject(&i)
	// verify contents are correct
	gotBytes, err = json.Marshal(baseObject)
	if err != nil {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
	}
	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}
	expectedObjectName = ""
	if expectedObjectName != fileName {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedObjectName, fileName)
	}
}

func TestIsEmptyObject(t *testing.T) {
	q := quota{}
	isEmpty := isEmptyObject(q)
	if !isEmpty {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "empty", "not empty")
	}

	q = quota{}
	q.Kind = "whatever"

	isEmpty = isEmptyObject(q)
	if isEmpty {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "not empty", "empty")
	}
}

func TestCreateNewLimitsObjectCPU(t *testing.T) {
	expectedBytes := []byte(`{"kind":"ResourceQuota","apiVersion":"v1","metadata":{"name":"default-quotas","namespace":"boogie-test"},"spec":{"hard":{"limits.cpu":"200m","limits.memory":"1Gi","persistentvolumeclaims":3}}}`)

	o := []optionalObject{
		optionalObject{
			Name:  oName{"cpu"},
			Count: oCount{200},
			Unit:  oUnit{"m"},
		},
		optionalObject{
			Name:  oName{"memory"},
			Count: oCount{1},
			Unit:  oUnit{"Gi"},
		},
		optionalObject{
			Name:  oName{"volumes"},
			Count: oCount{3},
		}}

	i := expectedInput{ProjectName: "boogie-test", Environment: "dev", Optionals: o}

	fileName, baseObject := createLimitsObject(&i)
	// verify contents are correct
	gotBytes, err := json.Marshal(baseObject)
	if err != nil {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
	}
	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}
	expectedObjectName := quotaFilename
	if expectedObjectName != fileName {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedObjectName, fileName)
	}

}
