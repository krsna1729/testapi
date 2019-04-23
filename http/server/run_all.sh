./setup_collectors.sh

./setup_chain.sh 
./run_tests.sh "Bare Metal" 1 "10s"  2>&1 | tee latency_all.log
./run_tests.sh "Bare Metal" 3 "10s"  2>&1 | tee -a latency_all.log
./run_tests.sh "Bare Metal" 4 "10s"  2>&1 | tee -a latency_all.log
./run_tests.sh "Bare Metal" 8 "10s"  2>&1 | tee -a latency_all.log
./run_tests.sh "Bare Metal" 40 "10s"  2>&1 | tee -a  latency_all.log
./run_tests.sh "Bare Metal" 80 "10s"  2>&1 | tee -a latency_all.log

./setup_container_local_chain.sh
./run_tests.sh "Host local containers" 1 "10s" 2>&1 | tee -a latency_all.log
./run_tests.sh "Host local containers" 3 "10s" 2>&1 | tee -a latency_all.log
./run_tests.sh "Host local containers" 4 "10s" 2>&1 | tee -a latency_all.log
./run_tests.sh "Host local containers" 8 "10s" 2>&1 | tee -a latency_all.log
./run_tests.sh "Host local containers" 40 "10s" 2>&1 | tee -a latency_all.log
./run_tests.sh "Host local containers" 80 "10s" 2>&1 | tee -a latency_all.log

./setup_container_chain.sh 
./run_tests.sh "Containers" 1 "10s" 2>&1 | tee -a latency_all.log
./run_tests.sh "Containers" 3 "10s" 2>&1 | tee -a latency_all.log
./run_tests.sh "Containers" 4 "10s" 2>&1 | tee -a latency_all.log
./run_tests.sh "Containers" 8 "10s" 2>&1 | tee -a latency_all.log
./run_tests.sh "Containers" 40 "10s" 2>&1 | tee -a latency_all.log
./run_tests.sh "Containers" 80 "10s" 2>&1 | tee -a latency_all.log
