# Always run all the framework components so that any additional variability introduced by them
# and the iptables rules introduced by them are always present in all paths

go build
docker rm -f grafana-latency zipkin-latency-net prometheus-latency-net root branch leaf
killall server
docker network rm apinet 

docker build -t mcastelino/test-api-server:latest .

# Docker DNS does not work with Kata (hence use explicit IP)
# Use a custom docker network so that you can control the network
docker network create apinet --subnet=192.168.211.0/24

docker run --cpuset-cpus 0-4 -d --name=grafana-latency --net=host -p 3000:3000 grafana/grafana


# Containers for the docker with network tests
# Docker DNS does not work with Kata (hence use explicit IP)
# Use a custom docker network so that you can control the network
docker run --network apinet --name=zipkin-latency-net -d -p 9411:9411 --ip 192.168.211.2  openzipkin/zipkin

docker run --network apinet --name=prometheus-latency-net -d --ip 192.168.211.3 \
        -v $PWD/prometheus_docker.yml:/etc/prometheus/prometheus.yml \
                                         prom/prometheus
