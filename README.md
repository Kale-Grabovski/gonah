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

Push container to custom registry:
```bash
docker ps | grep gonah # get containerID
docker container commit {{containerID}} gonah:v1
docker image tag gonah:v1 localhost:5000/gonah:v1
docker image push localhost:5000/gonah:v1
```

```bash
kubectl create secret generic regcred \
--from-file=.dockerconfigjson=/home/ka/.docker/config.json \
--type=kubernetes.io/dockerconfigjson
```

To run gitlab deployment set `SSH_*` variables at project settings and
change ssh-command to restart the process (see `.gitlab-ci.yml`).

Run tests:

```bash
make test
```
