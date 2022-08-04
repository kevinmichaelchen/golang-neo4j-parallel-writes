## Getting started

### Start neo4j
```bash
mkdir -p $HOME/neo4j/data

docker run \
  -d \
  -p 7474:7474 \
  -p 7687:7687 \
  -v $HOME/neo4j/data:/data \
  neo4j:latest
```

### Run Go code
```
go run main.go
```

## Performance Benchmarking
Need to compare with index vs without index.

```
CREATE INDEX idx_action_time FOR (n:Action) ON (n.time)
```