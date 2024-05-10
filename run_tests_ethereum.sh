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
#       - The test also relies on a temporary modification in which data insertion is capped at 2 M per file
#

main() {
    results=/home/andrey/Desktop/test_scripts/results/test_results_ethereum.txt
    local branches=(
        2.3-hlf-im-original
        2.3-hlf-im-version
        2.3-hlf-im-block
    )
    pushd /home/andrey/Documents/insert-tpch/blockchain-indexing/blockchainIndexing
    git checkout ethereum
    for branch in "${branches[@]}"; do
        echo "Building images for $branch" >> "$results"
        buildImages "$branch"

        echo "Starting network ..." >> "$results"
        ./startFabric.sh go 
        sleep 10
        pushd ./go

        # 1M
        echo "DATA SIZE: 1000000" >> "$results"
        insert_ethereum First100K/blockTransactions17000000-17010000.json 
        {
            echo "POINT QUERY"
            point_query 0xf89d7b9c864f589bbf53a82105107622b35eaa40 1900
            echo ""

            echo "VERSION QUERY"
            version_query 0xf89d7b9c864f589bbf53a82105107622b35eaa40 3000 1000
            echo "" 

            echo "BLOCK RANGE QUERY"
            block_range_query
            echo ""
        } >> "$results" 2>&1

		# 2M
        echo "DATA SIZE: 2000000" >> "$results"
        insert_ethereum Second100K/blockTransactions17100000-17125000.json
        {
            echo "POINT QUERY"
            point_query 0xf89d7b9c864f589bbf53a82105107622b35eaa40 2489
            echo ""

            echo "VERSION QUERY"
            version_query 0xf89d7b9c864f589bbf53a82105107622b35eaa40 3000 1000
            echo "" 

            echo "BLOCK RANGE QUERY"
            block_range_query
            echo ""
        } >> "$results" 2>&1

		# 4M
        echo "DATA SIZE: 4000000" >> "$results"
        insert_ethereum Second100K/blockTransactions17125001-17150000.json
        {
            echo "POINT QUERY"
            point_query 0xf89d7b9c864f589bbf53a82105107622b35eaa40 4979
            echo ""

            echo "VERSION QUERY"
            version_query 0xf89d7b9c864f589bbf53a82105107622b35eaa40 3000 1000
            echo ""

            echo "BLOCK RANGE QUERY"
            block_range_query
            echo ""
        } >> "$results" 2>&1

		# 6M
        echo "DATA SIZE: 6000000" >> "$results"
		insert_ethereum Second100K/blockTransactions17150001-17175000.json
        {
            echo "POINT QUERY"
            point_query 0xf89d7b9c864f589bbf53a82105107622b35eaa40 7469
            echo ""

            echo "VERSION QUERY"
            version_query 0xf89d7b9c864f589bbf53a82105107622b35eaa40 10000 3000
            echo "" 

            echo "BLOCK RANGE QUERY"
            block_range_query
            echo ""
        } >> "$results" 2>&1

		# 8M
        echo "DATA SIZE: 8000000" >> "$results"
        insert_ethereum Second100K/blockTransactions17175001-17200000.json
        {
            echo "POINT QUERY"
            point_query 0xf89d7b9c864f589bbf53a82105107622b35eaa40 9959
            echo ""

            echo "VERSION QUERY"
            version_query 0xf89d7b9c864f589bbf53a82105107622b35eaa40 10000 3000
            echo ""

            echo "BLOCK RANGE QUERY"
            block_range_query
            echo ""
        } >> "$results" 2>&1

        # 10M
        echo "DATA SIZE: 10000000" >> "$results"
        insert_ethereum First100K/blockTransactions17030001-17050000.json
        {
            echo "POINT QUERY"
            point_query 0xf89d7b9c864f589bbf53a82105107622b35eaa40 14000
            echo ""

            echo "VERSION QUERY"
            version_query 0xf89d7b9c864f589bbf53a82105107622b35eaa40 10000 3000 
            echo ""

            echo "BLOCK RANGE QUERY"
            block_range_query
            echo ""
        } >> "$results" 2>&1

        popd
        echo "Stopping network ..." >> "$results"
        ./networkDown.sh 
        echo "" >> "$results"

    done
}

insert_ethereum() {
    for file in "$@"; do
        echo "Inserting file: $file" >> "$results"
        go run application.go -t BulkInvokeParallel -f "$HOME/Documents/insert-tpch/ethereum/$file" 
    done
}

point_query() {
    for _ in {1..7}; do
        go run application.go -t GetHistoryForVersion -k "$1" -v "$2"
    done
}

version_query() {
    local i
    for (( i=1000; i<="$2"; i+="$3" )); do
        for _ in {1..7}; do
            if ! go run application.go -t GetHistoryForVersionRange -k "$1" -s 1 -e "$i"; then
                continue
            fi
        done
        echo ""
    done
}

block_range_query() {
    local i
    for (( i=250; i <= 300; i+= 50)); do
        for _ in {1..7}; do
            if ! go run application.go -t GetHistoryForBlockRange -s 100 -e "$i" -u 200; then
                continue
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
