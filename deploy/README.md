# Server deployment

This project can be deployed on a Linux server with:
- a native PostgreSQL service
- systemd units for the Go API, frontend, and Kratos
- Caddy as the reverse proxy for your DNS domain

Repository deployment files now mirror the current server setup under:

- `deploy/caddy/Caddyfile`
- `deploy/systemd/myproj-api.service`
- `deploy/systemd/myproj-frontend.service`
- `deploy/systemd/kratos.service`
- `deploy/systemd/myproj-api.env.example`
- `deploy/systemd/myproj-frontend.env.example`
- `deploy/kratos/kratos.yml.example`

The API requires these environment variables to be set before it will start:
- `HOST`
- `PORT`
- `DATABASE_URL`
- `OPENAI_API_KEY`
- `MINIO_ENDPOINT`
- `MINIO_ACCESS_KEY`
- `MINIO_SECRET_KEY`
- `MINIO_BUCKET`
- `MINIO_USE_SSL`
- `ORY_URL`

## 1. Point your domain to the server

Create DNS records for your server IP:
- `A example.com -> <server-ip>`
- `A www.example.com -> <server-ip>`

Replace `example.com` everywhere below with your real domain.

## 2. Install packages on the server

For Ubuntu/Debian:

```bash
sudo apt update
sudo apt install -y caddy postgresql postgresql-contrib golang rsync
```

If PostgreSQL, MinIO, or Ory are hosted elsewhere, only install what you need locally.

## 3. Optional: create the local PostgreSQL database

If PostgreSQL will run on the same server, create the database and user:

```bash
sudo -u postgres psql <<'SQL'
CREATE USER appuser WITH PASSWORD 'change-this-password';
CREATE DATABASE appdb OWNER appuser;
GRANT ALL PRIVILEGES ON DATABASE appdb TO appuser;
SQL
```

Then use a matching connection string in `/etc/myproj/api.env`, for example:

```env
DATABASE_URL=postgres://appuser:change-this-password@127.0.0.1:5432/appdb?sslmode=disable
```

A native `postgresql` systemd service is installed by the package manager, so you do not need a custom unit file for it.

## 4. Create the app user and directories

```bash
sudo useradd --system --create-home --home-dir /opt/myproj --shell /usr/sbin/nologin myproj || true
sudo mkdir -p /opt/myproj/bin /opt/myproj/backend /etc/myproj
sudo chown -R myproj:myproj /opt/myproj /etc/myproj
```

## 5. Copy the project to the server

One simple option:

```bash
rsync -avz ./ your-user@your-server:/tmp/myproj-src/
ssh your-user@your-server
sudo rsync -av /tmp/myproj-src/ /opt/myproj/
sudo chown -R myproj:myproj /opt/myproj
```

## 6. Build the API binary

```bash
cd /opt/myproj/backend
sudo -u myproj /usr/bin/env GOCACHE=/opt/myproj/.cache/go-build go build -o /opt/myproj/bin/myproj-api ./cmd/api
```

## 7. Create the environment file

Copy the example file:

```bash
sudo cp /opt/myproj/deploy/systemd/myproj-api.env.example /etc/myproj/api.env
sudo chown myproj:myproj /etc/myproj/api.env
sudo chmod 600 /etc/myproj/api.env
sudo nano /etc/myproj/api.env
```

Set the real values for your database, OpenAI, MinIO, and Ory services.

If PostgreSQL runs on the same server, a typical local URL is:

```env
HOST=127.0.0.1
PORT=8080
DATABASE_URL=postgres://appuser:change-this-password@127.0.0.1:5432/appdb?sslmode=disable
```

## 8. Install the systemd unit

```bash
sudo cp /opt/myproj/deploy/systemd/myproj-api.service /etc/systemd/system/myproj-api.service
sudo systemctl daemon-reload
sudo systemctl enable --now myproj-api.service
```

Check status and logs:

```bash
sudo systemctl status myproj-api.service
sudo journalctl -u myproj-api.service -f
```

## 9. Configure Caddy for your domain

Copy the Caddyfile template and adjust the domain if needed:

```bash
sudo mkdir -p /etc/caddy
sudo cp /opt/myproj/deploy/caddy/Caddyfile /etc/caddy/Caddyfile
sudo nano /etc/caddy/Caddyfile
sudo caddy validate --config /etc/caddy/Caddyfile
sudo systemctl reload caddy
```

Caddy will manage HTTPS automatically for your public domain when DNS is pointed to the server.

## 10. Install and start systemd services

```bash
sudo cp /opt/myproj/deploy/systemd/myproj-api.service /etc/systemd/system/myproj-api.service
sudo cp /opt/myproj/deploy/systemd/myproj-frontend.service /etc/systemd/system/myproj-frontend.service
sudo cp /opt/myproj/deploy/systemd/kratos.service /etc/systemd/system/kratos.service
sudo systemctl daemon-reload
sudo systemctl enable --now myproj-api.service myproj-frontend.service kratos.service
```

## 11. Firewall

If UFW is enabled:

```bash
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

## Updating the service later

After new code is copied to the server:

```bash
cd /opt/myproj/backend
sudo -u myproj /usr/bin/env GOCACHE=/opt/myproj/.cache/go-build go build -o /opt/myproj/bin/myproj-api ./cmd/api
sudo systemctl restart myproj-api.service
```

## Notes

- The API is intended to listen on `127.0.0.1:8080` behind Nginx via `HOST=127.0.0.1` and `PORT=8080`.
- `myproj-api.service` does not hard-depend on a local PostgreSQL unit because your database may be hosted remotely.
- If you want, you can later add a CI/CD step to build and restart the service automatically.

## GitHub Actions CI/CD

This repository includes:

- `.github/workflows/ci.yml` for backend/frontend verification on pushes and pull requests
- `.github/workflows/cd.yml` for deployment to your server on pushes to `main`
- `deploy/scripts/deploy.sh` to build both apps and restart systemd services on the server

### Required GitHub secrets

Add these repository secrets:

- `SERVER_HOST` – server hostname or IP
- `SERVER_USER` – SSH user used by GitHub Actions
- `SERVER_SSH_KEY` – private SSH key for that user

Optional secrets:

- `SERVER_PORT` – defaults to `22`
- `REMOTE_TMP_DIR` – defaults to `/tmp/myproj-release`
- `APP_DIR` – defaults to `/opt/myproj`
- `APP_USER` – defaults to `myproj`
- `APP_GROUP` – defaults to `myproj`
- `API_SERVICE` – defaults to `myproj-api`
- `FRONTEND_SERVICE` – defaults to `myproj-frontend`

### Server requirements for CD

Important: commit only templates and examples. Do not commit real values from `/etc/myproj/*`, `/etc/kratos/kratos.yml`, or any OAuth/OpenAI/DB/MinIO secrets.


The SSH user used by GitHub Actions must be able to run `sudo` for:

- `rsync`
- `mkdir`
- `chown`
- `systemctl restart ...`
- `systemctl status ...`

The server also needs:

- the project present under `/opt/myproj` (or your chosen `APP_DIR`)
- a working `myproj-api.service`
- a working `myproj-frontend.service`
- frontend production env file, for example `/opt/myproj/frontend/.env.production`
- backend env file, for example `/etc/myproj/api.env`

### What the CD workflow does

On every push to `main`, GitHub Actions will:

1. run backend tests and build verification
2. run frontend dependency install and build verification
3. upload the repository to a temporary directory on the server
4. sync files into the app directory while preserving `.env.production` and `.env.local`
5. run `deploy/scripts/deploy.sh`
6. rebuild backend and frontend on the server
7. restart `myproj-api` and `myproj-frontend`

If you prefer, you can disable automatic deploys and trigger the `CD` workflow manually with `workflow_dispatch`.

### Legacy files

- `deploy/nginx/myproj.conf` is kept only as an older example.
- Current production setup uses Caddy, not Nginx.
