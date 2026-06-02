#!/usr/bin/env bash
set -euo pipefail

usage() {
    cat >&2 <<EOF
Usage: $0 -t <api_token> -s <api_secret> [-d <domain_id>] [--dry-run]

  -t          Domeneshop API token
  -s          Domeneshop API secret
  -d          Process a specific domain ID or name
  --all       Process all domains
  --dry-run   Print what would be deleted without making any changes
EOF
    exit 1
}

BASE_URL="https://api.domeneshop.no/v0"
API_TOKEN=""
API_SECRET=""
DOMAIN_ID=""
ALL_DOMAINS=false
DRY_RUN=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        -t) API_TOKEN="$2"; shift 2 ;;
        -s) API_SECRET="$2"; shift 2 ;;
        -d) DOMAIN_ID="$2"; shift 2 ;;
        --all) ALL_DOMAINS=true; shift ;;
        --dry-run) DRY_RUN=true; shift ;;
        -h|--help) usage ;;
        *) echo "Unknown argument: $1" >&2; usage ;;
    esac
done

[[ -z "$API_TOKEN" ]] && { echo "Error: -t (api_token) is required" >&2; usage; }
[[ -z "$API_SECRET" ]] && { echo "Error: -s (api_secret) is required" >&2; usage; }
[[ -z "$DOMAIN_ID" && "$ALL_DOMAINS" == "false" ]] && { echo "Error: either -d <domain_id> or --all is required" >&2; usage; }
[[ -n "$DOMAIN_ID" && "$ALL_DOMAINS" == "true" ]] && { echo "Error: -d and --all are mutually exclusive" >&2; usage; }

for cmd in curl jq; do
    command -v "$cmd" &>/dev/null || { echo "Error: '$cmd' is required but not installed." >&2; exit 1; }
done

api_call() {
    local method="$1"
    local path="$2"
    local tmp_body http_code body

    tmp_body=$(mktemp)
    http_code=$(curl -s -o "$tmp_body" -w "%{http_code}" -X "$method" \
        -u "${API_TOKEN}:${API_SECRET}" "${BASE_URL}${path}")
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

api_get() {
    api_call GET "$1"
}

api_delete() {
    local path="$1"
    if [[ "$DRY_RUN" == "true" ]]; then
        echo "  [dry-run] DELETE ${BASE_URL}${path}"
    else
        api_call DELETE "$path"
    fi
}

process_domain() {
    local domain_id="$1"
    local domain_name="$2"

    echo "Domain: ${domain_name} (id=${domain_id})"

    local records
    records=$(api_get "/domains/${domain_id}/dns")

    # Group by (host, type, data). Within each group keep the lowest ID, delete the rest.
    local duplicates
    duplicates=$(echo "$records" | jq -r '
        group_by(.host + " " + .type + " " + .data)
        | .[]
        | select(length > 1)
        | sort_by(.id)
        | .[1:][]
        | [(.id | tostring), .host, .type, .data]
        | join("\t")
    ')

    if [[ -z "$duplicates" ]]; then
        echo "  No duplicates found"
        return
    fi

    local count=0
    while IFS=$'\t' read -r record_id host type data; do
        echo "  Removing duplicate: id=${record_id} type=${type} host=${host} data=${data}"
        api_delete "/domains/${domain_id}/dns/${record_id}"
        (( count++ )) || true
    done <<< "$duplicates"

    if [[ "$DRY_RUN" == "true" ]]; then
        echo "  Would delete ${count} record(s)"
    else
        echo "  Deleted ${count} record(s)"
    fi
}

if [[ "$ALL_DOMAINS" == "true" ]]; then
    domains=$(api_get "/domains")
    while IFS=$'\t' read -r id name; do
        process_domain "$id" "$name"
    done < <(echo "$domains" | jq -r '.[] | [(.id | tostring), .domain] | join("\t")')
else
    if [[ "$DOMAIN_ID" =~ ^[0-9]+$ ]]; then
        domain_json=$(api_get "/domains/${DOMAIN_ID}")
        name=$(echo "$domain_json" | jq -r '.domain')
    else
        all_domains=$(api_get "/domains")
        match=$(echo "$all_domains" | jq -r --arg name "$DOMAIN_ID" '.[] | select(.domain == $name)')
        if [[ -z "$match" ]]; then
            echo "Error: domain '${DOMAIN_ID}' not found" >&2
            exit 1
        fi
        DOMAIN_ID=$(echo "$match" | jq -r '.id')
        name=$(echo "$match" | jq -r '.domain')
    fi
    process_domain "$DOMAIN_ID" "$name"
fi
