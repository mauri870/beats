This is the `queue` metricset of the ActiveMQ module.

The metricset provides metrics describing the available ActiveMQ queues, especially exchanged messages (enqueued, dequeued, expired, in-flight), connected consumers, producers and its current length.

To collect data, the module communicates with a Jolokia HTTP/REST endpoint that exposes the JMX metrics over HTTP/REST/JSON (JMX key: `org.apache.activemq:brokerName=localhost,destinationName=sample_queue,destinationType=Queue,type=Broker`).

The queue metricset comes with a predefined dashboard:

![metricbeat activemq queues overview](images/metricbeat-activemq-queues-overview.png)
