#!/bin/bash
# Droplet deploy script — source of truth for `~/idilica-backend` on the server.
# The Deploy workflow (main.yml) SSHes in and runs that file after scp'ing
# backend.tar.gz (static Go binaries: api-linux + seed-linux) to
# /home/ellebkey/apps/idilica. After editing this file, copy it to the droplet:
#   scp deploy/idilica-backend.sh ellebkey@<host>:~/idilica-backend && ssh ellebkey@<host> chmod +x idilica-backend
#
# Requisito (una vez): permitir el restart sin password, si no el sudo se
# cuelga en el pipeline. Con `sudo visudo -f /etc/sudoers.d/idilica`:
#   ellebkey ALL=(ALL) NOPASSWD: /usr/bin/systemctl restart idilica-api
set -e

DEPLOY_START=$(date +%s)
log() { echo "==> [$(date '+%H:%M:%S')] $1"; }

log "Deploy started"

cd /home/ellebkey/apps/idilica
log "Artifact: $(du -h backend.tar.gz | cut -f1) backend.tar.gz"

log "Swapping release (previous kept as api.old)"
rm -f api.old seed.old
[ -f api ] && mv api api.old
[ -f seed ] && mv seed seed.old
tar -xzf backend.tar.gz
mv api-linux api
mv seed-linux seed
chmod +x api seed
rm backend.tar.gz

cp ~/secrets/.env.idilica .env
log "Secrets in place"

# Migraciones: el esquema se sincroniza al arrancar (AutoMigrate ≈ sequelize.sync)

log "Restarting systemd service"
sudo systemctl restart idilica-api

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
