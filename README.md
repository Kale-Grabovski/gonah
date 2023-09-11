## dockertest and gitlab/github deploys

Run locally:

```
cp config-example.yaml config.yaml
docker-compose up
```

To run gitlab deployment set `SSH_*` variables at project settings and
change ssh-command to restart the process (see `.gitlab-ci.yml`).

Run tests:
```
go test ./...
```
