readonly TLS_CERT_PATH=/etc/boundary.crt.pem
readonly TLS_KEY_PATH=/etc/boundary.key.pem

# Install boundary
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://rpm.releases.hashicorp.com/AmazonLinux/hashicorp.repo
sudo yum -y install boundary-0.6.1-1 # pinning version so that we don't run into db migration errors.

# Installs the boundary as a service for systemd on linux
TYPE=$1
NAME=boundary

sudo cat << EOF > /etc/systemd/system/${NAME}-${TYPE}.service
[Unit]
Description=${NAME} ${TYPE}
[Service]
ExecStart=/usr/bin/${NAME} server -config /etc/${NAME}-${TYPE}.hcl
User=boundary
Group=boundary
LimitMEMLOCK=infinity
AmbientCapabilities=CAP_IPC_LOCK
SecureBits=keep-caps
CapabilityBoundingSet=CAP_SYSLOG CAP_IPC_LOCK
[Install]
WantedBy=multi-user.target
EOF

# Add the boundary system user and group to ensure we have a no-login
# user capable of owning and running Boundary
sudo adduser --system --user-group boundary || true
sudo chown boundary:boundary /etc/${NAME}-${TYPE}.hcl
sudo chown boundary:boundary /usr/bin/boundary

# update ownership of tls key and cert
sudo chown boundary:boundary $TLS_CERT_PATH
sudo chmod 644 $TLS_CERT_PATH
sudo chown boundary:boundary $TLS_KEY_PATH
sudo chmod 640 $TLS_KEY_PATH

# Make sure to initialize the DB before starting the service. This will result in
# a database already initizalized warning if another controller or worker has done this
# already, making it a lazy, best effort initialization
if [ "${TYPE}" = "controller" ]; then
  sudo /usr/bin/boundary database init -skip-initial-login-role-creation -skip-auth-method-creation -skip-host-resources-creation -skip-scopes-creation -skip-target-creation -config /etc/${NAME}-${TYPE}.hcl || true
fi

sudo chmod 664 /etc/systemd/system/${NAME}-${TYPE}.service
sudo systemctl daemon-reload
sudo systemctl enable ${NAME}-${TYPE}
sudo systemctl start ${NAME}-${TYPE}.service