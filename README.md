# Anomaly Tracker Server
This repo holds the server side implementation of creating, tracking, and managing anomalies for EVE Online.
The aim of this project is to provide a flexible and lightweight API for interacting with anomalies discovered while exploring. It is currently not ready for anything but experimental use and is subject to breaking changes galore. 

This project is implemented using Golang, and a basic understanding of it is assumed.

## Setting up the Project
The simplest way to get the project is to use `go get`:
```
$ go get github.com/rickbau5/anomaly-tracker-server
```

From here, use your method of choice for building/running the app:
```
$ go run cmd/anomaly-tracker-server/main.go
$ go install cmd/anomaly-tracker-server/main.go
$ go build -o ./bin/anomaly-tracker-server cmd/anomaly-tracker-server/main.go
```
When running, specify the MySQL host to connect to. `docker/mysql` contains MySQL commands to 
initialize a new database and set of tables, use those to properly configure the MySQL database for the app.

## Development
A `docker-compose.yml` file is provided that will start up all dependent services and initialize them for a completely contained environment for the app to run within. Generally, they can be started one and then used throughout the development process:

```
$ docker-compose up -d
```

With the required services started, the app can now be run. It will by default connect to the instances 
provided by docker-compose.

```
$ go run cmd/anomaly-tracker-server/main.go
```

## Contributing
Submit any bug reports, suggestions, and etc. in the issues section. Merge Requests will be reviewd and considered.

