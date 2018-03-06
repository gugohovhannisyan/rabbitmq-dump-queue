# rabbitmq-dump-queue

Dump messages from a RabbitMQ queue to files, without affecting the queue.

## Installation

### Download a release

Precompiled binary packages can be found on the
[releases](https://github.com/dubek/rabbitmq-dump-queue/releases) page.

### Compile from source

If you have [Go](https://golang.org/doc/install) installed, you can install
rabbitmq-dump-queue from source by running:

```
go get https://github.com/gugohovhannisyan/rabbitmq-dump-queue
```

The `rabbitmq-dump-queue` executable will be created in the `$GOPATH/bin`
directory.


## Usage

To dump the first 50 messages of queue `incoming_1` to `/tmp`:

    rabbitmq-dump-queue -uri="amqp://user:password@rabbitmq.example.com:5672/" -queue=incoming_1 -max-messages=50 -output-dir=/tmp

This will create the files `/tmp/msg-0000`, `/tmp/msg-0001`, and so on.

The output filenames are printed one per line to the standard output; this
allows piping the output of rabbitmq-dump-queue to `xargs` or similar utilities
in order to perform further processing on each message (e.g. decompressing,
decoding, etc.).

To include the AMQP headers and properties in the output, add the `-full`
option to the command-line.  This will create the following files:

    /tmp/msg-0000
    /tmp/msg-0000-headers+properties.json
    /tmp/msg-0001
    /tmp/msg-0001-headers+properties.json
    ...

The JSON files have the following structure:

    {
      "headers": {
        "x-my-private-header": "my-value"
      },
      "properties": {
        "correlation_id": "XYZ-9876",
        "delivery_mode": 0,
        "priority": 5
      }
    }


By default, it will not acknowledge messages, so they will be requeued.
Acknowledging messages using the `-ack=true` switch will *remove* them from the
queue, allowing the user to process new messages (see implementation details).

    rabbitmq-dump-queue -uri="amqp://user:password@rabbitmq.example.com:5672/" -queue=incoming_1 -max-messages=50 -output-dir=/tmp -ack=true

Running `rabbitmq-dump-queue -help` will list the available command-line
options.



## Message requeuing implementation details

In order to fetch messages from the queue and later return them in the original
order, rabbitmq-dump-queue uses a standard [AMQP `basic.get` API
call](https://www.rabbitmq.com/amqp-0-9-1-reference.html#basic.get) without
automatic acknowledgements, and it doesn't manually acknowledge the received
messages.  Thus, when the AMQP connection is closed (after all the messages
were received and written to files), RabbitMQ returns all the un-acked messages
(all the messages) back to the queue in their original order.

This means that during the time rabbitmq-dump-queue receives and saves the
messages, the messages are not visible to other consumers of the queue.  This
duration is usually very short (unless you're downloading a lot of messages),
but make sure your system can handle such a situation (or shut down other
consumers of the queue during the time you use this tool).

Note that the same approach is used by RabbitMQ's management HTTP API (the
`/api/queues/{vhost}/{queue}/get` endpoint with `requeue=true`).


## Testing

To run the automated tests, have a RabbitMQ server listen on 127.0.0.1:5672
with user `guest` and password `guest`.  Then run:

    go build .
    go test -v .


## Contributing

Github pull requests and issues are welcome.


## License

rabbitmq-dump-queue is under the MIT License. See the [LICENSE](LICENSE) file
for details.
