#!/bin/bash
# Droplet deploy script — source of truth for `~/idilica-backend` on the server
# (corre como root). The Deploy workflow (main.yml) SSHes in and runs that file
# after scp'ing backend.tar.gz to /home/ellebkey/apps/idilica.
#
# Estructura en el servidor (simétrica con el frontend):
#   /home/ellebkey/apps/idilica/backend/   api, seed, .env
#   /home/ellebkey/apps/idilica/frontend/  la PWA
#
# After editing this file, copy it to the droplet:
#   scp deploy/idilica-backend.sh root@<host>:~/idilica-backend && ssh root@<host> chmod +x idilica-backend
set -e

DEPLOY_START=$(date +%s)
log() { echo "==> [$(date '+%H:%M:%S')] $1"; }

log "Deploy started"

cd /home/ellebkey/apps/idilica
log "Artifact: $(du -h backend.tar.gz | cut -f1) backend.tar.gz"

mkdir -p backend
tar -xzf backend.tar.gz -C backend
rm backend.tar.gz

cd backend
log "Swapping release (previous kept as api.old)"
rm -f api.old seed.old
[ -f api ] && mv api api.old
[ -f seed ] && mv seed seed.old
mv api-linux api
mv seed-linux seed
chmod +x api seed

cp ~/secrets/.env.idilica .env
log "Secrets in place"

# Migraciones: el esquema se sincroniza al arrancar (AutoMigrate)

log "Restarting systemd service"
systemctl restart idilica-api

# Health check — fail the pipeline if the app doesn't come back
PORT=$(grep -oP '^PORT=\K\d+' .env || echo 8101)
log "Health check on :$PORT"
for i in $(seq 1 10); do
  if curl -sf -o /dev/null "http://127.0.0.1:$PORT/api/health-check"; then
    log "App is up (responded on attempt $i)"
    log "Deploy finished in $(( $(date +%s) - DEPLOY_START ))s ✔"
    exit 0
  fi
  sleep 2
done

log "App did NOT respond after 20s — deploy FAILED (previous release in api.old)"
journalctl -u idilica-api --no-pager -n 30
exit 1
