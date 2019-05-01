RUNTIME=runc
PROFILE='/stress.cfg'
LOAD='/load.cfg'
MAX_PRIME='250000'
TASK1="4-7"
TASK2="8-11"
TASK3="12-15"

killall server
docker rm -f root branch leaf

docker run --network apinet --name=root --hostname=lroot --runtime="$RUNTIME" -d \
                                         --cpuset-cpus $TASK1 \
                                         --net=host \
                                         -e UPSTREAM_URI='localhost:8888' \
                                         -e DOWNSTREAM_URI='http://localhost:8889' \
                                         -e REPORTER_URI='http://192.168.211.2:9411/api/v2/spans' \
                                         -e METRICS_PORT='8887' \
                                         -p 8888:8888 \
                                         -e JOBFILE=$PROFILE \
                                         -e LOADFILE=$LOAD \
                                         -e PRIME_MAX=$MAX_PRIME \
                                         -e GOGC=off \
                                         -v $(pwd)/stress.cfg:/stress.cfg \
                                         -v $(pwd)/load.cfg:/load.cfg \
                                         mcastelino/test-api-server:latest

docker run --network apinet --name=branch --hostname=lbranch --runtime="$RUNTIME" -d \
                                         --cpuset-cpus $TASK2 \
                                         --net=host \
                                         -e UPSTREAM_URI='localhost:8889' \
                                         -e DOWNSTREAM_URI='http://localhost:8890' \
                                         -e REPORTER_URI='http://192.168.211.2:9411/api/v2/spans' \
                                         -e METRICS_PORT='8886' \
                                         -e JOBFILE=$PROFILE \
                                         -e LOADFILE=$LOAD \
                                         -e PRIME_MAX=$MAX_PRIME \
                                         -e GOGC=off \
                                         -v $(pwd)/stress.cfg:/stress.cfg \
                                         -v $(pwd)/load.cfg:/load.cfg \
                                         mcastelino/test-api-server:latest

docker run --network apinet --name=leaf --hostname=lleaf --runtime="$RUNTIME" -d \
                                         --cpuset-cpus $TASK3 \
                                         --net=host \
                                         -e UPSTREAM_URI='localhost:8890' \
                                         -e REPORTER_URI='http://192.168.211.2:9411/api/v2/spans' \
                                         -e METRICS_PORT='8885' \
                                         -e JOBFILE=$PROFILE \
                                         -e LOADFILE=$LOAD \
                                         -e PRIME_MAX=$MAX_PRIME \
                                         -e GOGC=off \
                                         -v $(pwd)/stress.cfg:/stress.cfg \
                                         -v $(pwd)/load.cfg:/load.cfg \
                                         mcastelino/test-api-server:latest
