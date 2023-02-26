dev: export DB_URL = host=localhost user=battleblocks password=Battleblocks11! dbname=battleblocks port=5433
dev: export PORT = :3044
dev:
	go run cmd/main.go
