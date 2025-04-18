cat > entrypoint.sh <<EEOF
#!/bin/bash

cat > /home/postgres/patroni.yml <<__EOF__
bootstrap:
  dcs:
    postgresql:
      use_pg_rewind: true
      parameters:
        archive_mode: '\${DB_ARCHIVE_MODE:-off}'
        archive_command: 'wal-g --config=\${WALG_BACKUP_CONFIG} wal-push %p'
  method: \${DB_BOOTSTRAP_METHOD:-createdb}
  createdb:
    command: 'bash -c "mkdir -p \${PATRONI_POSTGRESQL_DATA_DIR} && initdb -D \${PATRONI_POSTGRESQL_DATA_DIR} --auth-host=md5 --auth-local=trust -E UTF8 --locale=en_US.UTF-8 -k && chown \$(id -u):\$(id -g) -R \${PATRONI_POSTGRESQL_DATA_DIR}"'
  clone:
    command: 'bash -c "mkdir -p \${PATRONI_POSTGRESQL_DATA_DIR} && wal-g --config=\${WALG_RESTORE_CONFIG} backup-fetch \${PATRONI_POSTGRESQL_DATA_DIR} \${DB_BACKUP_NAME:-LATEST} && chown \$(id -u):\$(id -g) -R \${PATRONI_POSTGRESQL_DATA_DIR}"'
    keep_existing_recovery_conf: False
    recovery_conf:
__EOF__

if [[ "xyes" != "x\${DB_RESTORE_DISABLE_WAL_FETCH}" ]]; then
cat >> /home/postgres/patroni.yml <<__EOF__
      restore_command: 'wal-g --config=\${WALG_RESTORE_CONFIG} wal-fetch "%f" "%p"'
__EOF__
fi

cat >> /home/postgres/patroni.yml <<__EOF__
      recovery_target_action: promote
__EOF__

if  [[  "x" != "\${DB_RESTORE_TIME}x"  ]] ; then
cat >> /home/postgres/patroni.yml <<__EOF__
      recovery_target_time: '\${DB_RESTORE_TIME}'
__EOF__
fi

cat >> /home/postgres/patroni.yml <<__EOF__
  pg_hba:
  - host all all 0.0.0.0/0 md5
  - host replication \${PATRONI_REPLICATION_USERNAME} 0.0.0.0/0 md5
  post_init: /home/postgres/post_init.sh
restapi:
  connect_address: '\${PATRONI_KUBERNETES_POD_IP}:8008'
postgresql:
  connect_address: '\${PATRONI_KUBERNETES_POD_IP}:5432'
  authentication:
    superuser:
      password: '\${PATRONI_SUPERUSER_PASSWORD}'
    replication:
      password: '\${PATRONI_REPLICATION_PASSWORD}'
  parameters:
    archive_mode: '\${DB_ARCHIVE_MODE:-off}'
    archive_command: 'wal-g --config=\${WALG_BACKUP_CONFIG} wal-push %p'
__EOF__

unset PATRONI_SUPERUSER_PASSWORD PATRONI_REPLICATION_PASSWORD

exec /usr/bin/python3 /usr/bin/patroni /home/postgres/patroni.yml
EEOF

chmod +x entrypoint.sh

cat > post_init.sh <<EOF
#!/bin/bash
echo "running post init scripts if exits"

if [[ -d \${HOME}/post-init.d ]]; then
  echo "post init scripts found"
  for sql in \$(ls \${HOME}/post-init.d/*.sql); do
    cat \$sql | psql
  done
fi

EOF

chmod +x post_init.sh 

curl -L -o wal-g https://github.com/wal-g/wal-g/releases/download/v3.0.5/wal-g-pg-ubuntu-24.04-amd64
chmod +x wal-g



cat > Containerfile <<EOF
FROM docker.io/library/postgres:17.2-bookworm
RUN apt update && apt install patroni -y && rm -rf /var/lib/apt/lists/*   && localedef -i tr_TR -c -f UTF-8 -A /usr/share/locale/locale.alias tr_TR.UTF-8   && PGHOME=/home/postgres   && mkdir -p \$PGHOME   && chown postgres \$PGHOME   && sed -i "s|/var/lib/postgresql.*|\$PGHOME:/bin/bash|" /etc/passwd   && chmod 775 \$PGHOME   && chmod 664 /etc/passwd   && rm -fr /tmp/*   && apt autoremove -y && apt clean -y
ADD entrypoint.sh /
ADD wal-g /usr/local/bin
RUN mkdir -p /run/postgresql
EXPOSE 5432 8008
USER postgres
WORKDIR /home/postgres
ENV HOME=/home/postgres
ADD post_init.sh /home/postgres/post_init.sh
CMD ["/bin/bash", "/entrypoint.sh"]
EOF

CONTAINER_TAG=$(date "+%Y%m%d_%H%M%S")

podman-remote build -t pg-patroni-17:${CONTAINER_TAG} .


