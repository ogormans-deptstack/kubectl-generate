#!/usr/bin/env bash
# demo.sh - Interactive demo of kubectl-generate
# Usage: ./demo.sh
# Screen-record your terminal while this runs (Cmd+Shift+5 on macOS)
#
# Requires:
#   - kubectl-generate installed (make install)
#   - A running Kubernetes cluster (kind create cluster --name demo)
#   - Gateway API CRDs installed for HTTPRoute demo

set -uo pipefail

# Colors
CYAN='\033[0;36m'
GREEN='\033[0;32m'
NC='\033[0m'
PROMPT='\033[1;32m$\033[0m '

# Simulate typing a command, then run it
# Usage: run "display command" ["actual command"]
# If only one arg, display and actual are the same.
run() {
	local display="$1"
	local actual="${2:-$1}"
	echo ""
	echo -ne "${PROMPT}"
	for ((i = 0; i < ${#display}; i++)); do
		echo -n "${display:$i:1}"
		sleep 0.04
	done
	sleep 0.5
	echo ""
	eval "$actual"
	sleep 2
}

# Print a comment/header
comment() {
	echo ""
	echo -e "${CYAN}# $1${NC}"
	sleep 1
}

clear

comment "kubectl-generate: schema-driven YAML from your cluster"
sleep 1

comment "List available resource types (62 and counting)"
run "kubectl generate --list | head -20" "kubectl generate --list | grep -v 'List$' | grep -v '^API' | grep -vE '^(Binding|Status|DeleteOptions|Scale|SelfSubject|SubjectAccessReview|LocalSubjectAccessReview|ComponentStatus|Event|TokenRequest|TokenReview|WatchEvent|Eviction)$' | head -20"

comment "Generate a Deployment with overrides"
run "kubectl generate Deployment --name=web --image=myapp:v2 --replicas=3"

comment "Typo? Fuzzy matching suggests the right type"
run "kubectl generate Deploymnet || true"

comment "Generate a CronJob"
run "kubectl generate CronJob --name=backup"

comment "Generate a CRD (Gateway API HTTPRoute)"
run "kubectl generate HTTPRoute --name=api"

comment "Pipe directly to kubectl apply for validation"
run "kubectl generate Service --name=web | kubectl apply --dry-run=server -f -"

echo ""
echo -e "${GREEN}Done! Install: kubectl krew install generate${NC}"
sleep 3
