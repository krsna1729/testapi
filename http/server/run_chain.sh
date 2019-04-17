go build
docker rm -f zipkin-latency prometheus-latency grafana-latency root branch leaf
killall server

docker run --cpuset-cpus 0-4 --name=zipkin-latency -d -p 9411:9411 --net=host openzipkin/zipkin
docker run --cpuset-cpus 0-4 --name=prometheus-latency -d -p 9090:9090 --net=host -v $PWD/prometheus.yml:/etc/prometheus/prometheus.yml  prom/prometheus
docker run --cpuset-cpus 0-4 -d --name=grafana-latency --net=host -p 3000:3000 grafana/grafana

#Let the framework come up
MAX_PRIME='1500000'
PROFILE='./stress.cfg'
UPSTREAM_URI=localhost:8888 DOWNSTREAM_URI=http://localhost:8889 METRICS_PORT=8887 SERVICE_NAME=root PRIME_MAX=$MAX_PRIME JOBFILE=$PROFILE taskset 0xF0 ./server &
UPSTREAM_URI=localhost:8889 DOWNSTREAM_URI=http://localhost:8890 METRICS_PORT=8886 SERVICE_NAME=branch PRIME_MAX=$MAX_PRIME JOBFILE=$PROFILE taskset 0xF00 ./server &
UPSTREAM_URI=localhost:8890                                      METRICS_PORT=8885 SERVICE_NAME=leaf PRIME_MAX=$MAX_PRIME JOBFILE=$PROFILE taskset 0xF000 ./server &

./run_tests.sh
