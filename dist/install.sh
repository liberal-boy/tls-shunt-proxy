#!/usr/bin/env bash
PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin
export PATH

VSRC_ROOT='/tmp/tls-shunt-proxy'
DOWNLOAD_PATH='/tmp/tls-shunt-proxy/tls-shunt-proxy.zip'

API_URL="https://api.github.com/repos/liberal-boy/tls-shunt-proxy/releases/latest"
DOWNLOAD_URL="$(curl "${PROXY}" -H "Accept: application/json" -H "User-Agent: Mozilla/5.0 (X11; Linux x86_64; rv:74.0) Gecko/20100101 Firefox/74.0" -s "${API_URL}" --connect-timeout 10| grep 'browser_download_url' | cut -d\" -f4)"

echo "${DOWNLOAD_URL}"

mkdir -p "${VSRC_ROOT}"
curl -L -H "Cache-Control: no-cache" -o "${DOWNLOAD_PATH}" "${DOWNLOAD_URL}"
unzip -o -d /usr/local/bin/ "${DOWNLOAD_PATH}"
chmod +x /usr/local/bin/tls-shunt-proxy

useradd tls-shunt-proxy -s /usr/sbin/nologin

mkdir -p '/etc/systemd/system'
curl -L -H "Cache-Control: no-cache" -o '/etc/systemd/system/tls-shunt-proxy.service' 'https://raw.githubusercontent.com/liberal-boy/tls-shunt-proxy/master/dist/tls-shunt-proxy.service'

if [ ! -f "/etc/tls-shunt-proxy/config.yaml" ]; then
  mkdir -p '/etc/tls-shunt-proxy'
  curl -L -H "Cache-Control: no-cache" -o '/etc/tls-shunt-proxy/config.yaml' 'https://raw.githubusercontent.com/liberal-boy/tls-shunt-proxy/master/config.simple.yaml'
fi

mkdir -p '/etc/ssl/tls-shunt-proxy'
chown -R tls-shunt-proxy:tls-shunt-proxy /etc/ssl/tls-shunt-proxy

rm -r "${VSRC_ROOT}"
