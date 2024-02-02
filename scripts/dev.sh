#!/usr/bin/env bash

# USAGE:
# ./scripts/dev.sh ./cmd/http-server
#
# This script is used for development purposes.
# It automatically monitors the files in the project.
# If a change is detected, the script rebuilds and reruns the program.
# The script also manages the SIGINT signal: if the script is stopped, the running program is also properly stopped.
# Please, ensure to run the script pointing it to the correct go command to execute.

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
PROJ_DIR="$SCRIPT_DIR/.."

pushd "$PROJ_DIR" > /dev/null

DIRECTORIES_TO_WATCH="$PWD/internal $PWD/cmd $PWD/scripts $PWD/vendor"

monitor_file_changes() {
  if command -v inotifywait &> /dev/null; then
    inotifywait -e modify -e move -e create -e delete -r $DIRECTORIES_TO_WATCH &> /dev/null
  elif command -v fswatch &> /dev/null; then
    fswatch -01 --recursive --event Created --event Removed --event Updated --event Renamed $DIRECTORIES_TO_WATCH &> /dev/null
  else
    echo "Error: Neither fswatch nor inotifywait commands are available"
    exit 1
  fi
}

build_and_run_program() {
  printf "\n⚙️  COMPILING\n\n"

  temp_file=$(mktemp)
  go build -o "$temp_file" "$1"
  if [ $? -eq 0 ]; then
    chmod +x "$temp_file"
    shift; # remove first argument
    "$temp_file" $@ &
    program_pid=$!
  else
    printf "\n🛠️  FIX COMPILATION ISSUES ☝️\n\n"
  fi
}

kill_and_wait() {
  if [[ -n $program_pid ]] && kill -0 "$program_pid" 2>/dev/null; then
    printf "\n💀  TERMINATING\n\n"
    kill -s SIGINT "$program_pid";
    wait "$program_pid";
    unset program_pid
  fi
}

trap_handler() {
  kill_and_wait
  exit 0;
}

trap trap_handler SIGINT SIGTERM

build_and_run_program $@

while monitor_file_changes; do
  kill_and_wait
  build_and_run_program $@
done

popd > /dev/null
