kie.go start a listener from ServiceComb-Kie server, and watch config changes.

if you want to run kie.go, you should do that:
1. start a ServiceComb-Kie server. ([Quick start](https://kie.readthedocs.io/en/latest/getstarted/install.html))

2. (optional) in main function, replace `http://127.0.0.1:30110` to your ServiceComb-Kie server url.

3. run command to create a configuration to kie (if your ServiceComb-Kie server url is `http://127.0.0.1:30110`):

```shell
curl -X POST -H 'Content-Type:application/json' http://127.0.0.1:30110/v1/default/kie/kv -d '{"key": "user","labels":{"appId":"foo","serviceName":"bar","version":"1.0.0","environment":"prod"},"value":"admin","value_type":"text","status":"enabled"}'
```
and you will receive a response like:
```json
{
"id": "b8c27a45-d7fa-413a-8b14-bdc905c4f917",
"label_id": "25cb9189-4233-4af9-95e0-2f008ed33aa6",
"key": "user",
"value": "admin",
"value_type": "text",
"create_revision": 2,
"update_revision": 2,
"create_time": 1585818114,
"update_time": 1585818114,
"labels": {
"appId": "foo",
"environment": "prod",
"serviceName": "bar",
"version": "1.0.0"
}
}
```
record the `id` in the response
```shell
export id=b8c27a45-d7fa-413a-8b14-bdc905c4f917
```

4. run command to start demo:

```shell
go run kie.go
```

5. you can see that the console will print out the value `admin` of the key `user` created in the previous step.

6. run command to update the configuration, the parameter `id` is obtained in the response of creation

```shell
curl -X PUT -H 'Content-Type:application/json' http://127.0.0.1:30110/v1/default/kie/kv/${id} -d '{"value":"guest"}'
```

7. you can see a configuration update events and new values `guest` of the key `user` in the console
