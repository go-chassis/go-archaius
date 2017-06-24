package core

import (
	"testing"
)

func onTest(event *Event) {
	m := "zqtest1"
	ch, _ := event.Value.(chan string)
	ch <- m
}

func onTest2(event *Event) {
	m := "zqtest2"
	ch, _ := event.Value.(chan string)
	ch <- m
}

func Test_EventLoop(t *testing.T) {
	d := DefaultDispatcher()
	var fun1 EventCallback = onTest
	var fun2 EventCallback = onTest2
	d.AddEventListener("test.*", &fun1)
	d.AddEventListener("test.*", &fun2)

	vv := make(chan string)

	event := CreateEvent("testingSource", "test.a", "NEW", vv)
	d.DispatchEvent(event)

	i := 0
	for {
		tt, _ := <-vv
		if tt == "zqtest1" {
			i++
		} else if tt == "zqtest2" {
			i++
		}
		if i == 2 {
			break
		}
	}
}
