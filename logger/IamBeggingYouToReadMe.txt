Hello, have a seat.

First of all, we are building on the architecture used in go-utils, by the great Nameer. There a couple of things to consider in that regard, to be mentioned later.

There are two major things demoed here. Namely, traces and logs. Jaeger is used to record and view the traces where a Clickhouse DB is used to store the logs. There are some rationles for that (not sure about that, you tell me):
    - There are sometimes infos that are not associated with a particular address
    - Having the logs stored in a unified db allowed for advanced control over the queries and observability(?).
    - Can be connected to tools like Grafana for visualization and stuff (Did not think carefully about this one, just wanted to add a third reason)
There are probablyother reasons, which I forgot.

So basically the overall structure of this demo looks like:

Request flow
------------

                    +-------------------+
                    |  Service 1 (BFF)  |
                    +----+---------+----+
                        /            \
                       v              v
          +-------------------+  +-------------------+
          |     Service 2     |  |     Service 3     |
          +---------+---------+  +-------------------+
                    |
                    v
          +-------------------+
          |      Postgres     |
          +-------------------+


Telemetry flow (all three services)
------------------------------------

  +-----------+   +-----------+   +-----------+
  | Service 1 |   | Service 2 |   | Service 3 |
  +-----+-----+   +-----+-----+   +-----+-----+
        |               |               |
        +---------------+---------------+
                        |
                        v
              +----------------------+
              |    OTel Collector    |
              +-----------+----------+
                          |
              +-----------+-----------+
              |                       |
              v                       v
       +-------------+       +-------------+
       |    Jaeger   |       |  ClickHouse |
       +-------------+       +-------------+


All the services export traces (via spans) and logs to the OTEL Collector, which in turn exports the traces to Jaeger and the logs to the Clickhouse DB.
Please take a look (as described by the outer README.md :))

Now the code structure of the logger. As in the logger package, a Logger interface is defined with a setup func and a couple of getters, one for the tracer provider and the other for the logger provider. Three structs implement that interface, one for each logging_level. So the service initiate the suitable logger based on its config. The logger differ in the implementation of the functions LogError, LogWarn and LogInfo.

In conclusion, our services initiate the logger according to their logging level and use those functions which calls the corresponding span and logger to add to the traces or logs, correspondingly. The mentioned structure is defined in the logger package but is not used in the demo yet. I will delete this line when it is adapted :)

A couple of coniderations, and questions:
    - Each layer of the (controller-usecase-repo) has access to the logger. So, for example, errors triggered in the repo are logged in the repo and the upper layer does not need to worry about that.
    - The whole request is traced in a single span? so no need to create a new span for each layer.

Thank you...
