package channels

import "math/rand"

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
