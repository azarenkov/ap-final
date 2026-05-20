#!/usr/bin/env bash
# Seeds the train tickets system with routes + trains through the API gateway.
# Idempotent: dup `code` returns from train-service are harmless. Works under macOS bash 3.x.

set -e -o pipefail

API="${API:-http://localhost:8080}"
EMAIL="${SEED_EMAIL:-seed@trains.local}"
PASSWORD="${SEED_PASSWORD:-password123}"
NAME="${SEED_NAME:-Seed User}"

red()   { printf "\033[31m%s\033[0m\n" "$*"; }
green() { printf "\033[32m%s\033[0m\n" "$*"; }
gray()  { printf "\033[2m%s\033[0m\n" "$*"; }

curl_json() { curl -sS -H 'Content-Type: application/json' "$@"; }

extract() { python3 -c "import sys,json
try:
  d=json.load(sys.stdin); print(d.get('$1',''))
except Exception:
  pass"; }

gray "› registering seed user $EMAIL"
curl_json -X POST "$API/v1/users/register" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\",\"full_name\":\"$NAME\"}" >/dev/null 2>&1 || true

TOKEN="$(curl_json -X POST "$API/v1/users/login" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" | extract access_token)"
[[ -n "$TOKEN" ]] || { red "× login failed"; exit 1; }
green "✓ logged in, token ${TOKEN:0:14}…"

# Build a temp file with route slug → id mapping so we can look them up later.
ROUTES_FILE="$(mktemp)"
trap 'rm -f "$ROUTES_FILE"' EXIT

create_route() {
    local slug="$1" origin="$2" destination="$3" km="$4" minutes="$5"
    local body
    body=$(printf '{"origin":"%s","destination":"%s","distance_km":%d,"estimated_minutes":%d}' \
        "$origin" "$destination" "$km" "$minutes")
    local res id
    res="$(curl_json -X POST "$API/v1/routes" -H "Authorization: Bearer $TOKEN" -d "$body")"
    id="$(printf "%s" "$res" | extract id)"
    if [[ -n "$id" ]]; then
        printf "%s %s\n" "$slug" "$id" >> "$ROUTES_FILE"
        green "✓ route $slug → ${id:0:8}…"
    else
        gray "› route $slug failed: $res"
    fi
}

route_id() {
    awk -v s="$1" '$1==s {print $2; exit}' "$ROUTES_FILE"
}

in_hours() {
    # ISO timestamp at "tomorrow + h" hours. Accepts hours > 23 so callers can express
    # overnight services (e.g. dep=21, arr=33 → next-day 09:00) without juggling dates.
    local h="$1"
    local days=$(( h / 24 ))
    local hour=$(( h % 24 ))
    local base
    base="$(date -u -v+$((days + 1))d "+%Y-%m-%d" 2>/dev/null || date -u -d "+$((days + 1)) day" "+%Y-%m-%d")"
    printf "%sT%02d:00:00Z" "$base" "$hour"
}

create_train() {
    local code="$1" name="$2" slug="$3" dep_h="$4" arr_h="$5" seats="$6" price="$7"
    local rid
    rid="$(route_id "$slug")"
    if [[ -z "$rid" ]]; then gray "› skip $code (route $slug missing)"; return; fi
    local dep arr body res id
    dep="$(in_hours "$dep_h")"
    arr="$(in_hours "$arr_h")"
    body=$(printf '{"code":"%s","name":"%s","route_id":"%s","departure_time":"%s","arrival_time":"%s","total_seats":%d,"price_cents":%d}' \
        "$code" "$name" "$rid" "$dep" "$arr" "$seats" "$price")
    res="$(curl_json -X POST "$API/v1/trains" -H "Authorization: Bearer $TOKEN" -d "$body")"
    id="$(printf "%s" "$res" | extract id)"
    if [[ -n "$id" ]]; then
        green "✓ train $code → ${id:0:8}…"
    else
        gray "› train $code returned: $res"
    fi
}

# ----- routes -----
create_route "AST-ALA" "Astana"   "Almaty"    1200 960
create_route "ALA-CHY" "Almaty"   "Shymkent"   700 540
create_route "CHY-AKT" "Shymkent" "Aktobe"    1100 840
create_route "ATY-AKB" "Atyrau"   "Aktobe"     520 420
create_route "ALA-AST" "Almaty"   "Astana"    1200 960

# ----- trains -----
# Astana → Almaty (arrival hours > 23 = next-day arrivals)
create_train "IC-101" "Talgo Continental" "AST-ALA"  8 22 300 1500000
create_train "IC-103" "Tulpar Express"    "AST-ALA" 13 27 250 1200000
create_train "IC-105" "Night Sapsan"      "AST-ALA" 21 35 220 1800000

# Almaty → Shymkent
create_train "IC-201" "Spanish"           "ALA-CHY"  6 15 180  800000
create_train "IC-203" "Continental"       "ALA-CHY" 18 27 180  850000

# Shymkent → Aktobe
create_train "IC-301" "Steppe Express"    "CHY-AKT"  9 23 250 1100000

# Atyrau → Aktobe
create_train "IC-401" "Caspian"           "ATY-AKB" 14 21 200  650000

# Almaty → Astana (return)
create_train "IC-901" "Talgo Return"      "ALA-AST"  7 21 300 1500000
create_train "IC-903" "Sapsan Return"     "ALA-AST" 15 29 250 1300000

echo
green "Seed complete."
echo "Search at http://localhost:3000  (try Astana → Almaty)"
