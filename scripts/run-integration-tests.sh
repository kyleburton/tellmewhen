#!/bin/bash
set -eEuo pipefail

# ./tellmewhen -command-exits="date" -tellme-via-running="zenity --info done"
./tellmewhen -command-exits="date" -tellme-via-running="echo done"
