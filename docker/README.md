## Usage

### Prerequisites
- docker
- docker-compose
- jq 

### Configuration

There are many different approaches to bootstrap the application and all necessary dependencies. To make things more comfortable to use, `docker-compose.yml` file has been defined.

```bash
# Clone repo into your working dir
git clone git@github.com:pidumenk/gocrud.git

# Start services
docker compose -f ./docker/docker-compose.yml up -d

# Send POST request with custom data
curl -X POST -H "Content-Type: application/json" -d '{"name": "Linux", "species": "Pinguin", "breed": "Debian"}' localhost:8080/v1/pet
{"id":"65c4164aa2f49c0a123fc50c"}

# Retreive the existing value
curl localhost:8080/v1/pet/65c4164aa2f49c0a123fc50c | jq .

{
  "id": "65c4164aa2f49c0a123fc50c",
  "name": "Linux",
  "species": "Pinguin",
  "breed": "Debian"
}

# Stop services
docker compose -f ./docker/docker-compose.yml down
```

Manual provisioning:

The default bridge network allows simple container-to-container communication by IP address, and is created by default. A user-defined bridge network allows containers to communicate with each other, by using their container name as a hostname.

```bash
# Build gocrud app from Dockerfile
docker build -t gocrud -f docker/Dockerfile .

# Pull and run mongodb container
docker run --rm -d -p 27017:27017 --name mongodb mongo

# Export IP address of MongoDB running container as the default bridge network is used 
export DOCKER_MONGODB_IP=$(docker inspect -f '{{.NetworkSettings.IPAddress}}' mongodb)

# Run gocrud app 
docker run --rm -d -p 8080:8080 -e GOCRUD_MONGO_URI=mongodb://${DOCKER_MONGODB_IP}:27017 --name=gocrud gocrud

# Send POST request with custom data
curl -X POST -H "Content-Type: application/json" -d '{"name": "Linux", "species": "Pinguin", "breed": "Debian"}' localhost:8080/v1/pet
{"id":"65c4164aa2f49c0a123fc50c"}

# Retreive the existing value
curl localhost:8080/v1/pet/65c4164aa2f49c0a123fc50c | jq .

{
  "id": "65c4164aa2f49c0a123fc50c",
  "name": "Linux",
  "species": "Pinguin",
  "breed": "Debian"
}
```