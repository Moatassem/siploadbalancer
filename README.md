## SIP Load Balancer v1.0

Very fast load balancer for _UDP SIP traffic_

- Load Balancing Algorithms: Determines how incoming requests are distributed.
  1- Round Robin: Distributes requests sequentially.
  2- Most Idle: Sends requests to the most idle server.
  3- Least Cost: Sends requests to the server with the least cost.
  3- Least Hit: Sends requests to the server with the least hits.
  4- Weighted: Sends requests to servers based on their assigned weight.

Health Checks: Regularly check the status of each server to ensure it's capable of handling requests.
Session Persistence (Sticky Sessions): Ensures requests from the same client are always sent to the same server.
Failover: Ensures requests are rerouted to healthy servers if a server fails.
Scalability: Ability to handle increasing traffic by adding more servers.
