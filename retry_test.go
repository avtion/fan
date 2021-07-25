package main

import (
	"fmt"
	"testing"
	"time"
)

var counter int

func mockClient() error {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case <-timer.C:
	}
	if counter < 4 {
		err := fmt.Errorf("try failed, num: %d", counter)
		counter++
		return err
	}
	return nil
}

func TestExpo(t *testing.T) {
	if err := retry(mockClient,
		retryDefaultInitInterval, retryDefaultMaxTime, retryDefaultDefaultMaxRetry); err != nil {
		t.Fatal(err)
	}
	t.Log("ok")
}
