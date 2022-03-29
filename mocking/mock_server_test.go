package mocking_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"minimalgo/mocking"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

func TestDoPOST(t *testing.T) {
	server, err := startMockServer(t, []byte("Hello world"), 200, []byte(``))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer server.Close()

	//Send the request to the mocked endpoint
	err = mocking.DoPOST(fmt.Sprintf("http://localhost:%d", mockPort), "Hello world")
	assert.Nil(t, err)
}
