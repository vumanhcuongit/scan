#!/usr/bin/env sh

case $1 in
app)
  chmod +x /repo/app
  /repo/app --config-file /repo/configs/app.yaml
  ;;
worker)
  chmod +x /repo/worker
  /repo/worker --config-file /repo/configs/app.yaml
  ;;
*)
  echo "./scripts/run.sh [app]"
  ;;
esac
