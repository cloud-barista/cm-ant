# Standalone Development Environment Guide

Audience: developers who want to run and iterate on CM-ANT locally with the
bundled `docker-compose.yaml`.

This stack brings up CM-ANT together with the minimal set of subsystems it
depends on, so you can develop and test without a full multi-framework
deployment.

## 0. TL;DR

```bash
# 1) Prepare environment variables (.env is gitignored)
cp .env.example .env             # first time only — edit secrets if needed

# 2) Build the cm-ant image from source and start the stack
make up                          # runs 'docker compose up -d --build'
#    (make up stops with instructions if .env is missing)

# 3) Initialize CB-Tumblebug (required once — manual, see section 3)
#    CM-ANT refuses to start until CB-Tumblebug is 'initialized'.

# 4) Check
curl -s http://localhost:8880/ant/readyz     # {"ready":true,...}
```

### Make targets

| Target | What it does |
|--------|--------------|
| `make up` | Guard `.env`, build the cm-ant image from source, and start the stack (`docker compose up -d --build`) |
| `make compose-down` | Stop and remove the stack (`docker compose down`) |
| `make logs` | Follow stack logs (`docker compose logs -f`) |
| `make ps` | Show service status (`docker compose ps`) |
| `make env` | Check that `.env` exists; if not, print how to create it and stop |

> `make up` will not start anything until `.env` exists — if it is missing it
> prints the `cp .env.example .env` step and exits, so the stack never comes up
> with unset variables.

> **Key point:** CM-ANT performs a startup fail-fast — it will not boot until
> CB-Tumblebug reports `initialized=true`. So a one-time CB-Tumblebug
> initialization after `up -d` is mandatory (section 3).

## 1. What the stack runs (minimal set)

| Service | Role |
|---------|------|
| cm-ant | The application (built from local source) |
| ant-postgres | CM-ANT database (TimescaleDB) |
| cb-tumblebug | Multi-cloud infrastructure management — CM-ANT's core dependency |
| cb-tumblebug-etcd / cb-tumblebug-postgres | CB-Tumblebug metadata stores |
| cb-spider | CSP connectivity (authentication enforced) |
| mc-terrarium | Required dependency of CB-Tumblebug |
| openbao (dev-mode) | Secret backend (auto init/unseal) |

The versions are pinned in `docker-compose.yaml`. This is a **development**
stack — a minimal set that CM-ANT needs, not the full Cloud-Migrator lineup.

## 2. Prerequisites

- Docker and Docker Compose (v2 recommended)
- CB-Tumblebug source (used for the one-time initialization in section 3)
- `.env` prepared with `cp .env.example .env` (`make env` guards that it exists)
  - The defaults in `.env.example` are for local dev/test convenience.
    Replace them with your own values when you need real credentials.
  - `.env` may hold secrets, so it is **gitignored** — never commit it.

### 2.1 `config.yaml` vs `.env` (avoid confusion)

CM-ANT reads configuration in two layers:

1. **`config.yaml`** — the application defaults (server, load, cost, image,
   spec, database, and connection defaults). It ships inside the image and
   assumes `localhost` addresses.
2. **Environment variables (`.env` via docker-compose)** — these **override**
   `config.yaml` at runtime. CM-ANT maps env vars with the `ANT_` prefix
   automatically (viper `AutomaticEnv`):
   - `ANT_SPIDER_USERNAME` → `spider.username`
   - `ANT_TUMBLEBUG_PASSWORD` → `tumblebug.password`
   - `ANT_DATABASE_HOST` → `database.host`, and so on.

In practice the only values that overlap are the **connection settings**
(spider / tumblebug host, port, user, password, and the database). When running
in containers these are injected via `.env` so CM-ANT reaches the other
services by their container DNS names (`cb-spider`, `cb-tumblebug`). When you
run the source directly on the host, `config.yaml`'s `localhost` defaults apply.

## 3. Initialize CB-Tumblebug (required once, manual)

CM-ANT only starts once CB-Tumblebug reports `initialized=true`. Check it:

```bash
curl -s http://localhost:1323/tumblebug/readyz
# before init: {"ready":true,"initialized":false,"message":"...waiting for init.py"}
```

Initialization follows the official CB-Tumblebug guide (use the same version
tag that the stack runs):

```bash
# clone the CB-Tumblebug tag that matches the running image
git clone -b <cb-tumblebug-tag> https://github.com/cloud-barista/cb-tumblebug.git
cd cb-tumblebug/init

# (1) prepare credentials: fill in template.credentials.yaml and encrypt it
#     ./encCredential.sh  -> ~/.cloud-barista/credentials.yaml.enc
# (2) run initialization (registers credentials in OpenBao + loads Tumblebug
#     connections and common assets)
./multi-init.sh
```

`multi-init.sh` runs two stages: **(1)** register CSP credentials in OpenBao,
then **(2)** `init.py` registers connections in CB-Tumblebug and loads the
common spec/image catalog. On completion CB-Tumblebug flips to
`initialized=true`.

Verify:

```bash
curl -s http://localhost:1323/tumblebug/readyz
# {"ready":true,"initialized":true,"message":"CB-Tumblebug is ready and initialized"}
```

## 4. Use and verify CM-ANT

Once initialization completes, CM-ANT leaves its restart loop and boots
normally:

```bash
curl -s http://localhost:8880/ant/readyz
# {"ready":true,"initialized":true,"message":"CM-Ant is ready",
#  "dependencies":{"db":{"reachable":true,"authenticated":true},
#                  "spider":{"reachable":true,"authenticated":true},
#                  "tumblebug":{"reachable":true,"authenticated":true}}}
```

- Swagger UI: `http://localhost:8880/ant/swagger/index.html`
- Cost and performance features only work once **valid CSP credentials /
  connections** are registered in CB-Tumblebug (the ones you provided during
  initialization).

## 5. Restart behavior

| Action | initialized | cm-ant | CSP features | Rework needed |
|--------|:---:|:---:|:---:|------|
| Restart **cb-tumblebug** only | stays true (connConfig restored from volume) | healthy | ok | none |
| Restart **cm-ant** only | true | healthy (no dependency cascade) | ok | none |
| Restart **openbao** | true | **stays healthy (no crash)** | secrets lost (dev-mode is in-memory) | re-register credentials only |
| `docker compose down` then `up` | true (bind-mounted `container-volume/` persists) | healthy | secrets lost | re-register credentials only |
| Delete `container-volume/` / `down -v` | back to false | crash (fail-fast) | — | full re-init |

Notes:

- **Everyday development:** keep the stack up and cycle just cm-ant (or
  cb-tumblebug) — no re-init needed.
- **After restarting openbao, or after `down` then `up`:** CB-Tumblebug stays
  `initialized` and cm-ant boots fine, but openbao (dev-mode) has lost the CSP
  secrets, so CSP calls fail. Re-register the credentials (rerun the OpenBao
  credential registration, or simply `multi-init.sh` again) — a full re-init is
  not required.
- **Only when you delete `container-volume/`** does CB-Tumblebug fall back to
  the uninitialized state and cm-ant fail-fast again — then a full re-init is
  needed.

CB-Tumblebug detects the existing connection config on boot and restores
`SystemInitialized=true`. The connection metadata persists in the
`container-volume/` bind mount (postgres/etcd), while only the credential
*secrets* live in openbao (dev-mode, in-memory) and are lost on an openbao
restart.

## 6. Recommended workflow — run local source without rebuilding the image

While editing CM-ANT source, avoid rebuilding the Docker image on every change;
run the local source directly instead:

```bash
# 1) Start the stack and initialize it (sections 0-3) — cm-ant runs as an image
make up                   # then do the one-time CB-Tumblebug init

# 2) Stop just the image-based cm-ant (dependencies keep running)
docker compose stop cm-ant

# 3) Run the local source on the host, pointing at the host ports
ANT_SPIDER_HOST=http://localhost   ANT_SPIDER_PORT=1024 \
ANT_TUMBLEBUG_HOST=http://localhost ANT_TUMBLEBUG_PORT=1323 \
ANT_SPIDER_USERNAME=default ANT_SPIDER_PASSWORD=default \
ANT_TUMBLEBUG_USERNAME=default ANT_TUMBLEBUG_PASSWORD=default \
ANT_DATABASE_HOST=localhost \
go run cmd/cm-ant/main.go
```

- `docker compose stop cm-ant` frees port 8880 so the local process can take
  over. cb-tumblebug and cb-spider keep running as containers, reachable on the
  host ports (1323 / 1024).
- Iterate by editing the source and rerunning `go run` — no image rebuild.

## 7. Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| cm-ant stuck `Restarting` (crash loop) | CB-Tumblebug not initialized | Run initialization (section 3) |
| cm-ant log: `cb-tumblebug not initialized (Ready=true, Initialized=false)` | same | section 3 |
| Crashes again after `down` | openbao (dev-mode) lost its secrets | Re-register credentials (section 5) |
| Only CSP calls fail (boot is fine) | registered credentials are invalid | Re-init with valid CSP credentials |
