# Requirement

- install golang
- install docker

# Run server with local server

```
cd backend
bash ../scripts/run_local_frontend_server.sh
bash ../scripts/run_against_psql.sh  ./cmd/app/ --serve_bundle_url 'http://localhost:3535' --migration_file_path ./resources/database/migrations/ ../frontend/index.html  ../frontend/dist/ ./resources