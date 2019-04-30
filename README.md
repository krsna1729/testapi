# testapi

The goal of this framework is to easily model and measure the impact of workloads, frameworks, system level load and network loads on the responsiveness of a group of interacting services.

> Note: The current implementation only supports a simple linear chain. However the goal is to support any directed acyclic graph.

The framework attempts to use available latency isolation features like tasksets and Kubernetes guaranteed QoS Class to get more deterministic behavior.

# How to run

The interacting microservices can be launched as processes, containers or deployments in Kubernetes. They can also be run directly in custom manner across a cluster of nodes with a service chain of any desired length and topology by setting appropriate environment variables. The framework provides multiple built in workloads. It also includes leveraging an external workload generator [stress-ng](http://manpages.ubuntu.com/manpages/xenial/man1/stress-ng.1.html) to generate configurable synthetic load in response to a service request. This allows the services to model common workload profiles.

## Environment variables

| Env Variable | Description | 
| --- | --- |
| SERVICE_NAME | Name of the service as it will appear in traces and metrics |
| UPSTREAM_URI | Base URI which the service exposes |
| DOWNSTREAM_URI | Base URI to which the service makes downstream requests to. If not set, this is the last service in the chain (i.e. terminating service) |
| REPORTER_URI | URI of the zipkin trace collector |
| PRIME_MAX | The upper bound for the prime search when using the built in prime generator workload |
| GOGC=off | To disable the golang garbage collector to reduce any variability induced by the golang runtime |
| JOBFILE | stress-ng [job profile](https://github.com/ColinIanKing/stress-ng/tree/master/example-jobs) to simulate workload. This profile will be run when the service sees a request. The content of this file can be changed at runtime and the services will pick up the updated job profile |
| LOADFILE | stress-ng job profile to simulate load. This profile will be run in parallel to the requests. This allows the user to simulate load on the system and its impact on the request |

These environment variables can be used to stitch processes or containers across machines.

The sample deployments below use a service chain of length 3, root->branch->leaf as an illustration. 

## Exposed URIs

The services expose multiple URI's each of which performs different types of work. The current set includes
- `/` : This does no work. It only forwards the request down the chain and allows us to determine the cost of Networking and HTTP handling
- `/busywork`: Performs 10ms of CPU busywork prior to forwarding the request down the chain. This helps baseline the accuracy of the scheduler.
- `/prime`: Computes all primes from `0` to `PRIME_MAX` using the [go implementation of Segmented Sieve](https://github.com/kavehmz/prime). On completion of computation it forwards the request down the chain This should be constant time CPU and Memory intensive computation. However in the real world this has not proven to be quite true.
- `/fork`: Forks a child `date` process. This baselines the cost of forking.
- `/stress-ng`: Forks a `stress-ng` process with the `JOBFILE` as the input profile. This can be used to model most workloads. `stress-ng` itself launches multiple processes to spin up each type of work. Hence the additional fork from golang does not really impact the latency variation.
- `/load`: Forks a long running `stress-ng` process with the `LOADFILE` as the input profile. This launches this workload across the service chain and returns. The load itself continues to run for as long as it is configured to to run. The request returns immediately. This URI is used to trigger a constant long running work that run in parallel to the actual workload under test.
- `/load-status`: Waits for the long running `load` to terminate. This request blocks till the load triggered by `load` terminates across the cluster.


### stress-ng profile

The job profile should ideally be defined in terms of total operations that need to performed (vs time) in the case of the JOBFILE. This ensures that the exact same amount of work is done in response to each request. The default profile used is as follows

```
    metrics-brief
    cpu 1
    cpu-ops 1
    vm 1
    vm-ops 1
    matrix 1
    matrix-ops 1
    crypt 1
    crypt-ops 1
    af-alg 1
    af-alg-ops 1
```

The simple job profile results in about a 0.05s of latency on an unloaded system.

```
stress-ng --job stress.cfg
stress-ng: info:  [11806] defaulting to a 86400 second (1 day, 0.00 secs) run per stressor
stress-ng: info:  [11806] dispatching hogs: 1 cpu, 1 vm, 1 matrix, 1 crypt, 1 af-alg
stress-ng: info:  [11806] successful run completed in 0.05s
stress-ng: info:  [11806] stressor       bogo ops real time  usr time  sys time   bogo ops/s   bogo ops/s
stress-ng: info:  [11806]                           (secs)    (secs)    (secs)   (real time) (usr+sys time)
stress-ng: info:  [11806] cpu                   1      0.05      0.05      0.00        18.61        20.00
stress-ng: info:  [11806] vm                    1      0.00      0.00      0.00      1164.11         0.00
stress-ng: info:  [11806] matrix                1      0.00      0.00      0.00       970.01         0.00
stress-ng: info:  [11806] crypt                 1      0.01      0.01      0.00        68.28       100.00
stress-ng: info:  [11806] af-alg                3      0.01      0.00      0.00       354.90         0.00
```

In the case of a the LOADFILE it should be setup to run for a time duration longer than the test run of the actual responsiveness measurement. The job described below runs for 20s, which is longer than the typical 10s run we use for testing responsiveness.

```
    metrics-brief 
    timeout 20s 
    cpu 3 
    vm 3
    vm-bytes 256M 
    matrix 3 
    crypt 3 
    af-alg 3
```


## Running as processes/containers directly on your host

### Running as processes or containers on single system

The framework allows you to run the workload directly without kubernetes. The set of scripts under `./http/server` helps you understand how this can be achieved.

`./http/server/run_all.sh` will run the full gamut of scenarios
- processes interacting with one another
- containers interacting with one another (without container network isolation)
- containers interacting with one another over a container network

This helps you baseline the behavior of the system on that particular hardware and operating system.
The [hey](https://github.com/rakyll/hey) reported metrics are available at `./http/server/latency_all.log`

For example on a particular system under test

**Baseline http performance**
Shows a very tight 2 ms response range
```
hello:broot:/:hello:bbranch:/:hello:bleaf:/
Response time histogram:
  0.000 [1]     |
  0.001 [427]   |■■■■
  0.001 [406]   |■■■■
  0.001 [381]   |■■■■
  0.001 [1311]  |■■■■■■■■■■■■■■
  0.002 [3815]  |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.002 [1045]  |■■■■■■■■■■■
  0.002 [5]     |
  0.002 [0]     |
  0.002 [1]     |
  0.003 [1]     |
```


**Baseline scheduler**
`10ms * 3 services + baseline http = 31 to 33 ms` expected latency.
```
hello:broot:/busywork:hello:bbranch:/busywork:hello:bleaf:/busywork
Response time histogram:
  0.031 [1]     |
  0.032 [8]     |■■
  0.032 [16]    |■■■■
  0.032 [11]    |■■■
  0.032 [80]    |■■■■■■■■■■■■■■■■■■■
  0.032 [171]   |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.032 [22]    |■■■■■
  0.033 [3]     |■
  0.033 [0]     |
  0.033 [0]     |
  0.033 [1]     |
```

**Baseline prime computation**
Shows a range of 15-20 ms range. As the exact same prime computation is done which takes about 5ms. Hence the expected range is `5ms * 3 services + baseline http = 16-18ms`. There is an additional `2ms` variation which may be due to system calls and memory allocations.


```
hello:broot:/prime:hello:bbranch:/prime:hello:bleaf:/prime
Response time histogram:
  0.015 [1]     |
  0.015 [13]    |■■■■
  0.016 [81]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.016 [99]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.017 [116]   |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.017 [103]   |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.018 [79]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.018 [57]    |■■■■■■■■■■■■■■■■■■■■
  0.019 [22]    |■■■■■■■■
  0.019 [12]    |■■■■
  0.020 [7]     |■■
```

**Running as containers**

The expected range based on the previous measurement was `16-18ms`, but we see `18-29ms`

```
hello:lroot:/prime:hello:lbranch:/prime:hello:lleaf:/prime
Response time histogram:
  0.018 [1]     |■
  0.019 [18]    |■■■■■■■■■
  0.020 [21]    |■■■■■■■■■■■
  0.021 [55]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.022 [79]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.023 [66]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.024 [75]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.025 [67]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.026 [33]    |■■■■■■■■■■■■■■■■■
  0.027 [16]    |■■■■■■■■
  0.029 [7]     |■■■■
```
**Running as containers with networking**
Shows a the same spread as the previous case. This means the docker networking setup did not add additional variation. So the network namespace, docker bridge and veth does not seem to add significant variations.

```
hello:croot:/prime:hello:cbranch:/prime:hello:cleaf:/prime
Response time histogram:
  0.018 [1]     |■
  0.019 [24]    |■■■■■■■■■■■■■
  0.020 [30]    |■■■■■■■■■■■■■■■■
  0.021 [58]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.022 [63]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.023 [68]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.024 [75]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.025 [50]    |■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.026 [43]    |■■■■■■■■■■■■■■■■■■■■■■■
  0.028 [15]    |■■■■■■■■
  0.029 [8]     |■■■■
  ```

Also a point to note would be that the baseline scheduler still performs as expected. So the variation is coming from the actual work that is being performed.
```
hello:croot:/busywork:hello:cbranch:/busywork:hello:cleaf:/busywork
Response time histogram:
  0.032 [1]     |
  0.032 [53]    |■■■■■■■■■■■■■
  0.032 [160]   |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.032 [87]    |■■■■■■■■■■■■■■■■■■■■■■
  0.033 [6]     |■■
  0.033 [2]     |■
  0.033 [0]     |
  0.033 [0]     |
  0.033 [0]     |
  0.033 [0]     |
  0.034 [1]     |
```

Now that the baseline has been established this workload can be run in kubernetes to evaluate the impact of different types of networking plugins, service mesh technologies and service meshes to the latency spread. This can also be used to model the impact of Kubernetes QoS classes, replicas and affinity and hardware topology.

**Kubernetes baseline**

The sample output from a single node Kubernetes cluster is shown below. 

Here we see that the baseline http now is typically `1ms` but ranges from `1-7ms` vs `1-2ms`. We also see that the scheduler latency is typically `33ms` with some outliers.

```
hello:root:/:hello:branch:/:hello:leaf:/
Response time histogram:
  0.001 [1]     |
  0.001 [8873]  |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.002 [3]     |
  0.003 [3]     |
  0.003 [3]     |
  0.004 [1]     |
  0.004 [1]     |
  0.005 [2]     |
  0.006 [0]     |
  0.006 [0]     |
  0.007 [1]     |


hello:root:/busywork:hello:branch:/busywork:hello:leaf:/busywork
Response time histogram:
  0.031 [1]     |
  0.033 [312]   |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.036 [1]     |
  0.038 [0]     |
  0.040 [0]     |
  0.042 [0]     |
  0.044 [0]     |
  0.047 [0]     |
  0.049 [0]     |
  0.051 [0]     |
  0.053 [1]     |
```

**Kubernetes prime computation**

Here we see that the prime computation latency now ranges from `15-20ms` with no real outliers. So this is within 2ms of the expected range. 

```
hello:root:/prime:hello:branch:/prime:hello:leaf:/prime
Response time histogram:
  0.015 [1]     |
  0.015 [5]     |■■
  0.016 [7]     |■■
  0.016 [36]    |■■■■■■■■■■■
  0.017 [78]    |■■■■■■■■■■■■■■■■■■■■■■■■
  0.017 [100]   |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.018 [130]   |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.018 [120]   |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.019 [67]    |■■■■■■■■■■■■■■■■■■■■■
  0.019 [25]    |■■■■■■■■
  0.020 [4]     |■
```

Increasing the concurrency level to 3 however does increase the latency spread due to some outliers.

```
hello:root:/prime:hello:branch:/prime:hello:leaf:/prime
Response time histogram:
  0.007 [1]     |
  0.021 [2526]  |■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.036 [1]     |
  0.051 [0]     |
  0.065 [1]     |
  0.080 [2]     |
  0.095 [3]     |
  0.109 [0]     |
  0.124 [0]     |
  0.139 [0]     |
  0.153 [3]     |
```

### Traces and metrics

Traces and metrics helps you root cause the source of variability. The metrics are available at each service. The traces are available across the whole service chain. The traces have been setup with a probability sampler to reduce the framework overhead. But the traces still help root cause variations as the latency variations are typically visible within the samples.

- The traces are available at http://192.168.211.2:9411
- The metrics are available at http://192.168.211.3:9090
- Grafana is available to http://localhost:3000

You can also query the metrics directly from the individual service.

For example in the case of processes running directly on the host you can query then at
```
curl localhost:8885/metrics
curl localhost:8886/metrics
curl localhost:8887/metrics
```

### Load generation

> Note: Ensure that the SERVER_URI environment variable is setup properly to reflect the test

#### Using hey 
```
go get -u github.com/rakyll/hey

# Reusing the HTTP connection for all requests
hey -c 1 -z 10s $SERVER_URI

# If you want to include HTTP connections setup time (which adds a lot of variation)
hey -c 1 -z 10s -disable-keepalive $SERVER_URI
```

hey will report client visible latency, which can then be broken down using the prometheus exported metrics and zipkin traces. You can change the concurrency via `-c` and the runlength via `-z`.

A sample output of hey in a kubernetes cluster can be seen below. This shows the latency spread. The fastest response times are in line with stress-ng profile and the overhead of traversing the network stack.


#### Alternately using the builtin client
```
cd ./http/client
go build
while true; do COUNT=1000 ./client ; done
```

> Note: The built in client periodically reports the raw HDR histogram buckets

# Visualizing the latency

## Individual traces

The framework uses the opencensus provided opentracing support to record the runtime traces. These can be reported to a trace collector such as Zipkin and visualized.

![Zipkin](https://github.com/mcastelino/testapi/blob/master/documentation/images/zipkin_trace.GIF)

Here you can see the actual accounting of latency across the cluster.
- Individual busywork computation take 10ms as expected
- End to end latency
- Latency across service hops
- Latency experience by the upstream caller

Note: All this information is also available via openmetrics and can be gathered by prometheus. 

These metrics can also be directly queried from any service directly.

```
$ curl 192.168.211.4:8887/metrics
# HELP croot_opencensus_io_http_client_latency Latency distribution of HTTP requests
# TYPE croot_opencensus_io_http_client_latency histogram
croot_opencensus_io_http_client_latency_bucket{le="0.0"} 0.0
croot_opencensus_io_http_client_latency_bucket{le="1.0"} 33787.0
croot_opencensus_io_http_client_latency_bucket{le="10.0"} 38397.0
...
croot_opencensus_io_http_client_latency_bucket{le="13.0"} 38742.0
croot_opencensus_io_http_client_latency_bucket{le="16.0"} 39511.0
croot_opencensus_io_http_client_latency_bucket{le="20.0"} 40129.0
croot_opencensus_io_http_client_latency_bucket{le="25.0"} 41396.0
...
croot_opencensus_io_http_client_latency_bucket{le="+Inf"} 41402.0
croot_opencensus_io_http_client_latency_sum 78679.04198799946
croot_opencensus_io_http_client_latency_count 41402.0
# HELP croot_opencensus_io_http_server_latency Latency distribution of HTTP requests
# TYPE croot_opencensus_io_http_server_latency histogram
croot_opencensus_io_http_server_latency_bucket{le="0.0"} 0.0
croot_opencensus_io_http_server_latency_bucket{le="1.0"} 32929.0
croot_opencensus_io_http_server_latency_bucket{le="2.0"} 38388.0
...
croot_opencensus_io_http_server_latency_bucket{le="300.0"} 41402.0
croot_opencensus_io_http_server_latency_bucket{le="+Inf"} 41402.0
croot_opencensus_io_http_server_latency_sum 105745.96782699955
croot_opencensus_io_http_server_latency_count 41402.0
# HELP croot_prime_latency The distribution of the latencies for prime calculation
# TYPE croot_prime_latency histogram
croot_prime_latency_bucket{method="busyHandler",le="1.0"} 0.0
croot_prime_latency_bucket{method="busyHandler",le="11.0"} 1241.0
...
croot_prime_latency_bucket{method="busyHandler",le="+Inf"} 1245.0
croot_prime_latency_sum{method="busyHandler"} 12621.167159000006
croot_prime_latency_count{method="busyHandler"} 1245.0
croot_prime_latency_bucket{method="primeHandler",le="1.0"} 0.0
croot_prime_latency_bucket{method="primeHandler",le="8.0"} 1427.0
croot_prime_latency_bucket{method="primeHandler",le="9.0"} 1644.0
croot_prime_latency_bucket{method="primeHandler",le="10.0"} 1757.0
...
croot_prime_latency_bucket{method="primeHandler",le="+Inf"} 1760.0
croot_prime_latency_sum{method="primeHandler"} 12041.534547999996
croot_prime_latency_count{method="primeHandler"} 1760.0
```



## Higher Level Metrics

Latency of the response at root microservice can be visualized using histograms with the following formulas either using Grafana or Prometheus

![Grafana](https://github.com/mcastelino/testapi/blob/master/documentation/images/grafana_trace.GIF)


```
 histogram_quantile(0.99, sum(rate(root_opencensus_io_http_server_latency_bucket[1m])) by (le))
 histogram_quantile(0.95, sum(rate(root_opencensus_io_http_server_latency_bucket[1m])) by (le))
 histogram_quantile(0.90, sum(rate(root_opencensus_io_http_server_latency_bucket[1m])) by (le))
 histogram_quantile(0.50, sum(rate(root_opencensus_io_http_server_latency_bucket[1m])) by (le))
```

The contribution of the downstream services to this is captured by `root_opencensus_io_http_client_latency_bucket` 

To see latencies of the downstream services use the appropriate service names `branch_opencensus_io_http_server_latency_bucket` or `leaf_opencensus_io_http_server_latency_bucket`.

To see latencies experienced by the downstream services use the appropriate service names `branch_opencensus_io_http_client_latency_bucket`.

# Using Kubernetes

We are using [k3s](https://github.com/rancher/k3s/blob/master/README.md) to launch a lightweight Kubernetes cluster. This allows you to seamlessly deploy model latency across a cluster of machines without adding significant overhead. This also allows you to model noisy neighbour or adding more load to the microservice itself by adding load containers to the POD. This also allows you to constrain the workload's CPU and Memory profile.

> Note: You can use any kubernetes cluster, even an existing working cluster to deploy the service chain.

### Create a cluster

> Note: This step is needed if you do not have access to a set of nodes (virtual or physical)

Launch a cluster of VM's using vagrant.

The following command will create a 3 VM cluster

```
vagrant up
```

The remaining instructions assume that you are using the vagrant cluster. But these steps apply to any other setup where vagrant ssh is replaced by ssh to the appropriate node.

## Bootstrap the master node

```
vagrant ssh ubuntu-01
curl -sfL https://get.k3s.io | sh -
```

Get the Master's primary IP address and kubernetes token
```
MASTER=`hostname -I | cut -d' ' -f1`
NODE_TOKEN=`sudo cat /var/lib/rancher/k3s/server/node-token`
```

## Join the remaining nodes to the cluster

Assuming MASTER and NODE_TOKEN have been set on each node based on the values obtained from master.

```
vagrant ssh ubuntu-02
curl -sfL https://get.k3s.io | K3S_URL=https://$MASTER:6443 K3S_TOKEN=$NODE_TOKEN sh -
```

```
vagrant ssh ubuntu-03
curl -sfL https://get.k3s.io | K3S_URL=https://$MASTER:6443 K3S_TOKEN=$NODE_TOKEN sh -
```

## Check that the nodes are up

```
vagrant ssh ubuntu-01

vagrant@ubuntu-01:~$ kubectl get nodes
NAME        STATUS   ROLES    AGE    VERSION
ubuntu-01   Ready    <none>   15m    v1.13.4-k3s.1
ubuntu-02   Ready    <none>   117s   v1.13.4-k3s.1
ubuntu-03   Ready    <none>   29s    v1.13.4-k3s.1

```

## Deploy the prometheus operator

```
kubectl apply -f testapi/k8s/manifests/phase_2/
```

Verify that all the pods come up

## Deploy prometheus and zipkin 

```
kubectl apply -f testapi/k8s/manifests/phase_3/
```

Verify that all the pods come up

## Deploy the service chain

```
kubectl apply -f testapi/k8s/manifests/service_chain/
```

### Obtain the IPs of the root service, Prometheus and Zipkin

```
vagrant@ubuntu-01:~$ kubectl get svc | grep 'root\|testapi-prom\|zipkin'
root           ClusterIP   10.43.3.206     <none>        8888/TCP,8887/TCP   48m
testapi-prom   ClusterIP   10.43.237.117   <none>        9090/TCP            49m
zipkin         ClusterIP   10.43.99.56     <none>        9411/TCP            49m
```

```
export SERVER_URI=http://10.43.3.206:888
```

Now you can run the load and obtain metrics as explained before.

#### Reaching the services from outside

You can access prometheus from your host using ssh port forwarding.
Assuming that the vagrant VM has IP of `192.168.121.12`

```
ssh -NL 9090:10.43.237.117:9090 vagrant@192.168.121.12 -i ~/testapi/.vagrant/machines/ubuntu-01/libvirt/private_key
firefox localhost:9090
```

### Dynamically updating the workload profile

The configMap that contains the workload profile can be updated at runtime and the services will pick up the modified profile for the next request they service. This allows the dynamic reconfigration of the service workload profile at runtime without requiring the containers/processes to restart.

# Results obtained from a real Kubernetes cluster

Some results obtained under various scenarios in a real cluster are described in the [project wiki](https://github.com/mcastelino/testapi/wiki).

https://github.com/mcastelino/testapi/wiki
