#!/usr/bin/env bash
set -euo pipefail

usage() {
    cat >&2 <<EOF
Usage: $0 -t <api_token> -s <api_secret> -d <domain> -H <host> -i <ipv6> [--ttl <seconds>]

  -t          Domeneshop API token
  -s          Domeneshop API secret
  -d          Domain name or numeric ID
  -H          Host (subdomain, use @ for apex)
  -i          IPv6 address
  --ttl       TTL in seconds (default: 3600)
EOF
    exit 1
}

BASE_URL="https://api.domeneshop.no/v0"
API_TOKEN=""
API_SECRET=""
DOMAIN=""
HOST=""
IPV6=""
TTL=3600

while [[ $# -gt 0 ]]; do
    case "$1" in
        -t) API_TOKEN="$2"; shift 2 ;;
        -s) API_SECRET="$2"; shift 2 ;;
        -d) DOMAIN="$2"; shift 2 ;;
        -H) HOST="$2"; shift 2 ;;
        -i) IPV6="$2"; shift 2 ;;
        --ttl) TTL="$2"; shift 2 ;;
        -h|--help) usage ;;
        *) echo "Unknown argument: $1" >&2; usage ;;
    esac
done

[[ -z "$API_TOKEN" ]] && { echo "Error: -t (api_token) is required" >&2; usage; }
[[ -z "$API_SECRET" ]] && { echo "Error: -s (api_secret) is required" >&2; usage; }
[[ -z "$DOMAIN" ]]    && { echo "Error: -d (domain) is required" >&2; usage; }
[[ -z "$HOST" ]]      && { echo "Error: -H (host) is required" >&2; usage; }
[[ -z "$IPV6" ]]      && { echo "Error: -i (ipv6) is required" >&2; usage; }

for cmd in curl jq; do
    command -v "$cmd" &>/dev/null || { echo "Error: '$cmd' is required but not installed." >&2; exit 1; }
done

api_call() {
    local method="$1"
    local path="$2"
    local data="${3:-}"
    local tmp_body http_code body

    tmp_body=$(mktemp)
    if [[ -n "$data" ]]; then
        http_code=$(curl -s -o "$tmp_body" -w "%{http_code}" -X "$method" \
            -u "${API_TOKEN}:${API_SECRET}" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "${BASE_URL}${path}")
    else
        http_code=$(curl -s -o "$tmp_body" -w "%{http_code}" -X "$method" \
            -u "${API_TOKEN}:${API_SECRET}" \
            "${BASE_URL}${path}")
    fi
    body=$(cat "$tmp_body")
    rm -f "$tmp_body"

    case "$http_code" in
        403)
            echo "Error: 403 Forbidden - check your API token and secret" >&2
            exit 1
            ;;
        2??)
            echo "$body"
            ;;
        *)
            echo "Error: HTTP ${http_code} calling ${method} ${BASE_URL}${path}" >&2
            [[ -n "$body" ]] && echo "  Response: ${body}" >&2
            exit 1
            ;;
    esac
}

resolve_domain_id() {
    local input="$1"
    if [[ "$input" =~ ^[0-9]+$ ]]; then
        echo "$input"
        return
    fi
    local all_domains match
    all_domains=$(api_call GET "/domains")
    match=$(echo "$all_domains" | jq -r --arg name "$input" '.[] | select(.domain == $name)')
    if [[ -z "$match" ]]; then
        echo "Error: domain '${input}' not found" >&2
        exit 1
    fi
    echo "$match" | jq -r '.id'
}

DOMAIN_ID=$(resolve_domain_id "$DOMAIN")

existing=$(api_call GET "/domains/${DOMAIN_ID}/dns")
match=$(echo "$existing" | jq -r \
    --arg host "$HOST" \
    --arg data "$IPV6" \
    '.[] | select(.type == "AAAA" and .host == $host and .data == $data) | .id')

if [[ -n "$match" ]]; then
    echo "Record already exists: id=${match} host=${HOST} data=${IPV6}"
    exit 0
fi

payload=$(jq -n \
    --arg host "$HOST" \
    --arg data "$IPV6" \
    --argjson ttl "$TTL" \
    '{"type": "AAAA", "host": $host, "data": $data, "ttl": $ttl}')

response=$(api_call POST "/domains/${DOMAIN_ID}/dns" "$payload")
record_id=$(echo "$response" | jq -r '.id // empty')

echo "Created AAAA record: id=${record_id} host=${HOST} data=${IPV6} ttl=${TTL}"
