#! /bin/bash
docker ps |grep -E "client|lb|backend"| awk '{print $NF}'|sort |awk '{print "docker exec "$NF" ifconfig  2>&1"}' | bash -x 2>&1|grep exec -A2|grep  -v eth0