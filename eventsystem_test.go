package goarchaius

import (
	"fmt"
	"testing"
)

func onTest(event *Event) {
	m := map[string]string{"test1": "zqtest1"}
	ch, _ := event.Params["ch"].(chan map[string]string)
	ch <- m
}

func onTest2(event *Event) {
	m := map[string]string{"test2": "zqtest2"}
	ch, _ := event.Params["ch"].(chan map[string]string)
	ch <- m
}

func Test_EventLoop(t *testing.T) {
	d := DefaultDispatcher()
	var fun1 EventCallback = onTest
	var fun2 EventCallback = onTest2
	d.AddEventListener("test.*", &fun1)
	d.AddEventListener("test.*", &fun2)

	params := make(map[string]interface{})
	params["id"] = 1000
	params["ch"] = make(chan map[string]string)

	event := CreateEvent("test.a", params)
	fmt.Println("testtest")
	d.DispatchEvent(event)

	i := 0
	for {
		tt, _ := <-params["ch"].(chan map[string]string)
		if tt["test1"] == "zqtest1" {
			i++
		} else if tt["test2"] == "zqtest2" {
			i++
		}
		if i == 2 {
			break
		}
	}
}
