package main

import (
	"encoding/json"
	"testing"
)

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

func TestCreateTouchfileName(t *testing.T) {
	i := expectedInput{Environment: "boogie"}
	got := createTouchfileName(&i)
	want := "OPSH_ENV.BOOGIE"
	if want != got {
		t.Errorf("wanted %s, but got %s: \n", want, got)
	}
}

func TestCreateNewProjectFile(t *testing.T) {

	expectedBytes := []byte(`{
  "kind": "Project",
  "apiVersion": "project.openshift.io/v1",
  "metadata": {
    "name": "boogie-test"
  }
}`)

	i := expectedInput{ProjectName: "boogie-test"}

	fileName, gotBytes := createNewProjectFile(&i)
	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}
	expectedFileName := "1-boogie-test-new-project.json"
	if expectedFileName != fileName {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedFileName, fileName)
	}

}

func TestCreateNewRoleBindingFile(t *testing.T) {

	var expectedBytes [][]byte

	expectedBytes = append(expectedBytes, []byte(`{
  "kind": "RoleBinding",
  "apiVersion": "rbac.authorization.k8s.io/v1",
  "metadata": {
    "name": "boogie-test-edit-binding",
    "namespace": "boogie-test"
  },
  "subjects": [
    {
      "kind": "Group",
      "apiGroup": "rbac.authorization.k8s.io",
      "name": "RES-DEV-OPSH-DEVELOPER-BOOGIE_TEST"
    }
  ],
  "roleRef": {
    "kind": "ClusterRole",
    "apiGroup": "rbac.authorization.k8s.io",
    "name": "edit"
  }
}`))
	expectedBytes = append(expectedBytes, []byte(`{
  "kind": "RoleBinding",
  "apiVersion": "rbac.authorization.k8s.io/v1",
  "metadata": {
    "name": "boogie-test-view-binding",
    "namespace": "boogie-test"
  },
  "subjects": [
    {
      "kind": "Group",
      "apiGroup": "rbac.authorization.k8s.io",
      "name": "RES-DEV-OPSH-VIEWER-BOOGIE_TEST"
    }
  ],
  "roleRef": {
    "kind": "ClusterRole",
    "apiGroup": "rbac.authorization.k8s.io",
    "name": "view"
  }
}`))

	expectedBytes = append(expectedBytes, []byte(`{
  "kind": "RoleBinding",
  "apiVersion": "rbac.authorization.k8s.io/v1",
  "metadata": {
    "name": "boogie-test-admin-relman-binding",
    "namespace": "boogie-test"
  },
  "subjects": [
    {
      "kind": "ServiceAccount",
      "name": "relman",
      "namespace": "relman"
    }
  ],
  "roleRef": {
    "kind": "ClusterRole",
    "apiGroup": "rbac.authorization.k8s.io",
    "name": "admin"
  }
}`))

	i := expectedInput{ProjectName: "boogie-test", Environment: "dev"}

	fileNames, gotBytes := createNewRoleBindingFiles(&i)
	if string(expectedBytes[0]) != string(gotBytes[0]) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes[0], gotBytes[0])
	}
	expectedFileName := "10-boogie-test-new-edit-rolebinding.json"
	if expectedFileName != fileNames[0] {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedFileName, fileNames[0])
	}

	if string(expectedBytes[1]) != string(gotBytes[1]) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes[1], gotBytes[1])
	}
	expectedFileName = "10-boogie-test-new-view-rolebinding.json"
	if expectedFileName != fileNames[1] {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedFileName, fileNames[1])
	}

	if string(expectedBytes[2]) != string(gotBytes[2]) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes[2], gotBytes[2])
	}
	expectedFileName = "10-boogie-test-new-default-rolebinding.json"
	if expectedFileName != fileNames[2] {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedFileName, fileNames[2])
	}

}

func TestCreateNewLimitsFile(t *testing.T) {
	expectedBytes := []byte(`{
  "kind": "ResourceQuota",
  "apiVersion": "v1",
  "metadata": {
    "name": "default-quotas",
    "namespace": "boogie-test"
  },
  "spec": {
    "hard": {
      "limits.cpu": 2,
      "limits.memory": "1Gi",
      "persistentvolumeclaims": 3,
      "requests.storage": "100Gi"
    }
  }
}`)

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

	fileName, gotBytes := createNewLimitsFile(&i)
	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}
	expectedFileName := "10-boogie-test-new-quota.json"
	if expectedFileName != fileName {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedFileName, fileName)
	}

	expectedBytes = []byte(`{
  "kind": "ResourceQuota",
  "apiVersion": "v1",
  "metadata": {
    "name": "default-quotas",
    "namespace": "boogie-test"
  },
  "spec": {
    "hard": {
      "limits.cpu": 1,
      "limits.memory": "5Gi",
      "requests.storage": "5Gi"
    }
  }
}`)

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

	fileName, gotBytes = createNewLimitsFile(&i)
	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}

	expectedFileName = "10-boogie-test-new-quota.json"
	if expectedFileName != fileName {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedFileName, fileName)
	}

	expectedBytes = nil

	i = expectedInput{ProjectName: "boogie-test", Environment: "dev"}

	fileName, gotBytes = createNewLimitsFile(&i)
	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}
	expectedFileName = ""
	if expectedFileName != fileName {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedFileName, fileName)
	}
}

func TestCreateNewLimitsFileCPU(t *testing.T) {
	expectedBytes := []byte(`{
  "kind": "ResourceQuota",
  "apiVersion": "v1",
  "metadata": {
    "name": "default-quotas",
    "namespace": "boogie-test"
  },
  "spec": {
    "hard": {
      "limits.cpu": "200m",
      "limits.memory": "1Gi",
      "persistentvolumeclaims": 3
    }
  }
}`)

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

	fileName, gotBytes := createNewLimitsFile(&i)
	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}
	expectedFileName := "10-boogie-test-new-quota.json"
	if expectedFileName != fileName {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedFileName, fileName)
	}

}
