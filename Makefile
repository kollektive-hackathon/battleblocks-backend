dev: export DB_URL = postgresql://tfc:tfc@localhost:5432/battleblocks
dev: export PORT = :3044
dev:
	go run cmd/main.go
