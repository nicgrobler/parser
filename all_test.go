package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getFakeInput() *expectedInput {
	return &expectedInput{}
}

func TestInferADGroupName(t *testing.T) {
	i := expectedInput{Environment: "boogie", Role: "Admin", ProjectName: "extra-good"}
	got := inferADGroupName(&i)
	want := "RES-BOOGIE-OPSH-ADMIN-EXTRA_GOOD"
	assert.Equal(t, want, got, "should be equal")
}

func TestCreateTouchfileName(t *testing.T) {
	i := expectedInput{Environment: "boogie"}
	got := createTouchfileName(&i)
	want := "OPSH_ENV.BOOGIE"
	assert.Equal(t, want, got, "should be equal")
}

func TestLookupRole(t *testing.T) {
	// should be ok
	got := lookupRole("Admin")
	want := "admin"
	assert.Equal(t, want, got, "should be equal")
	// should return nothing
	var errorMessage string
	exitLog = func(message string) { errorMessage = message }
	got = lookupRole("Administrator")
	want = ""
	assert.Equal(t, want, got, "should be equal")
	errorShouldBe := "invalid user type specified"
	assert.Equal(t, errorShouldBe, errorMessage, "should be equal")
}
