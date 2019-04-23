docker rm -f root branch leaf
killall server

#Let the framework come up
MAX_PRIME='1500000'
PROFILE='./stress.cfg'
UPSTREAM_URI=localhost:8888 DOWNSTREAM_URI=http://localhost:8889 METRICS_PORT=8887 SERVICE_NAME=root PRIME_MAX=$MAX_PRIME JOBFILE=$PROFILE taskset 0xF0 ./server &
UPSTREAM_URI=localhost:8889 DOWNSTREAM_URI=http://localhost:8890 METRICS_PORT=8886 SERVICE_NAME=branch PRIME_MAX=$MAX_PRIME JOBFILE=$PROFILE taskset 0xF00 ./server &
UPSTREAM_URI=localhost:8890                                      METRICS_PORT=8885 SERVICE_NAME=leaf PRIME_MAX=$MAX_PRIME JOBFILE=$PROFILE taskset 0xF000 ./server &

echo "Bare Metal Tests:"
./run_tests.sh
