package channels_test

import (
	"context"
	"fmt"
	"math/rand"
	"minimalgo/channels"
	"sync"
	"testing"
	"time"
)

func TestCancellation(t *testing.T) {
	//Create a channel to signal the end of processing
	cancel := make(chan struct{})

	//Write so,me random numbers
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

func TestCancellationWithContext(t *testing.T) {
	c := make(chan int)
	//Using a context with cancel instead of a signal channel
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

func TestRoutine(t *testing.T) {
	c := channels.GenerateRandomNumbers(10)
	for n := range c {
		fmt.Println(n)
	}
	fmt.Println("Done")
}

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
