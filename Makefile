build:
	go build -o build/agent.exe ./agent/cmd/main.go
	go build -o build/orchestrator.exe ./orchestrator/cmd/main.go
	go build -o build/web.exe ./web/cmd/main.go

d-build:
	docker build -t agent:latest -f agent/Dockerfile .
	docker build -t orchestrator:latest -f orchestrator/Dockerfile .
	docker build -t web:latest -f web/Dockerfile .