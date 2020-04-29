package main

import (
	"testing"
)

func getFakeInput() *expectedInput {
	return &expectedInput{}
}

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
