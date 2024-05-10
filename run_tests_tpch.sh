#!/bin/bash
set -euo pipefail
IFS=$'\n\t'
#
# Purpose: Build images for each index version, insert data, run a series of tests
#
# Author: Daniel Garon
# Date: 2024-05-06
#
# NOTE: - This test script uses some hardcoded file paths specific to the machine used for testing
#       - The test also relies on a temporary modification using command line args to control data insertion
#

main() {
    results=/home/andrey/Desktop/test_scripts/results/test_results_tpch_keyX.txt
    local branches=(
        2.3-hlf-im-original
        2.3-hlf-im-version
        2.3-hlf-im-block
    )
    local dataLimits=(0 1 2 4 6 8 10)
    pushd /home/andrey/Documents/insert-tpch/blockchain-indexing/blockchainIndexing
    git checkout tpch
    for branch in "${branches[@]}"; do
        echo "Building images for $branch" >> "$results"
        buildImages "$branch"

        echo "Starting network ..." >> "$results"
        ./startFabric.sh go 
        sleep 10
        pushd ./go

        for ((i=1; i<${#dataLimits[@]}; i++)); do
            echo "DATA SIZE: $(( dataLimits[i-1]*1000000 )) to $(( dataLimits[i]*1000000 ))" >> "$results"
            insert_tpch "$(( dataLimits[i-1]*1000000 ))" "$(( dataLimits[i]*1000000 ))"
            {
                echo "POINT QUERY"
                point_query 9013472 1
                echo "" 

                echo "VERSION QUERY"
                version_query 9013472
                echo "" 
            
                echo "BLOCK RANGE QUERY"
                if (( dataLimits[i]<6 )); then
                    block_range_query 200 1
                else
                    block_range_query 500 100
                fi
                echo "" 
            } >> "$results" 2>&1
        done

        popd
        echo "Stopping network ..." >> "$results"
        ./networkDown.sh 
    done
}

insert_tpch() {
    local filenames=(
        "$HOME/Documents/insert-tpch/sortUnsort12KK/unsorted10Mtpch.json"
    )

    for file in "${filenames[@]}"; do
        echo "Inserting file: $file" >> "$results"
        go run application.go -t BulkInvokeParallel -f "$file" -ds "$1" -de "$2"
    done
}

point_query() {
    for _ in {1..7}; do
        go run application.go -t GetHistoryForVersion -k "$1" -v "$2"
    done
}

version_query() {
    local i
    for (( i=3; i<=7; i+=2 )); do
        for _ in {1..7}; do
            go run application.go -t GetHistoryForVersionRange -k "$1" -s 1 -e "$i"
        done
        echo ""
    done
}

block_range_query() {
    local i
    for (( i=200; i<="$1"; i+="$2" )); do
        for _ in {1..7}; do
            if ! go run application.go -t GetHistoryForBlockRange -s 100 -e "$i" -u 4; then
                return
            fi
        done
        echo ""
    done
}

buildImages() {
    pushd /home/andrey/Desktop/hlf-indexing-middleware
    git checkout "$1"
    make docker-clean 
    echo "y" | docker image prune
    make peer-docker
    make orderer-docker
    popd
}

main

exit 0
