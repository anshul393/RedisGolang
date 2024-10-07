build: 
	@go build -o redis_server *.go

run: build
	@./redis_server

redis-cli:
	@docker exec -it 90d redis-cli -h host.docker.internal -p 6666

redis-benchmark:
	@docker exec -it 90d redis-benchmark -h host.docker.internal -p 6666 -t SET,GET -q
