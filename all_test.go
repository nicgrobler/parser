package main

import (
	"testing"
)

func TestInferADGroupName(t *testing.T) {
	i := expectedInput{Environment: "boogie", Role: "Admin", ProjectName: "extra-good"}
	got := inferADGroupName(&i)
	want := "RES-BOOGIE-OPSH-ADMIN-EXTRA_GOOD"
	if want != got {
		t.Errorf("wanted %s, but got %s: \n", want, got)
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

func TestLookupRole(t *testing.T) {
	// should be ok
	got := lookupRole("Admin")
	want := "admin"
	if want != got {
		t.Errorf("wanted %s, but got %s: \n", want, got)
	}
	// should return nothing
	var errorMessage string
	exitLog = func(message string) { errorMessage = message }
	got = lookupRole("Administrator")
	want = ""
	if want != got {
		t.Errorf("wanted %s, but got %s: \n", want, got)
	}
	errorShouldBe := "invalid user type specified"
	if want != got {
		t.Errorf("wanted %s, but got %s: \n", errorShouldBe, errorMessage)
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
	expectedFileName := "boogie-test-new-project.json"
	if expectedFileName != fileName {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedFileName, fileName)
	}

}

func TestCreateNewRoleBindingFile(t *testing.T) {
	expectedBytes := []byte(`{
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
}`)

	i := expectedInput{ProjectName: "boogie-test", Environment: "dev", Role: "developer"}

	fileName, gotBytes := createNewRoleBindingFile(&i)
	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}
	expectedFileName := "boogie-test-new-rolebinding.json"
	if expectedFileName != fileName {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedFileName, fileName)
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
      "persistentvolumeclaims": 3
    }
  }
}`)

	o := optionalObjects{
		optionalObject{
			Name:  "cpu",
			Count: 2,
		},
		optionalObject{
			Name:  "memory",
			Count: 1,
			Unit:  "Gi",
		},
		optionalObject{
			Name:  "volumes",
			Count: 3,
		},
	}

	i := expectedInput{ProjectName: "boogie-test", Environment: "dev", Role: "developer", Optionals: &o}

	fileName, gotBytes := createNewLimitsFile(&i)
	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}
	expectedFileName := "boogie-test-new-quota.json"
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
      "limits.memory": "5Gi"
    }
  }
}`)

	o = optionalObjects{
		optionalObject{
			Name:  "cpu",
			Count: 1,
		},
		optionalObject{
			Name:  "memory",
			Count: 5,
			Unit:  "Gi",
		},
	}

	i = expectedInput{ProjectName: "boogie-test", Environment: "dev", Role: "developer", Optionals: &o}

	fileName, gotBytes = createNewLimitsFile(&i)
	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}

	expectedFileName = "boogie-test-new-quota.json"
	if expectedFileName != fileName {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedFileName, fileName)
	}

	expectedBytes = nil

	i = expectedInput{ProjectName: "boogie-test", Environment: "dev", Role: "developer"}

	fileName, gotBytes = createNewLimitsFile(&i)
	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}
	expectedFileName = ""
	if expectedFileName != fileName {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedFileName, fileName)
	}
}
