# wiki-graph

A simple project for finding the shortest path between two Wikipedia pages.

# Components

The project consists of the following components:
- **Server** accepts users' requests, saves them in PostgreSQL and enqueues a new task in RabbitMQ.
- **Client** is a CLI that takes user input, sends it to the server and waits for the task completion.
- **Worker** consumes tasks from the message queue and runs the [BFS algorithm](https://en.wikipedia.org/wiki/Breadth-first_search) to find the shortest path.

When workers run the BFS algorithm, supplementary information is stored in memory.
Information about intermediate graphs is not persisted and is not reused when a new request is being processed.

# Usage

[![asciicast](https://asciinema.org/a/663xm3EDdLftqj6l16BeqJG9O.svg)](https://asciinema.org/a/663xm3EDdLftqj6l16BeqJG9O)

# Installation

## Dependencies

The application requires running PostgreSQL and RabbitMQ instances. You can run them in the Docker environment.

Create an exchange that is bound to a queue in RabbitMQ. The server will be pushing tasks to this exchange
and workers will read them from the queue. It can be easily done with the help of 
[Management plugin](https://www.rabbitmq.com/management.html).

Also, you need to create a table in PostgreSQL:
```
BEGIN;

CREATE TABLE IF NOT EXISTS tasks (
      id varchar(64) primary key not null,
      created_at timestamp without time zone default now() not null,
      updated_at timestamp without time zone default now() not null,

      from_page varchar(512) not null,
      to_page varchar(512) not null,

      status integer default 0 not null,
      result jsonb
);

CREATE INDEX IF NOT EXISTS tasks_created_at_idx ON tasks USING btree(created_at);
CREATE INDEX IF NOT EXISTS tasks_updated_at_idx ON tasks USING btree(updated_at);
CREATE INDEX IF NOT EXISTS tasks_status_idx ON tasks USING btree(status);

COMMIT;
```

## Running

You can run all the components in the host system or in the Docker environment.

### Root system

Build the binaries:
```bash
make build
```

You can find three executable files (server, client and worker) in the `bin` directory now.

### Docker environment

There are three Docker images with tags that speak for themselves: `lodthe/wikigraph-server`, `lodthe/wikigraph-client`, 
`lodthe/wikigraph-worker`.

Also, there is a [docker-compose.yml](./docker-compose.yml) file that is configured to run containers with PostgreSQL, 
RabbitMQ and mentioned images. Envs are read from the [envfiles](./envfiles) directory.

## Configuration

The application reads configuration from envs. Check `.env.*.dist` files in the 
[envfiles](./envfiles) directory for examples.

The most important of them are as follows:

**.env.server**:
```bash
# PostgreSQL DSN.
DB_POSTGRES_DSN=host=localhost port=5432 user=user password=password dbname=wikigraph sslmode=disable

# The server will be listening on this address.
GRPC_SERVER_ADDRESS=0.0.0.0:9000

# The server will be pushing tasks to the specified AMQP exchange.
AMQP_CONNECTION_URL=amqp://user:pass@localhost
AMQP_EXCHANGE_NAME=wikigraph
AMQP_ROUTING_KEY=task
```

**.env.worker**:
```bash
# PostgreSQL DSN.
DB_POSTGRES_DSN=host=localhost port=5432 user=user password=password dbname=wikigraph sslmode=disable

# The worker will be consuming tasks from the specified AMQP queue.
AMQP_CONNECTION_URL=amqp://user:pass@localhost
# You should bind a queue to the exchange used by the server.
AMQP_QUEUE_NAME='wikigraph_tasks'
AMQP_ROUTING_KEY='task'

WIKIPEDIA_API_URL='https://en.wikipedia.org/w/api.php'
WIKIPEDIA_API_RPS='50'

# Maximum allowed distance between pages in requests.
BFS_DISTANCE_THRESHOLD='2'
```

**.env.client**:
```bash
GRPC_SERVER_ADDRESS='localhost:9000'
```