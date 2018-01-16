### Go-Archaius 
This is a dynamic configuration client for Go-Chassis which helps in configuration
management for micro-services developed using Go-Chassis sdk. The main objective of
this client is to pull or sync the configuration from config-sources for a particular
micro-service.

### Sources
This Go-Archaius client supports multiple sources for the configuration.
1. Command Line Sources - You can give the configurations key and values in the command lines arguments 
while starting the microservice.

2. Environment Variable Sources - You can specify the sources of conifguration in Environment variable.
3. External Sources - You can also specify the configuration sources to be some 
external config server from where the client can pull the configuration.

4. Files Sources - You can specify some specific files from where client can read 
the configuration for the microservices.

You can also specify multiple sources at a same time. Go-Archaius client keeps all 
the sources marked with their precendence, in case if two sources have same config
then source with higher precendence will be selected.


### Refresh Mechanism
Go-Archaius client support 2 types of refresh mechanism:
1. Web-Socket Based - In this client makes an web socket connection with
the config server and keeps getting an events whenever any data changes.
```
refreshMode: 0
```
2. Pull Configuration - In this type client keeps polling the configuration from
the config server at regular intervals.
```
refreshMode: 1
```

### Configuration for Go-Archaius client.
You can configure the client to pull the configuration from external config server,
 the server ip address and basic parameters can be added in chassis.yaml  
 ```
  config:
    client:
      serverUri: http://XX:YYY   #Config Server IP
      tenantName:  default #This configuration is for local environment, for paas platform there is a auth plugin for authentication. If dont provide the tenant name it will take default values.
      refreshMode: 1  # 配置动态刷新模式，0为configcenter在发生变化时主动推送，1为client端周期拉取，其他值均为非法，不会去连配置中心
      refreshInterval: 30 #refreshMode配置为1时，client端主动从配置中心拉取配置的周期，单位秒
      autodiscovery: false
      api:  #Optional
        version: v3

```

