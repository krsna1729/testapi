run_test() {
    sleep 5
    curl http://localhost:8888"$1"
    taskset 0xF0000 hey  -c 3 -z 10s -disable-keepalive http://localhost:8888"$1"
}

echo "Baseline: Pure HTTP forwarding"
run_test "/"

echo "Baseline: Prime computation"
run_test "/prime"

#echo "Fork Baseline: "
#run_test "/fork"

#echo "Under load:"
#run_test "/stress-ng"
