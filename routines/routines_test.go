package routines_test

import (
	"fmt"
	"sync"
	"testing"
)

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
