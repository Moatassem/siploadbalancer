# SIP Load Balancer v1.0

Very fast load balancer for _SIP UDP traffic_

## Load Balancing Algorithms: Determines how incoming requests are distributed:

1. **RoundRobin**: Distributes requests sequentially.
2. **MostIdle**: Sends requests to the most idle server.
3. **LeastCost**: Sends requests to the server with the least cost.
4. **LeastHit**: Sends requests to the server with the least hits.
5. **Weighted**: Sends requests to servers based on their assigned weight. If S1:3, S2:2 >> Result: S1, S2, S1, S2, S1, ...
6. **Random**: Sends requests to servers in a random order.

## Features:

- **Health Checks**: Regularly check the status of each server to ensure it's capable of handling requests.
- **Session Persistence (Sticky Sessions)**: Ensures requests from the same client are always sent to the same server.
- **Failover**: Ensures requests are rerouted to healthy servers if a server fails.
- **Scalability**: Ability to handle increasing traffic by adding more servers.

## Configuration:

_See existing [data.json](/data.json) to edit the configuration_

Any changes in the json file, SLB needs to be rebooted. In the future, that won't be necessary.

```json
{
  "ipv4": "192.168.1.2", // Server's IPv4 address
  "sipUdpPort": 5060, // SIP UDP port
  "httpPort": 9080, // HTTP TCP port
  "loadbalancemode": "RoundRobin", // Load balancing algorithm (case sensitive)
  "maxCallAttemptsPerSecond": 10000, // CAPS/Throttling limit (0=Disabled, -1=Unlimited, n=Custom)
  "probingInterval": 15, // SIP server health check interval (in seconds)
  "timeoutTimerDuration": 32, // Dialogue timeout (in seconds) [Ex. Egress server times out]
  "clearTimerDuration": 5, // Dialogue cleanup interval (in seconds)
  "servers": [
    {
      "ipv4": "192.168.1.2",
      "port": 5077,
      "description": "SR1",
      "weight": 3, // Used for Weighted algorithm
      "cost": 5 // Used for LeastCost algorithm
    },
    {
      "ipv4": "192.168.1.2",
      "port": 5070,
      "description": "SR2",
      "weight": 2,
      "cost": 5
    }
  ]
}
```

## Existing API calls:

- `GET /api/v1/stats`
  Get general stats of the server
- `GET /api/v1/config`
  Get running server configuration
- `GET /api/v1/cache`
  Get cached SIP sessions
