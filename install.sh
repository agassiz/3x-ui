#!/bin/bash

red='\033[0;31m'
green='\033[0;32m'
blue='\033[0;34m'
yellow='\033[0;33m'
plain='\033[0m'

cur_dir=$(pwd)

# check root
[[ $EUID -ne 0 ]] && echo -e "${red}Fatal error: ${plain} Please run this script with root privilege \n " && exit 1

# Check OS and set release variable
if [[ -f /etc/os-release ]]; then
    source /etc/os-release
    release=$ID
elif [[ -f /usr/lib/os-release ]]; then
    source /usr/lib/os-release
    release=$ID
else
    echo "Failed to check the system OS, please contact the author!" >&2
    exit 1
fi
echo "The OS release is: $release"

arch() {
    case "$(uname -m)" in
    x86_64 | x64 | amd64) echo 'amd64' ;;
    i*86 | x86) echo '386' ;;
    armv8* | armv8 | arm64 | aarch64) echo 'arm64' ;;
    armv7* | armv7 | arm) echo 'armv7' ;;
    armv6* | armv6) echo 'armv6' ;;
    armv5* | armv5) echo 'armv5' ;;
    s390x) echo 's390x' ;;
    *) echo -e "${green}Unsupported CPU architecture! ${plain}" && rm -f install.sh && exit 1 ;;
    esac
}

echo "Arch: $(arch)"

install_base() {
    case "${release}" in
    ubuntu | debian | armbian)
        apt-get update && apt-get install -y -q wget curl tar tzdata
        ;;
    centos | rhel | almalinux | rocky | ol)
        yum -y update && yum install -y -q wget curl tar tzdata
        ;;
    fedora | amzn | virtuozzo)
        dnf -y update && dnf install -y -q wget curl tar tzdata
        ;;
    arch | manjaro | parch)
        pacman -Syu && pacman -Syu --noconfirm wget curl tar tzdata
        ;;
    opensuse-tumbleweed)
        zypper refresh && zypper -q install -y wget curl tar timezone
        ;;
    *)
        apt-get update && apt-get install -y -q wget curl tar tzdata
        ;;
    esac
}

install_fail2ban() {
    echo -e "${blue}Checking Fail2Ban installation...${plain}"

    # Check if fail2ban is already installed
    if command -v fail2ban-client >/dev/null 2>&1; then
        echo -e "${green}Fail2Ban is already installed.${plain}"

        # Check if fail2ban service is running
        if systemctl is-active --quiet fail2ban 2>/dev/null; then
            echo -e "${green}Fail2Ban service is running.${plain}"
        else
            echo -e "${yellow}Starting Fail2Ban service...${plain}"
            systemctl start fail2ban
            systemctl enable fail2ban
        fi
        return 0
    fi

    echo -e "${yellow}Fail2Ban not found. Installing Fail2Ban for connection limit functionality...${plain}"

    case "${release}" in
    ubuntu | debian | armbian)
        apt-get update
        apt-get install -y fail2ban
        ;;
    centos | rhel | almalinux | rocky | ol)
        # For RHEL-based systems, we need EPEL repository
        if ! rpm -qa | grep -q epel-release; then
            echo -e "${yellow}Installing EPEL repository...${plain}"
            case "${release}" in
            centos)
                if [[ $(rpm -E %{rhel}) == "7" ]]; then
                    yum install -y epel-release
                else
                    dnf install -y epel-release
                fi
                ;;
            rhel | almalinux | rocky | ol)
                dnf install -y epel-release
                ;;
            esac
        fi

        if command -v dnf >/dev/null 2>&1; then
            dnf install -y fail2ban
        else
            yum install -y fail2ban
        fi
        ;;
    fedora | amzn | virtuozzo)
        dnf install -y fail2ban
        ;;
    arch | manjaro | parch)
        pacman -S --noconfirm fail2ban
        ;;
    opensuse-tumbleweed)
        zypper install -y fail2ban
        ;;
    *)
        echo -e "${red}Unsupported operating system for Fail2Ban installation: ${release}${plain}"
        echo -e "${yellow}Please install Fail2Ban manually: ${plain}"
        echo -e "${yellow}  Ubuntu/Debian: sudo apt install fail2ban${plain}"
        echo -e "${yellow}  CentOS/RHEL: sudo yum install fail2ban${plain}"
        echo -e "${yellow}  Fedora: sudo dnf install fail2ban${plain}"
        return 1
        ;;
    esac

    # Create optimized Fail2Ban configuration for 3x-ui
    echo -e "${yellow}Configuring Fail2Ban for 3x-ui...${plain}"

    # Create jail.local with minimal configuration
    cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
# Basic settings for 3x-ui connection limit functionality
bantime = 600
findtime = 600
maxretry = 3
backend = auto

# Disable default jails to avoid configuration conflicts
[sshd]
enabled = false

[apache-auth]
enabled = false

[apache-badbots]
enabled = false

[apache-noscript]
enabled = false

[apache-overflows]
enabled = false

[nginx-http-auth]
enabled = false

[nginx-limit-req]
enabled = false

[nginx-botsearch]
enabled = false

[postfix]
enabled = false

[dovecot]
enabled = false

[postfix-sasl]
enabled = false
EOF

    # Test configuration
    echo -e "${yellow}Testing Fail2Ban configuration...${plain}"
    if fail2ban-client -t >/dev/null 2>&1; then
        echo -e "${green}Fail2Ban configuration test passed.${plain}"
    else
        echo -e "${red}Fail2Ban configuration test failed. Creating fallback configuration...${plain}"
        # Create even simpler fallback configuration
        cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 600
findtime = 600
maxretry = 3
EOF
    fi

    # Start and enable fail2ban service
    if command -v systemctl >/dev/null 2>&1; then
        systemctl start fail2ban
        systemctl enable fail2ban

        # Wait a moment for service to start
        sleep 2

        # Verify installation
        if systemctl is-active --quiet fail2ban; then
            echo -e "${green}Fail2Ban installed, configured and started successfully!${plain}"

            # Test fail2ban-client
            if fail2ban-client status >/dev/null 2>&1; then
                echo -e "${green}Fail2Ban client is working. Connection limit functionality is ready!${plain}"
            else
                echo -e "${yellow}Fail2Ban service is running but client test failed. Connection limit may still work.${plain}"
            fi
        else
            echo -e "${red}Fail2Ban installed but failed to start. Checking logs...${plain}"
            journalctl -u fail2ban --no-pager -n 5
            return 1
        fi
    else
        echo -e "${yellow}systemctl not found. Please start Fail2Ban service manually.${plain}"
    fi

    return 0
}

gen_random_string() {
    local length="$1"
    local random_string=$(LC_ALL=C tr -dc 'a-zA-Z0-9' </dev/urandom | fold -w "$length" | head -n 1)
    echo "$random_string"
}

config_after_install() {
    local existing_hasDefaultCredential=$(/usr/local/x-ui/x-ui setting -show true | grep -Eo 'hasDefaultCredential: .+' | awk '{print $2}')
    local existing_webBasePath=$(/usr/local/x-ui/x-ui setting -show true | grep -Eo 'webBasePath: .+' | awk '{print $2}')
    local existing_port=$(/usr/local/x-ui/x-ui setting -show true | grep -Eo 'port: .+' | awk '{print $2}')
    local URL_lists=(
        "https://api4.ipify.org"
		"https://ipv4.icanhazip.com"
		"https://v4.api.ipinfo.io/ip"
		"https://ipv4.myexternalip.com/raw"
		"https://4.ident.me"
		"https://check-host.net/ip"
    )
    local server_ip=""
    for ip_address in "${URL_lists[@]}"; do
        server_ip=$(curl -s --max-time 3 "${ip_address}" 2>/dev/null | tr -d '[:space:]')
        if [[ -n "${server_ip}" ]]; then
            break
        fi
    done

    if [[ ${#existing_webBasePath} -lt 4 ]]; then
        if [[ "$existing_hasDefaultCredential" == "true" ]]; then
            local config_webBasePath=$(gen_random_string 18)
            local config_username=$(gen_random_string 10)
            local config_password=$(gen_random_string 10)

            read -rp "Would you like to customize the Panel Port settings? (If not, a random port will be applied) [y/n]: " config_confirm
            if [[ "${config_confirm}" == "y" || "${config_confirm}" == "Y" ]]; then
                read -rp "Please set up the panel port: " config_port
                echo -e "${yellow}Your Panel Port is: ${config_port}${plain}"
            else
                local config_port=$(shuf -i 1024-62000 -n 1)
                echo -e "${yellow}Generated random port: ${config_port}${plain}"
            fi

            /usr/local/x-ui/x-ui setting -username "${config_username}" -password "${config_password}" -port "${config_port}" -webBasePath "${config_webBasePath}"
            echo -e "This is a fresh installation, generating random login info for security concerns:"
            echo -e "###############################################"
            echo -e "${green}Username: ${config_username}${plain}"
            echo -e "${green}Password: ${config_password}${plain}"
            echo -e "${green}Port: ${config_port}${plain}"
            echo -e "${green}WebBasePath: ${config_webBasePath}${plain}"
            echo -e "${green}Access URL: http://${server_ip}:${config_port}/${config_webBasePath}${plain}"
            echo -e "###############################################"
        else
            local config_webBasePath=$(gen_random_string 18)
            echo -e "${yellow}WebBasePath is missing or too short. Generating a new one...${plain}"
            /usr/local/x-ui/x-ui setting -webBasePath "${config_webBasePath}"
            echo -e "${green}New WebBasePath: ${config_webBasePath}${plain}"
            echo -e "${green}Access URL: http://${server_ip}:${existing_port}/${config_webBasePath}${plain}"
        fi
    else
        if [[ "$existing_hasDefaultCredential" == "true" ]]; then
            local config_username=$(gen_random_string 10)
            local config_password=$(gen_random_string 10)

            echo -e "${yellow}Default credentials detected. Security update required...${plain}"
            /usr/local/x-ui/x-ui setting -username "${config_username}" -password "${config_password}"
            echo -e "Generated new random login credentials:"
            echo -e "###############################################"
            echo -e "${green}Username: ${config_username}${plain}"
            echo -e "${green}Password: ${config_password}${plain}"
            echo -e "###############################################"
        else
            echo -e "${green}Username, Password, and WebBasePath are properly set. Exiting...${plain}"
        fi
    fi

    /usr/local/x-ui/x-ui migrate
}

install_x-ui() {
    cd /usr/local/

    # Download resources
    if [ $# == 0 ]; then
        tag_version=$(curl -Ls "https://api.github.com/repos/agassiz/3x-ui/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        if [[ ! -n "$tag_version" ]]; then
            echo -e "${red}Failed to fetch x-ui version, it may be due to GitHub API restrictions, please try it later${plain}"
            exit 1
        fi
        echo -e "Got x-ui latest version: ${tag_version}, beginning the installation..."
        wget -N -O /usr/local/x-ui-linux-$(arch).tar.gz https://github.com/agassiz/3x-ui/releases/download/${tag_version}/x-ui-linux-$(arch).tar.gz
        if [[ $? -ne 0 ]]; then
            echo -e "${red}Downloading x-ui failed, please be sure that your server can access GitHub ${plain}"
            exit 1
        fi
    else
        tag_version=$1
        tag_version_numeric=${tag_version#v}
        min_version="2.3.5"

        if [[ "$(printf '%s\n' "$min_version" "$tag_version_numeric" | sort -V | head -n1)" != "$min_version" ]]; then
            echo -e "${red}Please use a newer version (at least v2.3.5). Exiting installation.${plain}"
            exit 1
        fi

        url="https://github.com/agassiz/3x-ui/releases/download/${tag_version}/x-ui-linux-$(arch).tar.gz"
        echo -e "Beginning to install x-ui $1"
        wget -N -O /usr/local/x-ui-linux-$(arch).tar.gz ${url}
        if [[ $? -ne 0 ]]; then
            echo -e "${red}Download x-ui $1 failed, please check if the version exists ${plain}"
            exit 1
        fi
    fi
    wget -O /usr/bin/x-ui-temp https://raw.githubusercontent.com/agassiz/3x-ui/main/x-ui.sh

    # Stop x-ui service and remove old resources
    if [[ -e /usr/local/x-ui/ ]]; then
        systemctl stop x-ui
        rm /usr/local/x-ui/ -rf
    fi

    # Extract resources and set permissions
    tar zxvf x-ui-linux-$(arch).tar.gz
    rm x-ui-linux-$(arch).tar.gz -f

    cd x-ui
    chmod +x x-ui
    chmod +x x-ui.sh

    # Create xray log directory for connection limit functionality
    mkdir -p /var/log/xray
    chmod 755 /var/log/xray

    # Install Fail2Ban for connection limit functionality
    install_fail2ban

    # Check the system's architecture and rename the file accordingly
    if [[ $(arch) == "armv5" || $(arch) == "armv6" || $(arch) == "armv7" ]]; then
        mv bin/xray-linux-$(arch) bin/xray-linux-arm
        chmod +x bin/xray-linux-arm
    fi
    chmod +x x-ui bin/xray-linux-$(arch)

    # Update x-ui cli and se set permission
    mv -f /usr/bin/x-ui-temp /usr/bin/x-ui
    chmod +x /usr/bin/x-ui
    config_after_install

    cp -f x-ui.service /etc/systemd/system/
    systemctl daemon-reload
    systemctl enable x-ui
    systemctl start x-ui

    # Verify connection limit functionality
    echo -e "${blue}Verifying connection limit functionality...${plain}"
    if command -v fail2ban-client >/dev/null 2>&1 && systemctl is-active --quiet fail2ban; then
        echo -e "${green}✓ Fail2Ban is running - Connection limit functionality is available${plain}"
    else
        echo -e "${yellow}⚠ Fail2Ban is not running - Connection limit functionality may not work${plain}"
        echo -e "${yellow}  You can install it later with: sudo apt install fail2ban${plain}"
    fi

    if [[ -d "/var/log/xray" ]]; then
        echo -e "${green}✓ Xray log directory created - Access logging is ready${plain}"
    else
        echo -e "${yellow}⚠ Xray log directory not found${plain}"
    fi

    echo -e "${green}x-ui ${tag_version}${plain} installation finished, it is running now..."
    echo -e ""
    echo -e "┌───────────────────────────────────────────────────────┐
│  ${blue}x-ui control menu usages (subcommands):${plain}              │
│                                                       │
│  ${blue}x-ui${plain}              - Admin Management Script          │
│  ${blue}x-ui start${plain}        - Start                            │
│  ${blue}x-ui stop${plain}         - Stop                             │
│  ${blue}x-ui restart${plain}      - Restart                          │
│  ${blue}x-ui status${plain}       - Current Status                   │
│  ${blue}x-ui settings${plain}     - Current Settings                 │
│  ${blue}x-ui enable${plain}       - Enable Autostart on OS Startup   │
│  ${blue}x-ui disable${plain}      - Disable Autostart on OS Startup  │
│  ${blue}x-ui log${plain}          - Check logs                       │
│  ${blue}x-ui banlog${plain}       - Check Fail2ban ban logs          │
│  ${blue}x-ui update${plain}       - Update                           │
│  ${blue}x-ui legacy${plain}       - legacy version                   │
│  ${blue}x-ui install${plain}      - Install                          │
│  ${blue}x-ui uninstall${plain}    - Uninstall                        │
└───────────────────────────────────────────────────────┘

${green}🎉 Connection Limit Feature Ready!${plain}
${blue}📋 How to use:${plain}
  1. Login to x-ui panel
  2. Edit any client
  3. Set 'Connection Limit' to desired number (e.g., 1)
  4. Click 'Click To Get IPs' to view client IPs
  5. System will automatically limit concurrent connections

${blue}🔧 Access Log:${plain} /var/log/xray/access.log
${blue}🛡️ Fail2Ban Status:${plain} systemctl status fail2ban"
}

echo -e "${green}Running...${plain}"
install_base
install_x-ui $1
