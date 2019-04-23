# Baseline performance on a given type of machine
# This captures the basic profile with no containers, no iptables and namespaces
killall server

MAX_PRIME='1500000'
PROFILE='./stress.cfg'
TASK1="0xF0"
TASK2="0xF00"
TASK3="0xF000"
UPSTREAM_URI=localhost:8888 DOWNSTREAM_URI=http://localhost:8889 METRICS_PORT=8887 SERVICE_NAME=root PRIME_MAX=$MAX_PRIME JOBFILE=$PROFILE taskset $TASK1 ./server &
UPSTREAM_URI=localhost:8889 DOWNSTREAM_URI=http://localhost:8890 METRICS_PORT=8886 SERVICE_NAME=branch PRIME_MAX=$MAX_PRIME JOBFILE=$PROFILE taskset $TASK2 ./server &
UPSTREAM_URI=localhost:8890                                      METRICS_PORT=8885 SERVICE_NAME=leaf PRIME_MAX=$MAX_PRIME JOBFILE=$PROFILE taskset $TASK3 ./server &
