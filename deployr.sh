#!/bin/bash

# Variables
NGINX_CONF_PATH="/etc/nginx/nginx.conf"
DEPLOY_DIR="/.deployr"
STATIC_REPO_URL="https://github.com/adityacodes30/deployr-daemon"  # Static repo URL
GO_EXECUTABLE="server"  # Go server binary name
DEPLOY_SCRIPT="deployr-daemon.sh"  # Shell script to run deployment

# Check if the correct number of arguments are provided
if [ "$#" -ne 3 ]; then
    echo "Usage: $0 <TARGET_NEXTJS_REPO_URL> <DOMAIN> <EMAIL>"
    exit 1
fi

# Assign arguments to variables
TARGET_NEXTJS_REPO_URL="$1"
DOMAIN="$2"
EMAIL="$3"
# 1. Update the system and install necessary packages
echo "Updating system..."
sudo yum update -y

echo "Installing Nginx, Node.js, NPM, and Certbot..."
sudo amazon-linux-extras enable nginx1
sudo yum install -y nginx nodejs npm certbot python3-certbot-nginx git

# 2. Start and enable Nginx
echo "Starting Nginx..."
sudo systemctl start nginx
sudo systemctl enable nginx

# 3. Replace Nginx configuration with the desired setup
echo "Replacing Nginx configuration..."
sudo tee "$NGINX_CONF_PATH" > /dev/null <<EOF
#!/bin/bash

Variables
NGINX_CONF_PATH="/etc/nginx/nginx.conf"
DEPLOY_DIR="/.deployr"
STATIC_REPO_URL="https://github.com/adityacodes30/deployr-daemon"
GO_EXECUTABLE="server"
DEPLOY_SCRIPT="deployr-daemon.sh"
TARGET_NEXTJS_REPO_URL="https://github.com/adityacodes30/saadhitwebsite.git"
DOMAIN="saadhit.com"
EMAIL="adityaework@gmail.com"

# 1. Update the system and install necessary packages
echo "Updating system..."
sudo yum update -y

echo "Installing Nginx, Node.js, NPM, and Certbot..."
sudo amazon-linux-extras enable nginx1
sudo yum install -y nginx nodejs npm certbot python3-certbot-nginx git

# 2. Start and enable Nginx
echo "Starting Nginx..."
sudo systemctl start nginx
sudo systemctl enable nginx

# 3. Replace Nginx configuration with a simplified setup for SSL acquisition
echo "Replacing Nginx configuration..."
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
            return 200 'Hello from Nginx!';
            add_header Content-Type text/plain;
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

echo "Restarting Nginx..."
sudo systemctl restart nginx

# 4. Run Certbot non-interactively
echo "Running Certbot for HTTPS configuration..."
sudo certbot --nginx --non-interactive --agree-tos -m "$EMAIL" -d "$DOMAIN"

# 5. Install PM2 globally
echo "Installing PM2..."
sudo npm install -g pm2

# 6. Clone the static deployr repository
echo "Cloning static deployr repository into /.deployr..."
if [ -d "$DEPLOY_DIR" ]; then
    sudo rm -rf "$DEPLOY_DIR"
fi
sudo mkdir "$DEPLOY_DIR"
sudo git clone "$STATIC_REPO_URL" "$DEPLOY_DIR"

# 7. Make the Go executable and deployment script executable
echo "Making Go server and deploy script executable..."
sudo chmod +x "$DEPLOY_DIR/$GO_EXECUTABLE"
sudo chmod +x "$DEPLOY_DIR/$DEPLOY_SCRIPT"

# 8. Start the Go server using PM2 with the Next.js repo URL as argument
echo "Starting Go server using PM2..."
pm2 delete go-server || true  # Stop existing instance if running
pm2 start "$DEPLOY_DIR/$GO_EXECUTABLE" --name go-server -- "$TARGET_NEXTJS_REPO_URL"

# 9. Update Nginx configuration to proxy to the application
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
            proxy_pass http://localhost:6213;
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

echo "Restarting Nginx..."
sudo systemctl restart nginx

# Save PM2 processes for reboot persistence
pm2 save

echo "Setup completed successfully!"
EOF

echo "Restarting Nginx..."
sudo systemctl restart nginx

# 4. Run Certbot non-interactively
echo "Running Certbot for HTTPS configuration..."
sudo certbot --nginx --non-interactive --agree-tos -m "$EMAIL" -d "$DOMAIN"

# 5. Install PM2 globally
echo "Installing PM2..."
sudo npm install -g pm2

# 6. Clone the static deployr repository
echo "Cloning static deployr repository into /.deployr..."
if [ -d "$DEPLOY_DIR" ]; then
    sudo rm -rf "$DEPLOY_DIR"
fi
sudo mkdir "$DEPLOY_DIR"
sudo git clone "$STATIC_REPO_URL" "$DEPLOY_DIR"

# 7. Make the Go executable and deployment script executable
echo "Making Go server and deploy script executable..."
sudo chmod +x "$DEPLOY_DIR/$GO_EXECUTABLE"
sudo chmod +x "$DEPLOY_DIR/$DEPLOY_SCRIPT"

# 8. Start the Go server using PM2 with the Next.js repo URL as argument
echo "Starting Go server using PM2..."
pm2 delete go-server || true  # Stop existing instance if running
pm2 start "$DEPLOY_DIR/$GO_EXECUTABLE" --name go-server -- "$TARGET_NEXTJS_REPO_URL"

# Save PM2 processes for reboot persistence
pm2 save

echo "Setup completed successfully!"
