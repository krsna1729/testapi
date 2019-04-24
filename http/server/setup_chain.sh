# Baseline performance on a given type of machine
# This captures the basic profile with no containers, no iptables and namespaces
killall server

MAX_PRIME='250000'
PROFILE='./stress.cfg'
TASK1="0xF0"
TASK2="0xF00"
TASK3="0xF000"
GOGC=off UPSTREAM_URI=localhost:8888 DOWNSTREAM_URI=http://localhost:8889 REPORTER_URI='http://192.168.211.2:9411/api/v2/spans' METRICS_PORT=8887 SERVICE_NAME=broot PRIME_MAX=$MAX_PRIME JOBFILE=$PROFILE taskset $TASK1 ./server &
GOGC=off UPSTREAM_URI=localhost:8889 DOWNSTREAM_URI=http://localhost:8890 REPORTER_URI='http://192.168.211.2:9411/api/v2/spans' METRICS_PORT=8886 SERVICE_NAME=bbranch PRIME_MAX=$MAX_PRIME JOBFILE=$PROFILE taskset $TASK2 ./server &
GOGC=off UPSTREAM_URI=localhost:8890 REPORTER_URI='http://192.168.211.2:9411/api/v2/spans' METRICS_PORT=8885 SERVICE_NAME=bleaf PRIME_MAX=$MAX_PRIME JOBFILE=$PROFILE taskset $TASK3 ./server &
