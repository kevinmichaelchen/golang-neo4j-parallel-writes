## Getting started

```bash
mkdir -p $HOME/neo4j/data

docker run \
  -d \
  -p 7474:7474 \
  -p 7687:7687 \
  -v $HOME/neo4j/data:/data \
  neo4j:latest

go run main.go
```
