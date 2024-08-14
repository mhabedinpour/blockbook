set -e

# start Ganache for integration test
ganache-cli --account="0xc87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3,100000000000000000000" \
  --account="0xae6ae8e5ccbfb04590405997ee2d52d2b330726137b875053c36d94e974d162f,100000000000000000000" &
ganache=`jobs -p`

# wait for ganache to start
sleep 1
# make sure ganache is still up and running
kill -0 $ganache || (echo "could not start ganache" && exit 1)

function exit_handler() {
    echo "Caught exit signal"
    kill -TERM "$ganache" || true
}

trap exit_handler SIGTERM
trap exit_handler SIGINT

go test $(go list ./... | grep -v /vendor/) -v -race -coverprofile cover.out || true
go tool cover -func=cover.out | grep total || true

# stop ganache
kill $ganache || true