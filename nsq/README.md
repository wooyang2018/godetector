#### 本地启动nsq
```
nsqd -data-path E:\nsq\data
```

#### 容器启动nsq
```
docker pull nsqio/nsq
docker run -itd --name nsqd -p 4150:4150 -p 4151:4151 nsqio/nsq /nsqd
```


#### 容器启动etcd
```
docker run  -itd -p 2379:2379  -p 2380:2380 --name etcd quay.io/coreos/etcd:v3.5.2   /usr/local/bin/etcd --name s1 --data-dir /etcd-data --listen-client-urls http://0.0.0.0:2379 --advertise-client-urls http://0.0.0.0:2379 --listen-peer-urls http://0.0.0.0:2380 --initial-advertise-peer-urls http://0.0.0.0:2380 --initial-cluster s1=http://0.0.0.0:2380 --initial-cluster-token tkn --initial-cluster-state new --log-level info --logger zap --log-outputs stderr
```

