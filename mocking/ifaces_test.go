package mocking_test

import (
	"github.com/stretchr/testify/assert"
	"minimalgo/mocking"
	"testing"
)

type BasePersonMock struct {
}

func (m *BasePersonMock) PrintName() string {
	return "Fred"
}
func (m *BasePersonMock) PrintLastName() string {
	return "Smith"
}

//FrankMock only implements PrintName()
type FrankMock struct {
	BasePersonMock
}

func (m *FrankMock) PrintName() string {
	return "Frank"
}

func TestPerson_PrintNameMocked(t *testing.T) {
	var personInterface mocking.PersonInterface

	personInterface = &BasePersonMock{}
	assert.Equal(t, "Fred", personInterface.PrintName())
	assert.Equal(t, "Smith", personInterface.PrintLastName())

	personInterface = &FrankMock{}
	assert.Equal(t, "Frank", personInterface.PrintName())
	assert.Equal(t, "Smith", personInterface.PrintLastName()) //uses the method from BasePersonMock
}

func TestPerson_PrintName(t *testing.T) {
	var person *mocking.Person              //Nil
	assert.Equal(t, "", person.PrintName()) //No nil pointer!! Nil has a Noop implementation

	person2 := &mocking.Person{
		Name: "Paul",
	}
	assert.Equal(t, "Paul", person2.PrintName())
}

func TestPerson_PrintNamePanic(t *testing.T) {
	var personInterface mocking.PersonInterface
	personInterface.PrintName() //this will cause a panic because the interface is nil and there is no implementation to catch the call
}
