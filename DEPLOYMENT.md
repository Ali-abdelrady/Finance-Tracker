# Docker deployment

## Requirements

- Docker Engine
- Docker Compose v2 (`docker compose`)

## Start the application

1. Copy `.env.example` to `.env`.
2. Replace `POSTGRES_PASSWORD` and `JWT_SECRET` with long, unique alphanumeric secrets. Keep `.env` private and do not commit it.
3. From the project root, run:

   ```sh
   docker compose up --build -d
   ```

4. Open `http://<vm-address>:80` (or use the port configured by `APP_PORT`).

The frontend is served by Nginx, which proxies `/api/` requests to the backend. Only the frontend port is published; PostgreSQL and the backend remain internal to the Compose network. The database is initialized from `backend/migration.sql` on first startup and stored in the `postgres_data` volume.

To follow service logs, run `docker compose logs -f`. To stop the containers without deleting database data, run `docker compose down`.

The initialization SQL only runs when PostgreSQL creates a new, empty data directory. Existing database volumes are not modified automatically. To change the published web port, update `APP_PORT` in `.env` before starting the stack.

The frontend Docker build runs Vite directly. The repository's regular TypeScript build currently reports pre-existing type errors, which are separate from producing the deployable static frontend bundle.
