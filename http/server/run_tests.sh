run_test() {
    sleep 5
    curl http://localhost:8888"$1"
    taskset 0xF0000 hey  -c "$2" -z "$3" -disable-keepalive http://localhost:8888"$1"
}

echo "Running Tests:" $1
echo "Concurrency level:" $2
echo "Runlength:" $3

echo "Baseline: Pure HTTP forwarding"
run_test "/" "$2" "$3"

echo "Baseline: Prime computation"
run_test "/prime" "$2" "$3"

#No need to run these tests as they are highly variable
exit 

echo "Fork Baseline: "
run_test "/fork" "$2" "$3"

echo "Under load:"
run_test "/stress-ng" "$2" "$3"
