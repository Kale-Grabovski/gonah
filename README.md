## dockertest and gitlab/github deploys

Add local registry as unsecure to /etc/docker/daemon.json:

```
"insecure-registries":["localhost:5000" ]
```

```bash
sudo service docker restart
```

Gen pwd for registry:

```bash
docker run --entrypoint htpasswd httpd:2 -Bbn admin password >> docker/registry/htpasswd
```

Run locally:

```bash
cp config-example.yaml config.yaml
docker-compose up
docker login http://localhost:5000
```

Push container to registry:

```bash
docker ps | grep gonah # get containerID
docker build -t gonah:v1.1.2 .
docker image tag gonah:v1.1.2 pzdc/gonah:v1.1.2
docker image push pzdc/gonah:v1.1.2
```

```bash
kubectl create secret generic regcred \
--from-file=.dockerconfigjson=/home/ka/.docker/config.json \
--type=kubernetes.io/dockerconfigjson
```

To run gitlab deployment set `SSH_*` variables at project settings and
change ssh-command to restart the process (see `.gitlab-ci.yml`).

Get service url for localhost when `minikube` is used:

```bash
minikube service gonah-service --url
```

Run tests:

```bash
make test
```

Port forward and cluster tailf:

```bash
k port-forward statefulset/gonah 8879:8877 
k logs -f -l app=gonah
curl http://localhost:8879/api/v1/users
```
