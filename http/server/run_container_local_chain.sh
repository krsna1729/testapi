docker rm -f root branch leaf
killall server

RUNTIME=runc
PROFILE='/stress.cfg'
MAX_PRIME='1500000'
docker run --network apinet --name=root --hostname=root --runtime="$RUNTIME" -d \
                                         --cpuset-cpus 4-7 \
                                         --net=host \
                                         -e UPSTREAM_URI='localhost:8888' \
                                         -e DOWNSTREAM_URI='http://localhost:8889' \
                                         -e REPORTER_URI='http://localhost:9411/api/v2/spans' \
                                         -e METRICS_PORT='8887' \
                                         -p 8888:8888 \
                                         -e JOBFILE=$PROFILE \
                                         -e PRIME_MAX=$MAX_PRIME \
                                         -v $(pwd)/stress.cfg:/stress.cfg \
                                         mcastelino/test-api-server:latest

docker run --network apinet --name=branch --hostname=branch --runtime="$RUNTIME" -d \
                                         --cpuset-cpus 8-11 \
                                         --net=host \
                                         -e UPSTREAM_URI='localhost:8889' \
                                         -e DOWNSTREAM_URI='http://localhost:8890' \
                                         -e REPORTER_URI='http://localhost:9411/api/v2/spans' \
                                         -e METRICS_PORT='8886' \
                                         -e JOBFILE=$PROFILE \
                                         -e PRIME_MAX=$MAX_PRIME \
                                         -v $(pwd)/stress.cfg:/stress.cfg \
                                         mcastelino/test-api-server:latest

docker run --network apinet --name=leaf --hostname=leaf --runtime="$RUNTIME" -d \
                                         --cpuset-cpus 12-15 \
                                         --net=host \
                                         -e UPSTREAM_URI='localhost:8890' \
                                         -e REPORTER_URI='http://localhost:9411/api/v2/spans' \
                                         -e METRICS_PORT='8885' \
                                         -e JOBFILE=$PROFILE \
                                         -e PRIME_MAX=$MAX_PRIME \
                                         -v $(pwd)/stress.cfg:/stress.cfg \
                                         mcastelino/test-api-server:latest

echo "Container Local Tests:"
./run_tests.sh
