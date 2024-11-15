#!/bin/bash

# Variables
NGINX_CONF_PATH="/etc/nginx/nginx.conf"
DEPLOY_DIR="/.deployr"
STATIC_REPO_URL="https://github.com/adityacodes30/deployr-daemon.git"
GO_EXECUTABLE="server"
DEPLOY_SCRIPT="deployr-daemon.sh"


i# Variables
if [ "$#" -ne 4 ]; then
    echo "Usage: $0 <TARGET_NEXTJS_REPO_URL> <DOMAIN> <EMAIL> <DEPLOYR_PUBKEY>"
    exit 1
fi

TARGET_NEXTJS_REPO_URL="$1"
DOMAIN="$2"
EMAIL="$3"
DEPLOYR_PUBKEY="$4"


echo "Test passed - DEPLOYR_PUBKEY is set and persistent"
echo "Current value: $DEPLOYR_PUBKEY"

echo "Updating system..."
sudo yum update -y

echo "Installing Nginx, Node.js, NPM, and Certbot..."
sudo amazon-linux-extras enable nginx1
sudo yum install -y nginx nodejs npm certbot python3-certbot-nginx git

echo "Starting Nginx..."
sudo systemctl start nginx
sudo systemctl enable nginx

sudo sh -c "echo 'DEPLOYR_PUBKEY=\"$DEPLOYR_PUBKEY\"' >> /etc/environment"

# Load it for current session
export DEPLOYR_PUBKEY="$DEPLOYR_PUBKEY"

echo "Testing DEPLOYR_PUBKEY..."

# Test 1: Check if variable is set
if [ -z "$DEPLOYR_PUBKEY" ]; then
    echo "Error: DEPLOYR_PUBKEY is not set"
    exit 1
fi

if [ "$DEPLOYR_PUBKEY" = "$4" ]; then
    echo "✅ DEPLOYR_PUBKEY is set correctly"
else
    echo "Error: DEPLOYR_PUBKEY does not match the input value"
    echo "Expected: $4"
    echo "Got: $DEPLOYR_PUBKEY"
    exit 1
fi


if grep -q "DEPLOYR_PUBKEY=\"$4\"" /etc/environment; then
    echo "DEPLOYR_PUBKEY is properly set in /etc/environment"
else
    echo "Error: DEPLOYR_PUBKEY not found in /etc/environment"
    exit 1
fi

echo "Updating Nginx configuration to proxy to the application..."
sudo tee "$NGINX_CONF_PATH" > /dev/null <<EOF
user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log;
pid /run/nginx.pid;

include /usr/share/nginx/modules/*.conf;

events {
    worker_connections 1024;
}

http {
    log_format  main  '\$remote_addr - \$remote_user [\$time_local] "\$request" '
                      '\$status \$body_bytes_sent "\$http_referer" '
                      '"\$http_user_agent" "\$http_x_forwarded_for"';

    access_log  /var/log/nginx/access.log  main;

    sendfile            on;
    tcp_nopush          on;
    tcp_nodelay         on;
    keepalive_timeout   65;
    types_hash_max_size 4096;

    include             /etc/nginx/mime.types;
    default_type        application/octet-stream;

    include /etc/nginx/conf.d/*.conf;

    server {
        listen       80;
        listen       [::]:80;
        server_name  $DOMAIN;
        root         /usr/share/nginx/html;

        include /etc/nginx/default.d/*.conf;

        location / {
            proxy_pass http://localhost:3000;
            proxy_http_version 1.1;
            proxy_set_header Upgrade \$http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host \$host;
            proxy_cache_bypass \$http_upgrade;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
        }

        location /deployr/ {
            proxy_pass http://localhost:6213/;
            proxy_http_version 1.1;
            proxy_set_header Upgrade \$http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host \$host;
            proxy_cache_bypass \$http_upgrade;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
        }

        error_page 404 /404.html;
        location = /404.html {
        }

        error_page 500 502 503 504 /50x.html;
        location = /50x.html {
        }
    }
}
EOF

sleep 1

echo "Restarting Nginx..."
sudo systemctl restart nginx

sleep 2


echo "Running Certbot for HTTPS configuration..."
sudo certbot --nginx --non-interactive --agree-tos -m "$EMAIL" -d "$DOMAIN"

sleep 1


echo "Installing PM2..."
sudo npm install -g pm2


echo "Cloning static deployr repository into /.deployr..."
if [ -d "$DEPLOY_DIR" ]; then
    sudo rm -rf "$DEPLOY_DIR"
fi
sudo mkdir "$DEPLOY_DIR"
sudo git clone "$STATIC_REPO_URL" "$DEPLOY_DIR"

sleep 1


echo "Making Go server and deploy script executable..."
sudo chmod +x "$DEPLOY_DIR/$GO_EXECUTABLE"
sudo chmod +x "$DEPLOY_DIR/$DEPLOY_SCRIPT"


echo "Starting Go server using PM2..."
pm2 delete go-server || true  
pm2 start "$DEPLOY_DIR/$GO_EXECUTABLE" --name go-server -- "$TARGET_NEXTJS_REPO_URL"



echo "Restarting Nginx..."
sudo systemctl restart nginx


pm2 save

echo "Setup completed successfully!"