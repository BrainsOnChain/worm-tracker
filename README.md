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
2. The Hyperliquid Block Fetcher
2. The Worm Tracker Frontend

## The HTTP Server
The HTTP server is a simple server that listens for requests from the frontend
and returns the worm data. The html is served from the root `/` and the worm
data is served from `/worm/positions`.

Worm positions are read from the worm database (SQLite) and returned as a JSON.

## The Hyperliquid Block Fetcher
The Hyperliquid Block Fetcher is background runner that listens for new blocks
on the Hyperliquid blockchain. When a new block is found that contains logs from
the DeepWorms contract, the fetcher will parse the logs and save the worm data
to the worm database (SQLite).

## The Worm Tracker Frontend
The Worm Tracker Frontend is a simple web app that displays the worm data in a
2d plane. The worm data is fetched from the HTTP server and displayed as a
series of points and lines.

# Running the Project
To run the project, you will need to be able to run a Go server.

1. Clone the repo
2. Run the server with `go run .`
3. Open the frontend in a browser at `http://localhost:8080`



