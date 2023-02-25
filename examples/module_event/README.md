module_event.go keep a file under archaius's management, and watch module changes,
so that if there are multiple changes under the module **test.person**,
the events will be triggered once, and listener will receive the event list

```
go build module_event.go
./module_event
```

change configs under **test.person** in the module_event.yaml

check the stdout to see events
