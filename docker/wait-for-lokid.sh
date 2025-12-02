#!/usr/bin/env bash
set -Eeuo pipefail

LAST_ERR=""

on_err() {
  local status="$1"
  local errline="$2"
  local code="${status:-1}"
  if [ -n "${LAST_ERR:-}" ]; then
    echo "❌ Healthcheck failed: $LAST_ERR (exit $code)" >&2
  else
    echo "❌ Healthcheck failed (line $errline, exit $code)" >&2
  fi
  exit "$code"
}

trap 'on_err $? $LINENO' ERR

# One-shot healthcheck for lokid.
# Exits 0 if healthy; 1 otherwise.
#
# Usage:
#   wait-for-lokid.sh [--config PATH] [--strict] [--need-peers] [--quiet]
#
# Options:
#   -c, --config     Path to lokid.conf (default: ~/.lokid/lokid.conf)
#       --strict     Also require blocks == headers
#       --need-peers Require at least one peer connection
#   -q, --quiet      Minimal output
#   -h, --help       Show this help

CONFIG="${HOME}/.lokid/lokid.conf"
STRICT=false
NEED_PEERS=false
QUIET=false

usage() { sed -n '1,200p' "$0" | sed -n '1,40p'; }

log() { [ "$QUIET" = false ] && echo "$@"; }
die() {
  local message="$*"
  LAST_ERR="$message"
  echo "ERROR: $message" >&2
  on_err 1 "${BASH_LINENO[0]:-0}"
}

fail_with_log() {
  local message="$1"
  LAST_ERR="$message"
  if [ "$QUIET" = false ]; then
    echo "$message"
  fi
  return 1
}

# ---- parse args ----
while [[ $# -gt 0 ]]; do
  case "$1" in
    -c|--config) CONFIG="${2-}"; shift 2 ;;
    --strict)    STRICT=true; shift ;;
    --need-peers)NEED_PEERS=true; shift ;;
    -q|--quiet)  QUIET=true; shift ;;
    -h|--help)   usage; exit 0 ;;
    *)           die "Unknown argument: $1" ;;
  esac
done

[ -f "$CONFIG" ] || die "Config file not found: $CONFIG"

# ---- tiny INI parser (key=value) ----
conf_get() {
  local key="$1"
  awk -F= -v key="$key" '
    BEGIN { IGNORECASE=1 }
    {
      line=$0
      gsub(/[[:space:]]+$/, "", line)
      sub(/^[[:space:]]+/, "", line)
      if (line ~ /^;/ || line ~ /^#/ || length(line)==0) next
      split(line, kv, "=")
      k=kv[1]; sub(/[[:space:]]+$/, "", k); gsub(/^[[:space:]]+/, "", k)
      if (tolower(k)==tolower(key)) {
        v=substr(line, index(line, "=")+1)
        gsub(/^[[:space:]]+/, "", v); gsub(/[[:space:]]+$/, "", v)
        print v; exit
      }
    }
  ' "$CONFIG"
}

# ---- load config ----
RPCUSER="$(conf_get rpcuser || true)"
RPCPASS="$(conf_get rpcpass || true)"
RPCCERT="$(conf_get rpccert || true)"
NOTLS="$(conf_get notls || true)"

[ -n "$RPCUSER" ] || die "rpcuser missing in $CONFIG"
[ -n "$RPCPASS" ] || die "rpcpass missing in $CONFIG"

# Fixed endpoint used in your previous version
HOSTPORT="127.0.0.1:15213"

# TLS vs no-TLS
SCHEME="https"
CURL_TLS_ARGS=()
if [ -n "$NOTLS" ] && [[ "$NOTLS" =~ ^(1|true|TRUE|yes|YES)$ ]]; then
  SCHEME="http"
else
  if [ -n "$RPCCERT" ]; then
    CURL_TLS_ARGS=(--cacert "$RPCCERT")
  else
    DEFAULT_CERT="${HOME}/.lokid/rpc.cert"
    [ -f "$DEFAULT_CERT" ] && CURL_TLS_ARGS=(--cacert "$DEFAULT_CERT")
  fi
fi

URL="${SCHEME}://${HOSTPORT}/"

rpc() {
  local method="$1" params="${2:-[]}"
  curl -sS --fail --user "${RPCUSER}:${RPCPASS}" \
    -H 'content-type: text/plain' \
    "${CURL_TLS_ARGS[@]}" \
    --data-binary "{\"jsonrpc\":\"1.0\",\"id\":\"hc\",\"method\":\"$method\",\"params\":$params}" \
    "$URL"
}

# ---- single health check ----
info="$(rpc getblockchaininfo)" || die "RPC getblockchaininfo failed"

ibd="$(echo "$info" | jq -r 'if ((.result.initialblockdownload?) | type) == "boolean" then (.result.initialblockdownload|tostring) else "__missing__" end')" || die "Unable to read initialblockdownload flag"
[ "$ibd" = "__missing__" ] && die "initialblockdownload missing from RPC response"
progress="$(echo "$info" | jq -r 'if ((.result.verificationprogress?) | type) == "number" then (.result.verificationprogress|tostring) else "__missing__" end')" || die "Unable to read verificationprogress"
[ "$progress" = "__missing__" ] && die "verificationprogress missing from RPC response"
blocks="$(echo "$info" | jq -r 'if ((.result.blocks?) | type) == "number" then (.result.blocks|tostring) else "__missing__" end')" || die "Unable to read block count"
[ "$blocks" = "__missing__" ] && die "block count missing from RPC response"
headers="$(echo "$info" | jq -r 'if ((.result.headers?) | type) == "number" then (.result.headers|tostring) else "__missing__" end')" || die "Unable to read header count"
[ "$headers" = "__missing__" ] && die "header count missing from RPC response"

# Must be out of IBD
if [ "$ibd" = "true" ]; then
  fail_with_log "waiting… IBD=true"
fi

# Require near-1 verification progress (>= 0.99999)
if ! awk 'BEGIN{exit !('"$progress"' >= 0.99999)}'; then
  fail_with_log "waiting… progress=$progress"
fi

# Strict mode: blocks must equal headers
if [ "$STRICT" = true ] && [ "$blocks" -ne "$headers" ]; then
  fail_with_log "waiting… blocks=$blocks/$headers"
fi

# Need peers: at least one connection
if [ "$NEED_PEERS" = true ]; then
  net="$(rpc getnetworkinfo)" || die "RPC getnetworkinfo failed"
  conns="$(echo "$net" | jq -er '.result.connections')" || die "Unable to read peer connection count"
  if [ "$conns" -lt 1 ]; then
    fail_with_log "waiting… peers=0"
  fi
fi

$QUIET || echo "✅ lokid healthy"
exit 0
