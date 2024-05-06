#!/bin/bash
sudo echo '{
  "metrics-addr": "0.0.0.0:9323",
  "experimental": true,
  "storage-driver": "overlay2",
  "insecure-registries": ["10.9.99.10:5001", "10.9.99.10:5002", "localhost:5001", "localhost:5002"]
}
' >/etc/docker/daemon.json
sudo service docker restart
cd /vagrant/docker
docker stop registry
docker rm registry
docker stop apache2
docker rm apache2
yes | sudo docker system prune -a
yes | sudo docker system prune --volumes

sudo mkdir -p /etc/docker
sudo echo '{
  "metrics-addr": "0.0.0.0:9323",
  "experimental": true,
  "storage-driver": "overlay2",
  "insecure-registries": ["10.9.99.10:5001", "10.9.99.10:5002", "localhost:5001", "localhost:5002"]
}
' >/etc/docker/daemon.json

echo "Creating Private Docker Registry"
# https://docs.docker.com/registry/deploying/#customize-the-published-port
docker run -d --restart=always \
  --name registry \
  -v "$(pwd)"/auth:/auth \
  -e "REGISTRY_AUTH=htpasswd" \
  -e "REGISTRY_AUTH_HTPASSWD_REALM=Registry Realm" \
  -e REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd \
  -e REGISTRY_HTTP_ADDR=0.0.0.0:5002 \
  --memory 256M -p 5002:5002 registry:2

cat <<EOF | sudo tee /etc/docker/auth.json
{
  "username": "admin",
  "password": "password",
  "email": "admin@localhost"
}
EOF

echo "Docker Login to Registry"
sleep 10;
sudo --preserve-env=PATH -u vagrant docker login -u="admin" -p="password" http://10.9.99.10:5002

docker ps
echo -e '\e[38;5;198m'"++++ Docker stats"
docker stats --no-stream -a
