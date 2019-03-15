# testapi

The goal of this framework is to easily model and measure the impact of workloads, frameworks, system level load and network loads on the responsiveness of a group of interacting services.

> Note: The current implementation only supports a simple linear chain. However the goal is to support any directed acyclic graph.

# How to run

The interacting microservices can be launched as processes, containers or on Kubernetes. This can also be run directly in custom manner across a cluster of nodes with a service chain of any desired length and topology by setting the environment variables to the proper values prior to launching each micro service.

## Environment variables

- SERVICE_NAME: Name of the service as it will appear in traces and metrics.
- UPSTREAM_URI: URI which the service exposes
- DOWNSTREAM_URI: URI to which the service makes downstream requests to. If not set, this is the last service in the chain (i.e. terminating service).
- REPORTER_URI: URI of the zipkin trace collector
- CPU_BUSYWORK: Amount of CPU busywork the service performance after it gets the request but before it sends the request downstream

> Note: Other types of work will be added to the model in the future, including disk, network and memory.

These environment variables can be used to stitch processes or containers across machines.

The sample deployments below use a service chain of length 3, root->branch->leaf as an illustration.


## Running as processes/containers directly on your host

### Running as processes
```
export SERVER_URI=http://localhost:8888
cd ./http/server
./run_chain.sh
```

### Running as containers on your host

```
export SERVER_URI=http://192.168.211.4:888
cd ./http/server
./run_container_chain.sh
```

### Traces and metrics
- The traces are available at http://localhost:9411
- The metrics are available at http://localhost:9090
- Grafana is available to http://localhost:3000

### Load generation

> Note: Ensure that the SERVER_URI environment variable is setup properly.

#### Using hey 
```
go get -u github.com/rakyll/hey
while true; do hey -c 1 -z 10s -disable-keepalive $SERVER_URI; done
```

#### Alternately using the builtin client
```
cd ./http/client
go build
while true; do COUNT=1000 ./client ; done
```

> Note: The built in client periodically reports the raw HDR histogram buckets

# Visualizing the latency

Latency of the response at root microservice can be visualized using histograms with the following formulas either using Grafana or Prometheus

```
 histogram_quantile(0.99, sum(rate(root_opencensus_io_http_server_latency_bucket[1m])) by (le))
 histogram_quantile(0.95, sum(rate(root_opencensus_io_http_server_latency_bucket[1m])) by (le))
 histogram_quantile(0.90, sum(rate(root_opencensus_io_http_server_latency_bucket[1m])) by (le))
 histogram_quantile(0.50, sum(rate(root_opencensus_io_http_server_latency_bucket[1m])) by (le))
```

The contribution of the downstream services to this is captured by `root_opencensus_io_http_client_latency_bucket` 

To see latencies of the downstream services use the appropriate service names `branch_opencensus_io_http_server_latency_bucket` or `leaf_opencensus_io_http_server_latency_bucket`.

To see latencies experienced by the downstream services use the appropriate service names
`branch_opencensus_io_http_client_latency_bucket`.

# Using Kubernetes

We are using [k3s](https://github.com/rancher/k3s/blob/master/README.md) to launch a lightweight Kubernetes cluster. This allows you to seamlessly deploy model latency across a cluster of machines without adding signifant overhead. This also allows you to model noisy neighbour or adding more load to the microservice itself by adding load containers to the POD. This also allows you to constrain the workload's CPU and Memory profile.

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

Note: The service chain can simulate any workload by adding
additional workload containers to the pod definition.

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
