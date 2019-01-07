[![CircleCI](https://circleci.com/gh/DECODEproject/iotencoder.svg?style=svg)](https://circleci.com/gh/DECODEproject/iotencoder)

# iotencoder

Implementaton of proposed stream encoder interface for the DECODE
IoTPilot/Scale Model.

This component is responsible for subscribing to MQTT topics representing
streams of data from a device, encoding the incoming data and writing it to a
datastore.

Uses an experimental template structure from here:
https://github.com/thingful/go-build-template

## Building

Run `make` or `make build` to build our binary compiled for `linux/amd64`
with the current directory volume mounted into place. This will store
incremental state for the fastest possible build. To build for `arm` or
`arm64` you can use: `make build ARCH=arm` or `make build ARCH=arm64`. To
build all architectures you can run `make all-build`.

Run `make container` to package the binary inside a container. It will
calculate the image tag based on the current VERSION (calculated from git tag
or commit - see `make version` to view the current version). To build
containers for the other supported architectures you can run
`make container ARCH=arm` or `make container ARCH=arm64`. To make all
containers run `make all-container`.

Run `make push` to push the container image to `REGISTRY`, and similarly you
can run `make push ARCH=arm` or `make push ARCH=arm64` to push different
architecture containers. To push all containers run `make all-push`.

Run `make clean` to clean up.

To remove all containers, volumes run `make teardown`.

## Testing

To run the test suite, use the make task `test`. This will run all testcases
inside a containerized environment but pointing at a different DB instance to
avoid overwriting any data stored in your local development DB.

In addition, there is a simple bash script (in `client/client.sh`) that uses
curl to exercise the basic functions of the API. The script inserts 4
entries, then paginates through them, before deleting all inserted data. The
purpose of this script is just to sanity check the functionality from the
command line.

## Configuration

The binary generated for this application is called `iotstore`. It has the following four subcommands:

* `help` - displays help informmation
* `migrate` - allows database migrations to be created and applied
* `server` - the primary command that starts up the server.

For operational use the `server` command is the only one that is generally
required.

**Configuration for `server` command**

| Flag                  | Environment Variable           | Description                                                 | Default value                   | Required |
| --------------------- | ------------------------------ | ----------------------------------------------------------- | ------------------------------- | -------- |
| --addr or -a          | IOTENCODER_ADDR                | The address to which the server binds                       | 0.0.0.0:8080                    | No       |
| --broker-addr or -b   | IOTENCODER_BROKER_ADDR         | Address at which the MQTT broker is listening               | tcp://mqtt.smartcitizen.me:1883 | No       |
| --cert-file or -c     | IOTENCODER_CERT_FILE           | The path to a TLS certificate file to enable TLS            |                                 | No       |
| --database-url        | IOTENCODER_DATABASE_URL        | Connection string for Postgres database                     |                                 | Yes      |
| --datastore or -d     | IOTENCODER_DATASTORE           | Address at which the datastore component is listening       |                                 | Yes      |
| --encryption-password | IOTENCODER_ENCRYPTION_PASSWORD | Password used to encrypt secret tokens we write to Postgres |                                 | Yes      |
| --hashid-length or -l | IOTENCODER_HASHID_LENGTH       | Minimum length of generated ids for streams                 | 8                               | No       |
| --hashid-salt         | IOTENCODER_HASHID_SALT         | Salt value used for generating ids for streams              |                                 | Yes      |
| --key-file or -k      | IOTENCODER_KEY_FILE            | The path to a TLS key file to enable TLS                    |                                 | No       |
| --redis-url           | IOTENCODER_REDIS_URL           | URL at which Redis is listening                             |                                 | Yes      |
| --verbose             | IOTENCODER_VERBOSE             | Flag that if set enables verbose mode                       | False                           | No       |
|                       | SENTRY_DSN                     | Optional DSN string for Sentry error reporting              |                                 | No       |
