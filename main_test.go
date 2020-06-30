package main

import (
	"testing"
)

func TestValidation(t *testing.T) {

	FirstName := "aleksei"
	LastName := "aleksei"
	Gender := "aleksei"
	Address := "aleksei"
	Email := "aleksei"

	result := isFormError(&FirstName, &LastName, &Gender, &Address, &Email)

	if !result {
		t.Error("Form validated by error", result)
	}

	FirstName = "aleksei"
	LastName = "aleksei"
	Gender = "aleksei"
	Address = ""
	Email = "aleksei@gmail.com"

	result = isFormError(&FirstName, &LastName, &Gender, &Address, &Email)

	if !result {
		t.Error("Form validated by error", result)
	}

	FirstName = "alekseiawdddddddddddddddddddddddddddddwadddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd" +
		"ddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddawdddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
	LastName = "aleksei"
	Gender = "aleksei"
	Address = "wdawdawdawd"
	Email = "aleksei@gmail.com"

	result = isFormError(&FirstName, &LastName, &Gender, &Address, &Email)

	if !result {
		t.Error("Form validated by error", result)
	}
}
