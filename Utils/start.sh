#!/bin/bash


create_network ()
{
    err=$(podman network create $1 2>&1) && \
    echo "Network $1 create with success" || \
    echo -e "Network $1 failed to be created\n$err"
}

ssl_network=("ssl_git" "ssl_mail" "ssl_pass" "ssl_orga" "ssl_penpot" "ssl_jenkins" "ssl_reverseme")

folder=$(ls | grep -vE "Utils|post-install|tags")

echo "========K4W43GL3 starter's=========="

for network in "${ssl_network[@]}"; do
    create_network $network || true
done

for docker in $folder; do
    echo Starting $docker
    (cd $docker && sudo docker compose up -d)
done
