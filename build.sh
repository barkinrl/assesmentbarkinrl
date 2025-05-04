#!/bin/bash

# This work is licensed under Apache License, Version 2.0 or later.
# Please read and understand latest version of Licence.

echo $@

mkdir -p bin internal/static 

set -x

cmd=${1:-build}

GOOS=${GOOS:-linux}
CGO_ENABLED=${CGO_ENABLED:-1}
PROJECT=github.com/kazimsarikaya/assesmentbarkinrl

# set default values for environment variables
export OIDC_ISSUER="http://keycloak:8080/realms/assesment-realm"
export OIDC_AUDIENCE="assesment-frontend"

if [ "x$cmd" == "xbuild" ]; then
  REV=$(git describe --long --tags --match='v*' --dirty 2>/dev/null || git rev-list -n1 HEAD)
  NOW=$(date +'%Y-%m-%d_%T')

  shift
  # may be second argument which is one of all, frontend, backend
  # default is all if not provided
  # if provided, then only that part will be built
  # get the second argument
  buildfor=${1:-all}

  if [[ "x$buildfor" == "xall" ]] || [[ "x$buildfor" == "xfrontend" ]]; then
    cd frontend 
    npm install
    npm run build || exit 1
    find dist -type f ! -name '*.gz' -delete # delete all files except .gz files
    rsync -avz --delete dist/ ../internal/static/ || exit 1
    cd ..
  fi
  
  if [[ "x$buildfor" == "xall" ]] || [[ "x$buildfor" == "xbackend" ]]; then
      cat > ./internal/static/static.go <<EOF
package static

import "embed"

//go:embed *
var Static embed.FS
EOF
    GOV=$(go version)
    go mod tidy
    go mod vendor

    GOOS=$GOOS CGO_ENABLED=$CGO_ENABLED go build -ldflags "${LDFLAGS} -X '$PROJECT/internal/config.version=$REV' -X '$PROJECT/internal/config.buildTime=$NOW' -X '$PROJECT/internal/config.goVersion=${GOV}'"  -o ./bin/go_react_mui-$GOOS ./cmd
  fi 

elif [[ "x$cmd" == "xtest" ]]; then
  shift
  ./test.sh $@
else
  echo unknown command $cmd $@
fi
