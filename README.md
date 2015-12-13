# Blued
Blued is a decentralized solution for service discovery building on [Serf](https://github.com/hashicorp/serf). Serf is a node discovery and orchestration tool and is the only tool discussed so far that is built on an eventually-consistent gossip model with no centralized servers. It provides a number of features, including group membership, failure detection, event broadcasts, and a query mechanism. However, Serf does not provide any high-level features such as service discovery, health checking. Blued is a complete system providing all of those features.

The internal gossip protocol used within Blued is in fact powered by the Serf library: Blued leverages the membership and failure detection features and builds upon them to add service discovery. By contrast, the discovery feature of Serf is at a node level, while Blued provides a service and node level abstraction.

The health checking provided by Serf is very low level and only indicates if the agent is alive. Blued extends this to provide  host and service-level checks. Health checks are integrated with a central catalog that operators can easily query to gain insight into the cluster.

Blued is opinionated in its usage while Serf is a more flexible and general purpose tool. In CAP terms, is the same as Serf, Blued is an AP system and sacrifices consistency for availability. 

## Quick Start

#### compile Blued

If you wish to work on Blued itself, you'll first need Go installed (version 1.3+ is required). Make sure you have Go properly installed, including setting up your GOPATH.

Next, clone this repository into $GOPATH/src/github.com/bluefw/blued and then just type ```go build```. In a few moments, you'll have a working blued executable:

#### create cluster

Next, let's start a couple Blued agents. Agents run until they're told to quit and handle the communication of maintenance tasks of Blued. In a real Blued setup, each node in your system will run one or more Blued agents (it can run multiple agents if you're running multiple cluster types. e.g. web servers vs. memcached servers).

Start each Blued agent in a separate terminal session so that we can see the output of each. Start the first agent:
```
$ blued agent -node=foo -bind=127.0.0.1:5000 -rpc-addr=127.0.0.1:7373 -rest-addr=127.0.0.1:8341
...
```

Start the second agent in another terminal session (while the first is still running):
```
$ blued agent -node=bar -bind=127.0.0.1:5001 -rpc-addr=127.0.0.1:7374 -rest-addr=127.0.0.1:8342
...
```

At this point two Blued agents are running independently but are still unaware of each other. Let's now tell the first agent to join an existing cluster (the second agent). When starting a Blued agent, you must join an existing cluster by specifying at least one existing member. After this, Blued gossips and the remainder of the cluster becomes aware of the join. Run the following commands in a third terminal session.
```
$ blued join 127.0.0.1:5001
...
```
If you're watching your terminals, you should see both Blued agents become aware of the join. You can prove it by running blued members to see the members of the Blued cluster:
```
$ blued members
foo    127.0.0.1:5000    alive
bar    127.0.0.1:5001    alive
...
```

At this point, you can ctrl-C or force kill either Blued agent, and they'll update their membership lists appropriately. If you ctrl-C a Blued agent, it will gracefully leave by notifying the cluster of its intent to leave. If you force kill an agent, it will eventually (usually within seconds) be detected by another member of the cluster which will notify the cluster of the node failure.

#### register service
Blued provide restful api for service discovery feature. assumed you have an application service at ```http://127.0.0.1:80/rs``` provide two services ```a.b and a.c``` and the application consumer a service ```x.c```, you can register this via the following code:
```
$ curl -H "Content-Type: application/json" -X PUT -d \
     '{"addr":"http://127.0.0.1:80/rs","providers":["a.b","a.c"],"consumers":["x.c"]}' \
     http://127.0.0.1:8341/msd/register
```

The service registed to Blued will invalid after 60 second (you can alter the time by paramter ```service-ttl``` when start blued agent). So if you want make your service keeping active, you must refresh it within 60 second.
```
$ curl -X GET http://127.0.0.1:8341/msd/refresh/aHR0cDovLzEyNy4wLjAuMTo4MC9ycw==
```
"aHR0cDovLzEyNy4wLjAuMTo4MC9ycw==" is base64 code of "http://127.0.0.1:80/rs"

## API Doc


