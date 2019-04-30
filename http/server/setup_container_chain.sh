# Only the root container exposes the port. Downstream URIs are accessed within the container network
RUNTIME=runc
PROFILE='/stress.cfg'
LOAD='/load.cfg'
MAX_PRIME='250000'
TASK1="4-7"
TASK2="8-11"
TASK3="12-15"

killall server
docker rm -f root branch leaf

docker run --network apinet --name=root --hostname=croot --runtime="$RUNTIME" -d \
                                         --cpuset-cpus $TASK1 \
                                         -e UPSTREAM_URI='0.0.0.0:8888' \
                                         -e DOWNSTREAM_URI='http://192.168.211.5:8888' \
                                         -e REPORTER_URI='http://192.168.211.2:9411/api/v2/spans' \
                                         -p 8888:8888 \
                                         --ip 192.168.211.4 \
                                         -e JOBFILE=$PROFILE \
                                         -e LOADFILE=$LOAD \
                                         -e PRIME_MAX=$MAX_PRIME \
                                         -e GOGC=off \
                                         -v $(pwd)/stress.cfg:/stress.cfg \
                                         -v $(pwd)/load.cfg:/load.cfg \
                                         mcastelino/test-api-server:latest

docker run --network apinet --name=branch --hostname=cbranch --runtime="$RUNTIME" -d \
                                         --cpuset-cpus $TASK2 \
                                         -e UPSTREAM_URI='0.0.0.0:8888' \
                                         -e DOWNSTREAM_URI='http://192.168.211.6:8888' \
                                         -e REPORTER_URI='http://192.168.211.2:9411/api/v2/spans' \
                                         --ip 192.168.211.5 \
                                         -e JOBFILE=$PROFILE \
                                         -e LOADFILE=$LOAD \
                                         -e PRIME_MAX=$MAX_PRIME \
                                         -e GOGC=off \
                                         -v $(pwd)/stress.cfg:/stress.cfg \
                                         -v $(pwd)/load.cfg:/load.cfg \
                                         mcastelino/test-api-server:latest

docker run --network apinet --name=leaf --hostname=cleaf --runtime="$RUNTIME" -d \
                                         --cpuset-cpus $TASK3 \
                                         -e UPSTREAM_URI='0.0.0.0:8888' \
                                         -e REPORTER_URI='http://192.168.211.2:9411/api/v2/spans' \
                                         --ip 192.168.211.6 \
                                         -e JOBFILE=$PROFILE \
                                         -e LOADFILE=$LOAD \
                                         -e PRIME_MAX=$MAX_PRIME \
                                         -e GOGC=off \
                                         -v $(pwd)/stress.cfg:/stress.cfg \
                                         -v $(pwd)/load.cfg:/load.cfg \
                                         mcastelino/test-api-server:latest
