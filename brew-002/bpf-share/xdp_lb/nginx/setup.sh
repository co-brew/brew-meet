#!/bin/bash
# if docker iptables is not enabled, run the following line.
# iptables -t nat -A POSTROUTING -s 172.17.0.0/16 ! -d 172.17.0.0/17 -j MASQUERADE  

set -ex
base_image_id="martinho0330/ubuntu:22.04"
docker build . -f Dockerfile.ubuntu -t $base_image_id
if [ "$1" == "" ]; then
    docker ps -f name=backend-A |grep -v "CONTAINER ID" |awk '{print $1}' |xargs -i docker rm -f {}
    docker ps -f name=backend-B |grep -v "CONTAINER ID" |awk '{print $1}' |xargs -i docker rm -f {}
    ngx_image_id="martinho0330/nginx:latest"
    docker build . -f Dockerfile.nginx --network host -t $ngx_image_id
    docker run -d --rm --name backend-A -h backend-A --env TERM=xterm-color $ngx_image_id
    docker run -d --rm --name backend-B -h backend-B --env TERM=xterm-color $ngx_image_id
elif [ "$1" == "lb" ]; then 
    docker run --rm -it -v /home/martinho/bpf-share/xdp_lb:/xdp_lb --privileged \
        -h lb --name lb --env TERM=xterm-color $base_image_id
elif [ "$1" == "client" ]; then
    docker run --rm -it -h client --name client --env TERM=xterm-color $base_image_id
fi

