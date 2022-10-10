#!/usr/bin/env sh

case $1 in
app)
  chmod +x /repo/app
  /repo/app --config-file /repo/configs/app.yaml
  ;;
execution)
  chmod +x /repo/execution
  /repo/execution --config-file /repo/configs/app.yaml
  ;;
*)
  echo "./scripts/run.sh [app]"
  ;;
esac
