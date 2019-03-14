go build
docker rm -f zipkin-latency prometheus-latency grafana-latency
killall server

docker run --name=zipkin-latency -d -p 9411:9411 --net=host openzipkin/zipkin
docker run --name=prometheus-latency -d -p 9090:9090 --net=host -v $PWD/prometheus.yml:/etc/prometheus/prometheus.yml  prom/prometheus
docker run -d --name=grafana-latency --net=host -p 3000:3000 grafana/grafana
UPSTREAM_URI=localhost:8888 DOWNSTREAM_URI=http://localhost:8889 CPU_BUSYWORK=10 METRICS_PORT=8887 SERVICE_NAME=root ./server &
UPSTREAM_URI=localhost:8889 DOWNSTREAM_URI=http://localhost:8890 CPU_BUSYWORK=10 METRICS_PORT=8886 SERVICE_NAME=branch ./server &
UPSTREAM_URI=localhost:8890                                      CPU_BUSYWORK=10 METRICS_PORT=8885 SERVICE_NAME=leaf ./server &
