package synchronization_test

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

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
