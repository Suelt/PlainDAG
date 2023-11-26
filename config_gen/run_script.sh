#!/bin/bash


# 运行go程序
go run .

# 等待2秒
sleep 2

# 切换到ansible目录的上级目录
cd ../ansible-remote

# 使用密码执行ansible playbook以配置服务器
ansible-playbook -i ./hosts conf-server.yaml --key-file "~/.ssh/key.pem"


# 使用密码执行ansible playbook以启动服务器
ansible-playbook -i ./hosts run-server.yaml --key-file "~/.ssh/key.pem"



sleep 200

# 杀死进程
ansible-playbook -i ./hosts kill-server.yaml --key-file "~/.ssh/key.pem"

# 使用密码执行ansible playbook以获取结果
ansible-playbook -i ./hosts fetch-results.yaml --key-file "~/.ssh/key.pem"



cd result
# sudo rm -rf node1_0.txt node1_1.txt node1_2.txt
cd ~/shareWithPC/code/go/PlainDAG-DSN/calc_results

python3 calc-latency-throughput.py

cat FinalResults.txt

# python3 calc-latency-throughput.py

# cat FinalResults.txt
