kie.go start a listener from ServiceComb-Kie server, and watch config changes.

if you want to run kie.go, you should do that:
1. start a ServiceComb-Kie server. ([Quick start](https://kie.readthedocs.io/en/latest/getstarted/install.html))

2. (optional) in main function, replace `http://127.0.0.1:30110` to your ServiceComb-Kie server url.

3. run command to create a configuration to kie (if your ServiceComb-Kie server url is `http://127.0.0.1:30110`):

   ```shell
   curl -X PUT -H 'Content-Type:application/json' http://127.0.0.1:30110/v1/default/kie/kv/user -d '{"labels":{"app":"foo","service":"bar","version":"1.0.0","environment":"prod"},"value":"admin","value_type":"text"}'
   ```

4. run command to start demo:

    ```shell
    go run kie.go
    ```

5. you can see that the console will print out the value `admin` of the key `user` created in the previous step.

6. run command to update the configuration:

   ```shell
   curl -X PUT -H 'Content-Type:application/json' http://127.0.0.1:30110/v1/default/kie/kv/user -d '{"labels":{"app":"foo","service":"bar","version":"1.0.0","environment":"prod"},"value":"guest","value_type":"text"}'
   ```
   
7. you can see a configuration update events and new values `guest` of the key `user` in the console
