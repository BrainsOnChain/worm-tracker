# Worm Tracker
This project reads the DeepWorms neuralogical output data from the blockchain
and transaltes it into positions and directions on a 2d plane. The data is
persisted in it's own DB and is used to create a visual representation of the
worm's movements.

# Requirements
This project requires the following:
1. Go 1.22.5
2. SQLite
3. A web browser

# Overview
This repo contains three main components:
1. The HTTP server
2. The Storage Layer
3. The Hyperliquid Block Fetcher

## The HTTP Server
The HTTP server is a simple server that listens for requests from the frontend
and returns the worm data. It consists of two endpoints.

### `/worm/positions?id=`
This endpoint returns the worm data as a JSON. The `id` parameter is the id of
the last position that the client knows of. The server will return all positions
that have an id greater than the `id` parameter with a max of 100 positions.

Response Sample 
```json
[
    {
        "id": 1,
        "blockNumber": 1,
        "transactionHash": "0x1234",
        "x": 0.0,
        "y": 0.0,
        "direction": 0.0,
        "price": 0.0,
        "timestamp": "2021-10-10T00:00:00Z"
    },
    {
        ...
    }
]
```

### `/worm/historical`
This endpoint is used to fetch a sample of historical postions. It returns two
arrays of positions. First is the `recent` which returns the last 100 positions
and the second is the `historical` which is a uniformly distributed sample of
400 positions from the entire history.

Response Sample
```json
{
    "recent": [
        {
            "id": 1,
            "blockNumber": 1,
            "transactionHash": "0x1234",
            "x": 0.0,
            "y": 0.0,
            "direction": 0.0,
            "price": 0.0,
            "timestamp": "2021-10-10T00:00:00Z"
        },
        {
            ...
        }
    ],
    "historical": [
        {
            "id": 1,
            "blockNumber": 1,
            "transactionHash": "0x1234",
            "x": 0.0,
            "y": 0.0,
            "direction": 0.0,
            "price": 0.0,
            "timestamp": "2021-10-10T00:00:00Z"
        },
        {
            ...
        }
    ]
}
```

## Storage Layer
Currently this application uses SQLite as the storage layer. The worm data is
stored in a single `positions` table. We also track the last block number that
we have fetched data from in the `last_block` table.

## The Hyperliquid Block Fetcher
The Hyperliquid Block Fetcher is background runner that listens for new blocks
on the Hyperliquid blockchain. When a new block is found that contains logs from
the DeepWorms contract, the fetcher will parse the logs and save the worm data
to the worm database (SQLite).

# Running the Project
To run the project, you will need to be able to run a Go server.

1. Clone the repo
2. Run the server with `go run .`
3. cURL the server to get the worm data



