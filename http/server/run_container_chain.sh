# Rebuild all artifacts
docker rm -f root branch leaf 
killall server

# Only the root container exposes the port. Downstream URIs are accessed within the container network
RUNTIME=runc
PROFILE='/stress.cfg'
MAX_PRIME='1500000'
docker run --network apinet --name=root --hostname=root --runtime="$RUNTIME" -d \
                                         --cpuset-cpus 4-7 \
                                         -e UPSTREAM_URI='0.0.0.0:8888' \
                                         -e DOWNSTREAM_URI='http://192.168.211.5:8888' \
                                         -e REPORTER_URI='http://192.168.211.2:9411/api/v2/spans' \
                                         -p 8888:8888 \
                                         --ip 192.168.211.4 \
                                         -e JOBFILE=$PROFILE \
                                         -e PRIME_MAX=$MAX_PRIME \
                                         -v $(pwd)/stress.cfg:/stress.cfg \
                                         mcastelino/test-api-server:latest

docker run --network apinet --name=branch --hostname=branch --runtime="$RUNTIME" -d \
                                         --cpuset-cpus 8-11 \
                                         -e UPSTREAM_URI='0.0.0.0:8888' \
                                         -e DOWNSTREAM_URI='http://192.168.211.6:8888' \
                                         -e REPORTER_URI='http://192.168.211.2:9411/api/v2/spans' \
                                         --ip 192.168.211.5 \
                                         -e JOBFILE=$PROFILE \
                                         -e PRIME_MAX=$MAX_PRIME \
                                         -v $(pwd)/stress.cfg:/stress.cfg \
                                         mcastelino/test-api-server:latest

docker run --network apinet --name=leaf --hostname=leaf --runtime="$RUNTIME" -d \
                                         --cpuset-cpus 12-15 \
                                         -e UPSTREAM_URI='0.0.0.0:8888' \
                                         -e REPORTER_URI='http://192.168.211.2:9411/api/v2/spans' \
                                         --ip 192.168.211.6 \
                                         -e JOBFILE=$PROFILE \
                                         -e PRIME_MAX=$MAX_PRIME \
                                         -v $(pwd)/stress.cfg:/stress.cfg \
                                         mcastelino/test-api-server:latest

echo "Container Chain with Networking:"
./run_tests.sh
