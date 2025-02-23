#!/usr/bin/env bash

set -ex

CMD="$1"
FILE_PREFIX="/tmp/ptcpdump"
FNAME="${FILE_PREFIX}_exist_connection.pcapng"
LNAME="${FILE_PREFIX}_exist_connection.log"
RNAME="${FILE_PREFIX}_exist_connection.read.txt"


function test_ptcpdump() {
  timeout 30s ${CMD} -c 10 ${PTCPDUMP_EXTRA_ARGS} -i any -v --print -w "${FNAME}" \
      'port 22' | tee "${LNAME}" &
  sleep 10
  echo foo
  echo bar
  echo foobar
  echo 233
  echo 666
  wait

  cat "${LNAME}"
  cat "${LNAME}" | grep -F 'cmd /usr/sbin/sshd'
  cat "${LNAME}" | grep '.22 .*Flags \[.\+\]'
}

function test_tcpdump_read() {
  which tcpdump || (apt update || true && apt install -y tcpdump)
  tcpdump -nr "${FNAME}"
  tcpdump -nr "${FNAME}" | grep '.22 .*Flags \[.\+\]'

}

function main() {
    test_ptcpdump
    test_tcpdump_read
}

main
