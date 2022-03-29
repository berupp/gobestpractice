# gobestpractice

A collection of best practices, conventions, patterns, tips and gotchas around golang and building modules

* [Completely new to golang?](#new-to-golang)
* [How-to write a proper golang module](#how-to-write-a-proper-golang-module)
    * [Package](#package)
      * [Constructor patterns](#constructor-patterns)
      * [Package functions](#package-functions)
      * [Exposing errors](#exposing-errors)
      * [Logging](#logging)
    * [Make it a module](#make-it-a-module)
* [Unit testing](#unit-testing)
    * [Variadic tests](#variadictests)
    * [Mocking strategies](#mocking-strategies)
* [Integration tests](#integration-tests)
* [Go routines](#go-routines)
* [Go channels](#go-channels)
  * [Patterns](#patterns)
* [Synchronization techniques](#synchronization-techniques)

# New to golang 
If you are completely new to golang, please first read and work through [GoByExample](https://gobyexample.com/). It is the best introduction to golang syntax and concepts. Once you worked through that, you are essentially
able to write any program in go. But as always when you learn a new language, adopting the language features and concepts in the right context is difficult. This is what the following 
topics try to help with.


# How-to write a proper golang module

A good golang module is a small re-usable component with a clearly defined and documented API. Golang relies on
certain conventions, applying them naturally leads to the creation of high quality modules. Using conventions also facilitate testability and 
above all: maintainability of the code.

## Package

The smallest unit that can be exposed as a go module is a `package`. A module may consist of multiple packages, but each 
of those would also be their own module. The package:

* must only expose types, functions, constants and variables which are supposed to be publicly visible. Optimally as few entities as possible. This text refers this as the API or package API
* must have a descriptive name that serves as part of the API flow. It should be short, avoid `camelCase` or `under_score` and not 'steal' popular variable names. [More details](https://go.dev/blog/package-names)
* must have a `func New(...)` constructor for all exposed package objects
* can have static functions
* must have tests

By convention a package exposes one or more objects that are created using `New()` constructors

Example:
```go
package dgraph
type Client struct{}
func New() *Client { return &Client{}}
```
A `New()` constructor can have any number of parameters and descriptive suffixes to allow different ways to initialize the object.

Good :thumbsup:
```go
//If package only exposes single package object (ideal unit)
dgraphClient := dgraph.New()
//If package exposes multiple different package objects
dgraphClient := dgraph.NewClient()
dgraphMonitor := dgraph.NewMonitor()

//Dependency Injection
dgraphClient := dgraph.New(connection)
//Overloading - See Constructor patterns for better alternative
dgraphClient := dgraph.NewWithTLS(connection, tlsConfig)
//Allow error handling
dgraphClient, err := dgraph.New()
```

By convention, alliteration must be avoided.

Bad :thumbsdown:
```go
//Alliteration
dgraphClient := dgraph.NewDgraphClient()

//Lots of parameters are always an incompatibility risk, see: Constructor patterns how to avoid
dgraphClient := dgraph.NewClient(ip, port, domain, errorHandler, prometheusCLient, snmpTrap)

//Not using New... convention
dgraphClient := dgraph.CreateClient()
//Not providing constructor function at all, leaving the API user to guess how to instantiate the object
dgraphClient := new(dgraph.Client)
dgraphClient := &dgraph.Client{}
```

### Constructor patterns
When initializing objects using a `New()` constructor, you often need to customize the returned object. Using a list of parameters is
discouraged, as golang does not allow overloading. Therefore, adding parameters will break the API or require the creation of verbose `NewWithXAndYAndZ()`
constructors. This can quickly become a permutation problem. Instead, you should use one of the two patterns described below, to configure objects on initialization.

#### The standard
The default constructor pattern that allows non-breaking addition of parameters is using a configuration object
```go
type Config struct {
	ip string
	port int
}
```
The `New()` constructor has usually a single parameter, the config object, which allows additions in the future without breaking backwards compatibility
```go
dgraphClient := dgraph.New(&dgraph.Config{
	ip : "localhost",
	port: 9090
})
```
For complex configuration objects, it is good practice creating a constructor function that returns a configuration object with proper defaults.
```go
config := dgraph.NewDefaultConfig()
dgraphClient := dgraph.New(config)
```

#### Options pattern
The functional-style options pattern is a very nice to use, yet a bit verbose to implement, alternative.

For public facing APIs with lots of different options, it is a good choice. The `New()` constructor takes 
a variadic variable of `connectionOption`. The two provided implementations are `WithIP(ip string)` and `WithPort(port int)`.

The `New(...)` constructor first initializes the `Connection` object with default parameters and then iterates through the list of provided 
options to apply them to the `Connection` object before returning it.

```go
package options_pattern

type Connection struct {
	ip   string
	port int
}

type connectionOption func(*Connection)

func WithIP(ip string) func(connection *Connection) {
	return func(connection *Connection) {
		connection.ip = ip
	}
}
func WithPort(port int) func(connection *Connection) {
	return func(connection *Connection) {
		connection.port = port
	}
}

func New(opts ...connectionOption) *Connection {
	conn := &Connection{
		ip:   "default",
		port: 0,
	}
	//Apply all options
	for idx := range opts {
		opts[idx](conn)
	}
	return conn
}
```
Using this pattern makes for very intuitive and descriptive API calls that allow backward-compatible option additions and deprecation:
```go
func TestOptionsPattern(t *testing.T) {
	{
		connection := options_pattern.New(options_pattern.WithIP("localhost"))
		
		assert.Equal(t, "ip: localhost, port: 0", connection.ToString())
	}
	{
		connection := options_pattern.New(
			options_pattern.WithIP("localhost"),
			options_pattern.WithPort(90008))

		assert.Equal(t, "ip: localhost, port: 90008", connection.ToString())
	}
}
```

---
:warning: You have the option to prevent struct initialisation if a `New()` constructor is offered. Make the object `type`
package private (lowercase):
```go
package dgraph
type client struct{} //Nobody can initialize this struct now
func New() *client { return &client{}}
```
This comes with the huge downside that package users are unable to reference the type directly, which prevents variable declarations and can make working
with the package cumbersome. It is therefore **discouraged**.
```go
package persistence

var dgraphClient *dgraph.client //Does not compile, package user cannot declare variables of your type

func smh() {
	dgraphClient := dgraph.New() //No issues
}
```
---

### Package functions

Package functions are typically used to configure a package within the context of the application. 
An example would be providing a logger implementation if the package supports it or simply setting a log level.

```go
dgraph.SetLogger(log.Logger) 
dgraph.SetLogLevel(dgraph.DEBUG)
```

Another use case for package functions are utility operations. A good example for this is golangs `strings` package:
```go
hasLa := strings.Contains("Blabla", "la")
helloWorld := strings.Join([]string{"hello", "world"}, ",")
```

:bulb: **Bad practice: Utility packages**

Do NOT EVER create a utility package. These attempts to reuse code WILL fail because context-less utility 
packages will never be known to the next maintainer. They will not find the functions they are supposed to reuse and duplicate them anyway.
Utility functions should be provided in the package context where they are valuable. In most cases they can be replaced with a method on an object.
```go
//Bad
util.ConvertXToY
//Better on the object. Has tool support (IDE auto-completion)  
y := x.ConvertToY()

//Bad, nobody will ever know it exists
util.CalculateGeoCentroid(long,lat float32) 
//Better in a geo package for context, maintainer more likely to look here for a geo centroid function
geo.CalculateCentroid(long,lat float32) 
```

## Exposing Errors

In 99.9% cases, using `fmt.Errorf()` is the recommended way to return `error`s to your API users. It is used to create errors or add further context to downstream errors:

```go
//Returning an error
return fmt.Errorf("login failed to due invalid credentials")

//Returning a parameterized error
return fmt.Errorf("login failed to due invalid credentials for user %s", username)

//Adding context to an existing error
return fmt.Errorf("errors connecting to postgres: %s", err.Error())

//Bad capitalized error message
return fmt.Errorf("Try again")
```
:bulb: Go convention forbids capitalized error messages unless the first word is a proper noun

But sometimes you want to enable API users to distinguish different error scenarios. There are essentially two ways
to accomplish that:
* Package defined errors
* Custom error types

#### Package defined errors

Package defined errors are your go-to solution if you have a function that can return multiple error conditions
that the user wants to handle differently. An example would be a database call, where the driver connection could fail (retry) 
or the SQL could be corrupted (programming error, do not retry).

Note that package defined errors have a **big disadvantage**: They cannot be parameterized. As such their use is limited

Define package error
```go
package errorhandling

import "fmt"

var (
	//ConnectionError is a package defined error that allows users to react to different error conditions, 
	//but cannot be parameterized
	ConnectionError = fmt.Errorf("connection failed")
)
```
Handling package defined errors as the API consumer:
```go
func TestPreDefinedErrorHandling(t *testing.T) {
	err := error_handling.ReturnPredefinedError()
	//You can use a switch
	switch err {
	case error_handling.ConnectionError:
		assert.Equal(t, "connection failed", err.Error())
	default:
		t.Fatal("unexpected error")
	}

	//Or a simple comparison
	if err == error_handling.ConnectionError {
		assert.Equal(t, "connection failed", err.Error())
	} else {
		t.Fatal("unexpected error")
	}
}
```
Use of this approach ONLY if you can get away with non-parameterized errors on a function that can return errors which require
individual handling.

#### Custom error types

Custom error types overcome the limitation of package defined errors: they can be extensively parameterized.
Create a type that implements `func Error() string`. This satisfies the `error` interface.

Example of defining and returning a custom error type
```go
type CustomError struct {
	Status int
	Reason string
}

//Error satisfies the error interface. Be aware what you return if you implement with or without pointer receiver
func (e CustomError) Error() string {
	return fmt.Sprintf("failed with status %d: %s", e.Status, e.Reason)
}

func ReturnCustomError() error {
    //Returning CustomError value (not pointer) as Error() is implemented with value receiver
    return CustomError{
        Status: 22,
        Reason: "Just cause",
    }
}
```

Handling a custom error type as API caller:

```go
func TestCustomErrorHandling(t *testing.T) {
	err := errorhandling.ReturnCustomError()

	switch err.(type) {
	case errorhandling.CustomError:
		customError := err.(errorhandling.CustomError) //This typecast is considered fine as we did the type check already
        
		//Inspect and handle the error
		assert.Equal(t, 22, customError.Status)
		assert.Equal(t, "Just cause", customError.Reason)
	default:
		t.Fatal("unexpected error")
	}
}
```
ONLY use custom error types if you have a function that returns errors which require individual handling. This should be used rarely!

### Logging

The best way to provide logging in your package is to define a logger `interface` and allow users to set their own logger.
When you define a logging interface, stay as simple as possible to permit compatibility with popular logging libraries:

```go
package packagelog

//Logger is the module's logging interface. It is compatible with the standard os logger.
//It won't be great for a lot of popular logging libraries, which commonly use InfoF and Errorf instead
type Logger interface {
	Printf(l string, args ...interface{})
	Fatalf(l string, args ...interface{})
	
	//Infof and Errorf give a lot of compatibility with existing logging libraries
	//Infof(l string, args ...interface{})
	//Errorf(l string, args ...interface{})
	
	//Warnf (sometimes Warningf) and Debugf are less common and should be avoided
	//Warnf(l string, args ...interface{})
	//Debugf(l string, args ...interface{})
}

//NoopLogger is the default provided logger
type NoopLogger struct{}

func (NoopLogger) Printf(l string, args ...interface{}) {}
func (NoopLogger) Fatalf(l string, args ...interface{}) {}

//moduleLogger is used for all logs of the module
var moduleLogger Logger = NoopLogger{}

//SetLogger allows the package user to provide his own implementation
func SetLogger(l Logger) {
	moduleLogger = l
}

func MyCoolFunction(name string, age int) {
	moduleLogger.Printf("Name: %s, Age: %d", name, age)
}
```
The package user can now provide his own logger, which ideally is simply his application logger:

```go
func TestMyCoolFunction(t *testing.T) {
	packagelog.MyCoolFunction("Paul", 43) //Logs nothing
	packagelog.SetLogger(log.Default()) //Set standard library's log.Logger
	packagelog.MyCoolFunction("Jill", 84) //Logs using standard library logger
}
```
If the application logger does not satisfy your logging interface natively, the package user can still build an adapter
that satisfies the interface and delegates to his application logger.

---
:bulb:
This example demonstrates the difference of interfaces in go and Java. In Java, a `class` has to explicitly 
implement an interface. In go, a `type` has to simply _satisfy_ the interface, by implementing all functions 
defined by it.


---
### Make it a module

Enter your package on command line and execute `go mod init`, which creates the `go.mod` file. You will be asked to run `go mod tidy` after to add
all required dependencies.

Congratulations, you created a reusable module which you obtain in other projects by using `go get ...`

:bulb: When using a module that is part of (i.e.: a package in) your project, you can add the following line to you projects `go.mod` file:

```go
replace github.com/Accedian/stitchIt/persistence => ./persistence
```
This directive makes it so that all imports of `github.com/Accedian/stitchIt/persistence` actually use the local package `persistence`.
Note that go 1.18 introduces _workspaces_, which is a more advanced feature deal with multi-module environments.


# Unit testing
Unit testing as important as the code you write. It assures that your 

* code is working
* code is well-structured (as components or 'units')
* package API is well-defined and encapsulated

If you can't test your code with either unit or at least integration tests, refactor it until you can. Code that can't be tested is badly designed even if the logic is genius.

## Variadic tests

By default, you should use variadic tests for the following benefits

* Reduced repetition of code for similar test cases
* Allows easy setup and tear down without using the package global `TestMain`
* Structure better to maintain
* **Fast addition of new test cases**

The last point being key: Variadic tests allow the quick addition of new testcases if a bug is found. 

```go
package variadictests

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestToLowerCase(t *testing.T) {
	//Test 'suit' set up here
	// ... setup ...
	defer func() {
		//Optimal test 'suit' tear down here:
		// ... tear down ...
	}()

	var tests = []struct {
		Name           string
		Input          string
		ExpectedOutput string
	}{
		{
			Name:           "Uppercase",
			Input:          "UPPERCASE",
			ExpectedOutput: "uppercase",
		},
		{
			Name:           "CamelCase",
			Input:          "CamelCase",
			ExpectedOutput: "camelcase",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			//per-test setup
			lower := strings.ToLower(test.Input)
			assert.Equal(t, test.ExpectedOutput, lower)
			//per-test tear down
		})
	}
	//Optional test 'suit' tear down here
	// ... tear down that may not execute on test failure ...
}
```

#### Tip: Complex type definition

```go

func TestToLowerCaseComplexObject(t *testing.T) {
	var tests = []struct {
		Name           string
		Input          func() []*string //func() provides a nice way to set p more complex inputs
		ExpectedOutput []string
	}{
		{
			Name: "Uppercase",
			Input: func() []*string {
				a := "UPPERCASE"
				return []*string{&a}
			},
			ExpectedOutput: []string{"uppercase"},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			input := test.Input() //Use the input
			assert.Equal(t, 1, len(input))
		})
	}
}
```
The `func()` is typically used to allow short-hand for setting up
* time
* pointers for scalar values
* complex objects


## Mocking strategies

### Provider mocking

Use where you need to mock out dependencies because the surrounding code must be testable in a unit test and cannot be tested with an integration test instead.
Usage of this should therefore be rare, since code around dependencies can usually be broken down into testable units, or are so intertwined with 
the dependency code, that an integration test is required.

Assume the following code: A lookup is made in a database and the resulting string is transformed to upper case. We want to test the upper case transformation, mocking out the database call.
To allow easy mocking, we implement the entire database access logic in a function `var GetStringFromDatabase`. This is a _Provider_ :
```go
package mocking

import "strings"

var GetStringFromDatabase = func(entityId string) string {
	//This is the production code accessing the database, ready a string by its ID and returning it
	return ""
}

func ToUpperCaseFromDatabase(input string) string {
	fromDB := GetStringFromDatabase(input) //mocked in test
	//code under test
	return strings.ToUpper(fromDB)
}
```

This allows us to exchange the implementation in our Provider in the test with a simple variable assignment:
```go
func TestToUpperCaseFromDatabase(t *testing.T) {
    original := mocking.GetStringFromDatabase //Store original to restore behavior after test
    defer func() {
        mocking.GetStringFromDatabase = original //Make sure behavior is restored after test
    }()
    
    mocking.GetStringFromDatabase = func(string) string { //Mock    
		return "mock"
    }
    
    result := mocking.ToUpperCaseFromDatabase("abc")
    assert.Equal(t, "MOCK", result)
}
```
* Pros
  * Quick and easy way to stub out dependencies with small, locally scoped mocks
  * Superior mock strategy compared interface mocking or tool generated mocks 
* Cons
  * Pollutes package API 
  :bulb: If you want to avoid pollution, you may choose to make the provider function private (`var getStringFromDatabase = func(entityId string) string`). Your tests cannot be in the `_test` package in this case. 
  * Providers can feel convoluted to implement
  * Mock problem: Garbage in garbage out. Wrong assumptions regarding the mock implementation render test useless

### Interface mocking

Interfaces are not the same in golang as they are in Java. In Java, interfaces define a contract to create abstraction and promote loose coupling.
Interfaces are heavily tied to inheritance and generics.

In golang interfaces are not as useful for abstraction purposes due to limited inheritance and missing generics. 

Even with golang 1.18 and the introduction of [generics](https://go.dev/doc/tutorial/generics), interfaces are not
playing a role in generics implementations.

Therefore, abstracting a package object with and interface is not commonly done as you'd see it in Java and as a result,
interface mocking a 3rd party library like a database client, is often not an option.

Interfaces are typically used to define a contract in your package API, for users to satisfy.
Allowing package API users to [set their own logger](#logging) is a great example for this.


When mocking interfaces, keep the following in mind:

* Avoid mock tooling (3rd party generators and syntax libraries): They require ramp up, often use obscure verification syntax and can introduce instability into the build
* Interface mocking can be very verbose

When mocking interfaces, you can create a base mock and extend from it to reduce verbosity of the mock code

```go
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
```

---
:bulb:

If your package API exposes a `type` instead of an `interface`, it is good practice providing a noop implementation,
by `nil`-checking the pointer receiver on each method and executing a default behavior. This prevents nil-pointers and 
promotes loose coupling, by essentially making the dependency on you package optional. 


```go
type Person struct {
	Name     string
	LastName string
}

func (m *Person) PrintName() string {
	if m == nil {
		return "" //Noop implementation
	}
	return m.Name
}
func (m *Person) PrintLastName() string {
	if m == nil {
		return ""
	}
	return m.LastName
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
```


---


### Mocking REST calls

We use a lot of REST interactions between microservices. An alternative to provider mocking for those interactions
is to start a small server in your test and mock the REST response. An example function could look like below:

```go
const mockPort = 55575

func startMockServer(t *testing.T, expectedRequest []byte, responseCode int, response []byte) (*httptest.Server, error) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		//Example: Verify the request body
		assert.Equal(t, expectedRequest, data, "Expected %s, Got %s", string(expectedRequest), string(data))

		//Mock desired response
		w.WriteHeader(responseCode)
		w.Write(response)
	}))
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", mockPort))
	if err != nil {
		return nil, err
	}
	server.Listener = l
	server.Start()
	return server, nil
}
```

Calling `startMockServer`, we provide the expected request body for the server to verify, as well as the `responseCode` and `response` body
we want to respond with.

Imagine a function `DoPOST`:
```go
func DoPOST(url string, body string) error {
	request, err := http.NewRequest("POST", url, ioutil.NopCloser(bytes.NewReader([]byte(body))))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}
	return nil
}
```

Now we could test the function using our mock server:
```go
func TestDoPOST(t *testing.T) {
	server, err := startMockServer(t, []byte("Hello world"), 200, []byte(``))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer server.Close() //Do not forget this

	//Send the request to the mocked endpoint
	err = mocking.DoPOST(fmt.Sprintf("http://localhost:%d", mockPort), "Hello world")
	assert.Nil(t, err)
}
```
This test fails if we change the response code to something other than `200`, or the payload is different from `Hello world`. This method allows for testing
request expectations as well as different response scenarios.

:warning: `go test` runs multiple tests in parallel by default unless you pass the `-p 1` flag. If you use mock servers, executing tests in 
parallel can quickly end in `port already in use` errors. To prevent these errors, you need to either come up with a synchronized way to assign unique ports to each
mock server instance, or run `go test -p 1`.




# Integration Tests

Integration tests are required wherever you interact with other services. The idea is to be as close to a production environment 
as possible, while keeping test execution times and resource utilization as low as possible. We also want to write the integration tests
in `go`, and avoid using obscure frameworks that require ramp up.

The basis for integration tests is `docker-compose`. Compose is a community adopted, well-known technology to define services.

Start by declaring all your required dependencies, as well as your service-under-test in a docker-compose.yml:
```yaml
version: "3.3"

services:
  dgraph:
    image: dgraph/standalone:v21.03.2
    ports:
    - "9080:9080"
    - "8080"
  stitchit:
    build:
      context: ../
      dockerfile: Dockerfile
    command: -mode dgraph -dgraph integration_tests_dgraph_1:9080
    environment:
      - LOGLEVEL=DEBUG
    depends_on:
      - dgraph
    restart: on-failure
```
In this example we define `stitchit`, which is the service-under-test. The `stitchit` container is built every time we change some code, so we
always test against the latest version. The compose file also defines the dependency `dgraph`.

Running `docker-compose -f docker-compose.yml up -d --build`, will build the new `stitchit` container and start both the `stitchit` and `dgraph` services.
To tear the whole thing down, including volumes, simply execute: `docker-compose -f integration_tests/docker-compose.yml down -v`.

Now that we have our service-under-test with minimal dependencies running, we just need to run our integration tests against it.

Our integration tests are standard golang unit tests, executed with `go test`. The advantage is that there is no learning curve for integration tests.

Example:
```go
//go:build integration
// +build integration

package integration_tests

func TestREST(t *testing.T) {
    body := []byte(`{
                            "data": {
                                "attributes": {
                                    "name": "Example",
                                    "location": {
                                        "latitude": 9.66654,
                                        "longitude": -8.5554
                                    }
                                },
                                "type": "asset"
                            }
                        }`)
    request, err := http.NewRequest(http.MethodPost, "http://integration_tests_stitchit_1:8080/api/v1/stitchit/assets", ioutil.NopCloser(bytes.NewReader(body)))
    if err != nil {
        t.Fatal(err.Error())
    }
    request.Header.Set(XForwardedTenantId, testTenant)
    request.Header.Set("Content-Type", "application/json")
    response, err := httpClient.Do(request)
    if err != nil {
        t.Fatal(err.Error())
    }
    assert.Equal(t, 201, response.StatusCode)
}
```
This test executes a REST request against the service-under-test and verifies the response. 

In order to interact with the service-under-test, our integration tests must have access to the same docker network.
The trick is to execute `go test` inside a docker container, which is part of the appropriate docker network: We `docker run`
an alpine image, mount the code, set the `GOPATH` and execute `go test` with the appropriate parameters. In this case we only run tests in the
_integration_tests_ package (`-v`) that have the [build tag]() `integration`. The build tag allows integration tests to be ignored when executing unit tests (`go test ./...`)

```makefile
docker run -it --rm --name stitchit_itest --network=integration_tests_default \
           -e GOPATH=/root/go -v "$(GOPATH):/root/go" -v "$(PROJECT_BASE_PATH):/root/workingdir" \
           -w "/root/workingdir" docker-go-sdk:0.42.0-alpine go test -v ./integration_tests/... -tags integration
```

Now we smash all this into a nice Makefile target for convenient one-line execution:
```makefile
itest: dockerbin
	docker-compose -f integration_tests/docker-compose.yml down -v
	docker-compose -f integration_tests/docker-compose.yml up -d --build
	docker run -it --rm --name stitchit_itest --network=integration_tests_default \
           -e GOPATH=/root/go -v "$(GOPATH):/root/go" -v "$(PROJECT_BASE_PATH):/root/workingdir" \
           -w "/root/workingdir" docker-go-sdk:0.42.0-alpine go test -v ./integration_tests/... -tags integration
```

### Summary

Using the presented approach we 
* Test close to production environment (containers in docker network)
* Use slim, well-known technology stack (docker-compose, golang and `go tools`)
* Easily define, start and tear down the test environment, including network and volumes
* Write integration tests in the same way we write unit tests

# Go routines
Go routines can be described as _light-weight threads_. Their scheduling is managed by the go runtime, which utilizes threads based on demand. It is entirely possible to have hundreds
of go routines multiplexed on a single thread. This is why they are called light-weight. It is entirely feasible running 
thousands or even hundreds of thousands of go routines in your program.
Go routines have a 4kb memory overhead and are very comfortable to use. Often they are employed in conjunction with go's other signature feature, [Go channels](#go-channels)
to implement asynchronous or parallel logic.

A go routine is started with the keyword `go`:

```go
//Anonymous go routine
go func() {
	//do stuff async
}()

func doit() {
	
}
//Running an existing function in a go routine
go doIt()
```

#### Variable scope

Go has closures. So passing variables into go routines that are declared inline, is very convenient:

```go
h := "hello world"
go func() {
	fmt.Println(h)
}()
```
The alternative is to using a closure is using the function parameter:
```go
h := "hello world"
go func(s string) {
	fmt.Println(s)
}(h)
```
Inline declared go routines making use of closures is very common, but must be avoided in loops due to how closures work:

##### Pitfall: Using a closure in a `for` loop
The following test iterates over 3 `Customer`s and prints their name in a go routine.
```go
type Customer struct {
	Name string
}

func TestGoRoutineClosurePitfall(t *testing.T) {
    customers := []Customer{
        {Name: "Avid"},
        {Name: "Olav"},
        {Name: "Jarl Varg"},
    }

	wg := sync.WaitGroup{}
	wg.Add(3)
	for _, customer := range customers {
		go func() {
			fmt.Println(customer.Name)
			wg.Done()
		}()
	}
	wg.Wait()
}
```

Output:
```go
Jarl Varg
Jarl Varg
Jarl Varg
```
This behavior is due to the go routine closure using a reference to `customer`, which gets a new value assigned on every loop iteration. Therefore, all go routines print `Jarl Varg`.
when they are executed.

Fix
```go
func TestGoRoutineClosurePitfall_Fixed(t *testing.T) {
	customers := []Customer{
		{Name: "Avid"},
		{Name: "Olav"},
		{Name: "Jarl Varg"},
	}
	
	wg := sync.WaitGroup{}
	wg.Add(3)
	
	for _, customer := range customers {
		go func(c Customer) {
			fmt.Println(c.Name)
			wg.Done()
		}(customer)
	}
	
	wg.Wait()
}

```

# Go channels

Go channels are golang's signature feature. They are a very powerful tool, but come with some caveats as well.
As a newcomer it is often tricky to figure out when to effectively use go channels. This paragraph lists some of the most useful patterns,
but first we start with some general advice and potential pitfalls:

## Buffered versus unbuffered channel
Buffered channels are created with a specific buffer size. Writes to the channel will not block, until the buffer is full
```go
buffered := make(chan string, 2)
buffered <- "a"
buffered <- "b"
buffered <- "c" //This blocks until another go routine consumes "a" from the channel
```
Buffered channels are *almost never* used. The reason being that whatever buffer size you choose, it will eventually we reached and cause a block.

So in 99.999% of cases you will use unbuffered channels. 

When using an unbuffered channel, writes to the channel block unless at least one other go routine consumes the channel. Conversely, any statement
receiving from a channel, blocks until there is something to receive. This is why unbuffered go channels are a synchronization mechanism.

```go
unbuffered := make(chan string)
go func() {
	read <- unbuffered //This will block until there is something to read
}()
unbuffered <- "hello"
```

## Rules around channels
Multiple go routines can write to a channel and multiple go routines can consume a channel. 

When multiple routines consume the same channel, each message is guaranteed to be only received **by one routine**.

Channels can be closed. Its purpose is to serve as a signal to channel consumers, that processing is
done and receiving from the channel can stop. This is not required for garbage collection, an unclosed channel that is no 
longer referenced will be garbage collected.

Writes to a closed channel or attempting to close it again causes a panic and must [be avoided](#general-channel-best-practices). Consuming a closed channel
is generally okay as the statement will not block, but may lead to [infinite loops](#pitfall-when-closing-channels-when-using-select-and-for).

## General channel best practices
It is good practice having a generator function that creates and returns the channel. The channel is asynchronously populated and closed when done.

```go
func GenerateRandomNumbers(amount int) chan int {
	output := make(chan int) //Create the channel
	//Populate the channel in a go routine, this happens async, so the returned channel is ready to be consumed elsewhere while it is not populated
	go func() {
		for i := 0; i < amount; i++ {
			output <- rand.Int()
		}
		//Once we are done populating the channel, we close it, this will cause consumer loops to exit gracefully
		close(output)
	}()
	return output
}
```
Using the function:
```go
func TestRoutine(t *testing.T) {
	c := go_routines.GenerateRandomNumbers(10)
	for n := range c {
		fmt.Println(n)
	}
	fmt.Println("Done")
}
```

The advantage of this pattern is that the **entire channel lifecycle** is defined in one place. The caller of `GenerateRandomNumbers`
only has to worry about receiving from the channel and can rely on it being closed when no more data is sent.

Remember

* Writing to a closed channel causes a `panic`
* Closing a closed channel causes a `panic`
* Reading from a closed channel evaluates to `true` with empty value, which can cause infinite loops frying your CPU
* Spreading your channel lifecycle throughout the code might lead to death threats from the next maintainer

## Writing to a channel

To write to a channel simply use the `<-` operator
```go
c := make(chan int, 1)
c <- 5 
```
:bulb: In this example, the channel write does not block, although we do not have a consumer, because it is buffered of size `1`.

#### Optional write
In some cases, you want to write to an unbuffered channel, but you don't care if there is a consumer on the other side, and you don't want
the operation to block. In this case, you want to simply discard the element instead. You can achieve this with a `select` and a _default_ case. In the following example we write to
`c`, but there is no consumer and `c` is not a buffered channel. Without the `select`, the `<-` statement would block, but now it simply executes the _default_ case.
Once we have a consumer, the element is written to the channel.

Example: Write or discard
```go
func TestOptionalWrite(t *testing.T) {
    c := make(chan int)
    select {
        case c <- 5:
			t.Fatal("should not have been executed")
        default:
            fmt.Println("Discarded message")
    }
    //Start consumer routine
    go func() {
        <-c //Read an element
    }()
    
    time.Sleep(time.Millisecond) //Make sure consumer routine is up
    
    select {
        case c <- 5:
            fmt.Println("Sent message")
        default:
            t.Fatal("should not have been executed")
    }
}
```

This is useful in cases where your package API exposes a channel but doesn't know whether the user is consuming it.

## Consuming a channel

There are different ways to consume a channel
* Direct assignment
* Loop over a channel with the `range` keyword
* Using `select` to read from multiple channels

### Direct assignment
You can directly assign the next element of the channel to a variable using the `<-` operator
```go
c := go_routines.GenerateRandomNumbers(10)
firstNumber <- c
secondNumber <- c
```

### Range loop
Most typically you will range over a channel
```go
c := go_routines.GenerateRandomNumbers(10)
for n := range c {
    fmt.Println(n)
}
```
:bulb: When using `range` to iterate over a channel, the loop is exited when the channel is closed.

### Select

Select allows you to consume from multiple channels at the same time. It is commonly used for
* Cancellation
* Timeouts
* General multi-channel consumption

#### Cancellation
One can use a _signal_ channel in conjunction with a `select`, to indicate the end of input processing. The following example has three go routines, one to populate the inout channel, 
one to send a cancellation after 5 seconds and one to consume from the input and cancellation channel using `select`:
```go
func TestCancellation(t *testing.T) {
    //Create a channel to signal the end of processing
    cancel := make(chan struct{})
    
    //Write some random numbers
    c := make(chan int)
    go func() {
        for {
            c <- rand.Int()
        }
    }()
    //This routine with wait for 5 seconds, then send a `struct{}{}` into the cancel channel
    go func() {
        <-time.After(time.Second * 5)
        cancel <- struct{}{}
    }()
    
    READ:
        for {
            select {
                case number := <-c:
                    fmt.Println(number)
                case <-cancel:
                    //After ~5 seconds, we will receive on the cancel channel and break out of the loop
                    break READ
            }
        }
    fmt.Println("Done")
}

```

While perfectly sufficient for simple cases, use of the `context` package is preferred. 
It allows simultaneous cancellation of multiple go routines and supports other more complex use 
cases, like child contexts and deadlines.

```go
func TestCancellationWithContext(t *testing.T) {
	c := make(chan int)
	//Using a context with cancel() function instead of a signal channel
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			c <- rand.Int()
		}
	}()

	go func() {
		<-time.After(time.Second * 5)
		cancel() //Invoke cancel() after 5 seconds
	}()

READ:
	for {
		select {
		case number := <-c:
			fmt.Println(number)
		case <-ctx.Done():
			break READ
		}
	}
	fmt.Println("Done")
}
```

#### Timeouts

Another common use of the `select` keyword is timeouts. Imagine a routine that reads from a kafka topic channel, closing
if no message was received after 5 seconds:

```go
func TestRefreshingTimeout(t *testing.T) {
	c := make(chan int)
	go func() {
		//Send only two numbers, then wait for the timeout
		c <- rand.Int()
		c <- rand.Int()
	}()

READ:
	for {
		select {
		case number := <-c:
			fmt.Println(number)
		case <-time.After(time.Second * 5):
			//This case is only executed, if we do not receive anything on 'c' for more than 5 seconds
			//because every loop iteration reinitialized the timer
			break READ
		}
	}
	fmt.Println("Done")
}
```

In order to have a fix timeout, you have to instantiate the timeout channel outside the loop:
```go
func TestFixedTimeout(t *testing.T) {
	c := make(chan int)
	timeout := time.After(time.Second * 5)
	go func() {
		for {
			c <- rand.Int()
		}
	}()

READ:
	for {
		select {
		case number := <-c:
			fmt.Println(number)
		case <-timeout: //This case is executed after 5 seconds, no matter what
			break READ
		}
	}
	fmt.Println("Done")
}
```
#### Pitfall when closing channels when using select and for

When you close a channel, you can still consume from it. Writing to a closed channel or closing it again causes a panic, but reading will
return immediately. This can cause very tragic outcomes when a channel is consumed in a for loop without using `range`, for example in conjunction with `select`.
```go
func TestClosingChannelPitfall(t *testing.T) {
	c := make(chan int)
	//Create a wait group so the test doesn't exit before the go routine is done
	wg := sync.WaitGroup{}
	wg.Add(1)
	
	go func() {
		for {
			select {
			case value := <-c:
				//Once the channel is closed, we will execute this case in an infinite loop with the default int value of `0`
				fmt.Println(value)
			}
		}
		wg.Done()
	}()
	close(c) //Closing the channel will *NOT* exit the for loop in this case
	wg.Wait()
}
```

To fix this you need to explicitly exit the loop
```go
func TestClosingChannelPitfallFixed(t *testing.T) {
	c := make(chan int)
	//Create a wait group so the test doesn't exit before the go routine is done
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
	LOOP: //Label for where to break out of
		for {
			select {
			case value, ok := <-c:
				if !ok {
					break LOOP //break out of loop if channel is closed
				}
				fmt.Println(value)
			}
		}
		wg.Done()
	}()
	close(c) //Closing the channel will *NOT* exit the for loop in this case
	wg.Wait()
}
```

## Patterns

The go blog post [Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines) is an absolute **must-read!**

It discusses 
* Pipeline
* fan-in and fan-out 
* cancellation techniques

Once these patterns are understood, usage of channels and routines feels a lot more natural and intuitive. Identifying the correct context to employ these features will be easier.

#### Pipeline
Is a way to route the data through a series of channels, processing the data in between.

#### Fan out
Can be thought of as the classical worker pattern: A single input channel feeds multiple go routines that each do the same work in parallel. Each element on the channel 
is received by one routine only, making for fairly even distribution across all routines consuming it.

#### Fan In
Fan-In is the consolidation of multiple input channels into a single output channel. Each input channel is consumed by its own go routine. All go routines are writing to the 
same output channel.

#### Channels and APIs
Be careful when exposing channels in your package APIs. It can be useful to provide an asynchronous interface to your users to consume data from, such as errors or events. 
But you must consider the channel lifecycle carefully. Remember that closing or writing to an already closed channel causes a panic.

If you want to allow the package API user to input data via a channel, do not create and return the channel to the user in your API. Instead, allow the user to pass in a channel.
Now you are the consumer of the input channel while the user is in charge of its lifecycle.

In you output data to the API user via a channel document the behavior (blocking, discarding) and lifecycle of your exposed channels well. The consumer should not be responsible
for the channel's lifecycle at all. In cases where consumption of the channel is optional, use the `select-default` pattern to write.

# Synchronization techniques

The most commonly used synchronization techniques are

* `sync` package: Mutex, RWMutex, WaitGroup and Once
* Channels

The `sync` package offers a `Pool` a thread-safe `Map` implementations as well as a `Condition`. Pool is used for big objects that are temporarily un-used, to reduce impact
on garbage collection. a thread-safe map is better implemented typed with a `sync.RWMutex`, but this may change have changed with generic??!? Condition is a more fine-grained WaitGroup, check it out.

### Mutex
In most cases the `sync.RWMutex` is your go-to tool for anything that needs thread-safe **access**. Before generics, a common use case
was creation of thread-safe collections, like in this example:
```go
package synchronization

import "sync"

type ThreadSafeMap struct {
	sync.RWMutex
	m map[string]string
}

func (t *ThreadSafeMap) Add(key, value string) {
	t.Lock()
	defer t.Unlock()
	t.m[key] = value
}
func (t *ThreadSafeMap) Remove(key string) {
	t.Lock()
	defer t.Unlock()
	delete(t.m, key)
}

func (t *ThreadSafeMap) Get(key string) string {
	t.RLock()
	defer t.RUnlock()
	return t.m[key]
}
```
Always make sure your lock is released properly, there are few things worse to debug than locking issues. Using `defer` is a surefire way to do so. 

### WaitGroup
The `sync.WaitGroup` is a great tool to synchronize **go routines**:
```go
func TestWaitGroup(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2) //Set counter to 2
	go func() {
		fmt.Println("Do something")
		<-time.After(time.Second)
		wg.Done() //Decreases counter
	}()
	go func() {
		fmt.Println("Do something else")
		<-time.After(time.Second * 2)
		wg.Done() //Decreases counter
	}()
	wg.Wait() //Wait for counter to become 0
	fmt.Println("All done")
}

```
Without the `WaitGroup`, this test would print "All done" and likely exit before both go routine are executed.

The typical use case for `WaitGroup`s are situations where you distribute work across multiple routines for parallel processing, and need wait for all of them to finish before proceeding with the main thread. As such it can be used to implement a simple fan-out/fan-in.


### Once

`sync.Once` is used whenever you have a piece of code, that you want to be executed only **once**, thread-safe-once. This use case is rare, but it has its use cases.

```go
func TestOnce(t *testing.T) {
	once := sync.Once{}
	var myFunc = func() {
		once.Do(func() {
			fmt.Println("See it only once")
		})
		fmt.Println("See it twice")
	}
	myFunc() // Only this call will execute the code in once
	myFunc()
}
```
Prints: 
```go
=== RUN   TestOnce
See it only once
See it twice
See it twice
--- PASS: TestOnce (0.00s)
PASS
```
The typical use case for `Once` are situations where _lazy initialization_ of a _global_ resource, like a `http.Client` instance for example, is met by potentially multiple routines hitting the initialization code at the same time.

#### Channels

Unbuffered `chan` can be used to synchronize go routines. It is not their primary intend, but is used in examples already discussed, like the following

```go
func TestCancellation(t *testing.T) {
    //Create a channel to signal the end of processing
    cancel := make(chan struct{})
    
    //Write some random numbers
    c := make(chan int)
    go func() {
        for {
            c <- rand.Int()
        }
    }()
    //This routine with wait for 5 seconds, then send a `struct{}{}` into the cancel channel
    go func() {
        <-time.After(time.Second * 5)
        cancel <- struct{}{}
    }()
    
    READ:
        for {
            select {
                case number := <-c:
                    fmt.Println(number)
                case <-cancel:
                    //After ~5 seconds, we will receive on the cancel channel and break out of the loop
                    break READ
            }
        }
    fmt.Println("Done")
}

```

The typical use case for using channels to synchronize go routines, is when you want to actively _signal_ something to the target routine. Either by sending something through the channel, or by closing it.
:bulb: Channels are also commonly used for timeout related synchronization, due to the convenience of `<- time.After(duration)` channels.