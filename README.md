# MongoDB Replica Set Infrastructure Documentation

This documentation describes the architecture and deployment process for a fault-tolerant **MongoDB Replica Set** cluster with automatic initialization and load balancing via **HAProxy**.

## System Architecture

The cluster consists of the following components:

* **4 data nodes** (`mongodb1` - `mongodb4`): Full members of the replica set.
* **1 Arbiter** (`mongo-arbiter`): A node without data, used only for voting when selecting the Primary.
* **HAProxy**: A balancer that separates read and write streams.
* **Mongo-init**: An initializer container that performs the initial configuration of the cluster.
* **Go-app**: Client application.

---

## Configuration (Environment Variables)

Parameters in the `.env` file:

| Variable | Description |
| --- | --- |
| `MONGO_INITDB_ROOT_USERNAME` | Database administrator (root). |
| `MONGO_INITDB_ROOT_PASSWORD` | Administrator password. |
| `MONGO_REPLICA_SET` | Name of the replica set (default `rs0`). |
| `MONGO_PORT` | Internal port of data nodes (27017). |
| `MONGO_ARBITER_PORT` | Arbitrator port (27018). |
| `MONGO_KEYFILE_PATH` | Path to the authentication file inside the container. |

---

## Security

1. **RBAC**: Access is protected by the root user's login and password.
2. **Internal Auth**: Uses `keyFile` (`mongodb.key`). This is necessary so that replica set nodes can trust each other.

---

## 1. Starting the cluster

* ### 1. Create environment variables
```bash
cp .env .example.env
```

* ### 2. Create the *mongodb.key* key

```bash
openssl rand -base64 754 > mongodb.key
chmod 600 mongodb.key
sudo chown 999:999 mongodb.key
```

* ### 3. Start the cluster

```bash
docker-compose up --build
```

## 2. Stop the cluster

1. `Ctrl + C`

2.
```bash
docker compose down -v
```

---

## ⚖️ Load balancing (HAProxy)

HAProxy provides a single entry point for the application, separating ports by operation type:

| External port | Traffic type | Logic |
| --- | --- | --- |
| **27019** | **Write** | Directs only to `mongodb1` (Primary by default). |
| **27018** | **Read** | Round-robin balancing between `mongodb2`, `mongodb3`, `mongodb4`. |
| **8404** | **Stats** | Performance statistics (login: `admin`, password: `password`). |

> **Note:** In this configuration, HAProxy is configured statically. If `mongodb1` fails and another node becomes Primary, the write port `27019` will need to be manually switched in `haproxy.cfg`.

---

## 🛠 Maintenance and monitoring

### Checking status from the console

```bash
docker exec -it mongodb1 mongosh -u root -p example --eval “rs.status()”
```

### Checking balancer status

Available via the web interface: `http://localhost:8404/stats`

---

## 7. Interaction from the application (Go)

The official `go-mongodb-driver` driver is used to work with the cluster. In this architecture, connections are split at the application level to optimize the load.

### Connection logic

The application creates two separate clients (or connection pools):

1. **Write Client**: Connects to port `27019` (HAProxy → Primary). Used for all data modification operations (`Insert`, `Update`, `Delete`).
2. **Read Client**: Connects to port `27018` (HAProxy → Secondaries). Used for selection queries (`Find`, `Aggregate`) with the `readPreference=secondary` parameter.

---

**Documentation is current as of:** 2026.

---
