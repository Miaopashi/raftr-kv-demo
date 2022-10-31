## Raft-KV-Demo

Raft-KV-Demo is a distributed KV system using raft protocol, which is access by memory.

This is a final homework of a [column class]([分布式协议与算法实战_分布式_分布式算法-极客时间](https://time.geekbang.org/column/intro/100046101)). Since I completed this course by pirated material, I'm not sure whether the lecturer gave a better implementation.

## Building

This repository uses [hashicorp-raft]([GitHub - hashicorp/raft: Golang implementation of the Raft consensus protocol](https://github.com/hashicorp/raft)), so you'll need Go version 16+ installed.

There are two port needed in a single node (such as: 8080, 8079, one for HTTP service, the other for raft RPC). they can be set by `-haddr` and `-raddr`.

## Documentiaon

#### Initiate raft node

For the first node to be initiated, you should use `-bootstrap` param. So that the first node will vote itself as the leader of cluster.

```shell
raft-kv-demo -id node1 -raddr cluster-host1:8080 -haddr cluster-host1:8079 -bootstrap
```

For other nodes, we can use `-join`, sending join http request to leader, to make the node join in the cluster. 

```shell
raft-kv-demo -id node2 -raddr cluster-host2:8080 -haddr cluster-host2:8079 -join localhost:8079
```

Of course, onece you initiate the node by `-join`, you **don't** need to use `-join` fot the next initiation.

#### Try KV options by curl

After initiate the cluster, you can try to send http request to every node in the cluster.

```shell
curl -XGET "http://cluster-host1:8079/key?key=raft&level=default"
curl -XGET "http://cluster-host2:8079/key?key=raft&level=stale"
curl -XGET "http://cluster-host2:8079/key?key=raft&level=consistent"

curl -X POST http://cluster-host1:8079/key -d '{"key":"raft","value":"paxos"}'

curl -XDELETE "http://cluster-host1:8079/key/?key=raft"
curl -XDELETE "http://cluster-host2:8079/key/?key=raft"
```

As you can see, fot `GET` request, you can chose different level of consistent of the querying.

In `default` level, you must get KV in (new/old) leader.

In  `stale` level, you can get KV both in follower ans leader.

In `consistent` level, you must get KV in the latest leader of the cluster.

For `POST` and `DELETE` request, you must send it to leader. 

#### Redirection

If the node you send message is not leader, the http service will return you a 307 redirection. So the client will help you send same message to the latest leader.

## Change member of cluster

You can send `join` or `remove` http request to add or delete node in cluster.

```shell
# addr param means RaftBind of the node need to be joined.
curl -XGET "http://cluster-host1:8079/join?id=node2&addr=cluster-host2:8080"

curl -XGET "http://cluster-host1:8079/remove?id=node2"
```
