#!/bin/sh

# Create xray log directory for connection limit functionality
mkdir -p /var/log/xray
chmod 755 /var/log/xray

# Configure and start fail2ban (required for connection limit functionality)
if [ "${XUI_ENABLE_FAIL2BAN:-true}" = "true" ]; then
    echo "Configuring Fail2Ban for connection limit functionality..."

    # Create fail2ban directories
    mkdir -p /var/run/fail2ban
    mkdir -p /var/log/fail2ban

    # Create optimized jail.local configuration
    cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 600
findtime = 600
maxretry = 3
backend = auto

# Disable default jails to avoid configuration conflicts in Docker
[sshd]
enabled = false

[apache-auth]
enabled = false

[nginx-http-auth]
enabled = false
EOF

    echo "Starting Fail2Ban..."
    fail2ban-client -x start
    if [ $? -eq 0 ]; then
        echo "✓ Fail2Ban started successfully - Connection limit functionality is ready"

        # Test fail2ban-client
        if fail2ban-client status >/dev/null 2>&1; then
            echo "✓ Fail2Ban client is working"
        else
            echo "⚠ Fail2Ban service started but client test failed"
        fi
    else
        echo "⚠ Warning: Failed to start Fail2Ban, connection limit may not work"
        echo "  This is normal in some Docker environments"
    fi
else
    echo "Fail2Ban disabled via XUI_ENABLE_FAIL2BAN=false"
fi

# Run x-ui
exec /app/x-ui
