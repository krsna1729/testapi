go build
docker rm -f zipkin prometheus grafana zipkin-net prometheus-net root branch leaf
killall server

docker run --cpuset-cpus 0-4 --name=zipkin -d -p 9411:9411 --net=host openzipkin/zipkin
docker run --cpuset-cpus 0-4 --name=prometheus -d -p 9090:9090 --net=host -v $PWD/prometheus.yml:/etc/prometheus/prometheus.yml  prom/prometheus
docker run --cpuset-cpus 0-4 -d --name=grafana --net=host -p 3000:3000 grafana/grafana


docker network rm apinet 
docker build -t mcastelino/test-api-server:latest .

# Cleanup

# Docker DNS does not work with Kata (hence use explicit IP)
# Use a custom docker network so that you can control the network
docker network create apinet --subnet=192.168.211.0/24
docker run --network apinet --name=zipkin-net -d \
                                         --ip 192.168.211.2 \
                                         openzipkin/zipkin

docker run --network apinet --name=prometheus-net -d \
        -v $PWD/prometheus_docker.yml:/etc/prometheus/prometheus.yml \
                                         --ip 192.168.211.3 \
                                         prom/prometheus
