apollo.go start a listener from apollo server, and watch config changes.

if you want to run apollo.go, you should do that:
1. in main function, replace `http://127.0.0.1:8000` to your apollo server url.
2. in main function, replace `demo` to your namespace list, for example, `namespace1,namespace2`, join every namespace name with `,`.
3. in main function, replace `demo-apollo` to your app id.

then, run command:

```go
go build apollo.go
./apollo
```

***NOTE***

If you want to see changes, modify config in your apollo web page, and check thw standard out to see events.
