# siploadbalancer
Very fast load balancer for SIP UDP traffic

Key Concepts covered:
Load Balancing Algorithms: Determines how incoming requests are distributed.
  1- Round Robin: Distributes requests sequentially.
  2- Least Connections: Sends requests to the server with the fewest active connections.
  3- IP Hash: Uses the client’s IP address to determine which server receives the request.
  4- Weighted Round Robin: Each server is assigned a weight, and requests are distributed based on these weights.

Health Checks: Regularly check the status of each server to ensure it's capable of handling requests.
Session Persistence (Sticky Sessions): Ensures requests from the same client are always sent to the same server.
Failover: Ensures requests are rerouted to healthy servers if a server fails.
Scalability: Ability to handle increasing traffic by adding more servers.
