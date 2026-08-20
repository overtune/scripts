#!/bin/bash
# @meta
# name: procwatch
# description: Find CPU/memory-hungry and detached dev processes with their working directories, and kill them
# category: sys
# args:
#   - name: mode
#     required: false
#     help: hot (default) | orphans | all | kill
#   - name: target
#     required: false
#     help: PID(s) to kill in kill mode; row limit in hot/orphans mode
# @end

set -uo pipefail

# ps renders %cpu with the locale's decimal separator; force C so awk and sort
# always see a dot.
export LC_ALL=C

CPU_MIN=${PROCWATCH_CPU_MIN:-20}
RSS_MIN_MB=${PROCWATCH_RSS_MIN_MB:-500}
HOT_LIMIT=${PROCWATCH_TOP:-10}
ORPHAN_LIMIT=25

SELF_PID=$$
SELF_PPID=$PPID
SELF_UID=$(id -u)

# Dev tools worth flagging when they end up detached. Matched against the
# executable's basename, exactly.
DEV_RE='^(node|npm|npx|pnpm|yarn|bun|deno|ts-node|tsx|nodemon|vite|esbuild|webpack|next|rollup|parcel|turbo|jest|vitest|python|python2|python3|ruby|php|java|gradle|gradlew|mvn|dotnet|cargo|rustc|go|air|dlv|dart|flutter|docker-compose|http-server|serve|ng|gulp|grunt)$'

usage() {
	cat <<'EOF'
procwatch — find hungry and detached dev processes, with their working directories

  procwatch [hot] [limit]      top processes by CPU and by memory (default)
  procwatch orphans [limit]    dev processes with no terminal / reparented to launchd
  procwatch all [limit]        both of the above
  procwatch kill <pid...>      terminate the given PIDs (comma or space separated)

Env: PROCWATCH_CPU_MIN (default 20), PROCWATCH_RSS_MIN_MB (default 500),
     PROCWATCH_TOP (default 10). Rows crossing a threshold are marked with *.
EOF
}

# ps_snapshot emits one TAB-separated row per process:
#   pid  ppid  cpu  rss_mb  tty  etime  user  comm
ps_snapshot() {
	ps -Ao pid=,ppid=,%cpu=,rss=,tty=,etime=,user=,comm= | awk '
		NF >= 8 {
			comm = ""
			for (i = 8; i <= NF; i++) comm = comm (i > 8 ? " " : "") $i
			printf "%s\t%s\t%s\t%.0f\t%s\t%s\t%s\t%s\n", $1, $2, $3, $4 / 1024, $5, $6, $7, comm
		}'
}

# cwd_map reads PIDs on stdin (one per line) and emits "pid<TAB>cwd" for those
# lsof can resolve, in a single lsof call.
cwd_map() {
	local pids
	pids=$(paste -sd, -)
	[ -n "$pids" ] || return 0
	lsof -w -a -p "$pids" -d cwd -Fn 2>/dev/null | awk '
		/^p/ { pid = substr($0, 2); next }
		/^n/ { if (pid != "") { print pid "\t" substr($0, 2); pid = "" } }'
}

# render prints a table for the TAB-separated rows passed as $1, resolving each
# row's working directory. Rows and the cwd map are merged into one stream with
# a leading type column so awk needs no temp files.
render() {
	local rows=$1
	if [ -z "$rows" ]; then
		echo "  (none)"
		return
	fi
	{
		printf '%s\n' "$rows" | cut -f1 | cwd_map | awk '{ print "C\t" $0 }'
		printf '%s\n' "$rows" | awk '{ print "R\t" $0 }'
	} | awk -F'\t' -v cpumin="$CPU_MIN" -v rssmin="$RSS_MIN_MB" '
		BEGIN {
			printf "%-2s %-7s %-7s %7s %7s %-6s %-11s %-22s %s\n", \
				"", "PID", "PPID", "CPU%", "MEM", "TTY", "UPTIME", "COMMAND", "CWD"
		}
		$1 == "C" { cwd[$2] = $3; next }
		$1 == "R" {
			pid = $2; cpu = $4; mem = $5; comm = $9
			n = split(comm, parts, "/")
			mark = (cpu + 0 >= cpumin + 0 || mem + 0 >= rssmin + 0) ? "*" : ""
			dir = (pid in cwd) ? cwd[pid] : "-"
			printf "%-2s %-7s %-7s %7.1f %6dM %-6s %-11s %-22s %s\n", \
				mark, pid, $3, cpu, mem, $6, $7, substr(parts[n], 1, 22), dir
		}'
}

mode_hot() {
	local limit=$1 snap
	snap=$(ps_snapshot)
	echo "Top $limit by CPU (* = CPU >= ${CPU_MIN}% or RSS >= ${RSS_MIN_MB}M)"
	render "$(printf '%s\n' "$snap" | sort -t$'\t' -k3,3 -gr | head -n "$limit")"
	echo
	echo "Top $limit by memory"
	render "$(printf '%s\n' "$snap" | sort -t$'\t' -k4,4 -gr | head -n "$limit")"
}

mode_orphans() {
	local limit=$1 rows
	rows=$(ps_snapshot | awk -F'\t' \
		-v devre="$DEV_RE" -v myuid="$SELF_UID" \
		-v self="$SELF_PID" -v selfppid="$SELF_PPID" '
		{
			pid = $1; ppid = $2; tty = $5; user = $7; comm = $8
			if (pid == self || pid == selfppid) next
			# Only our own processes: killing anyone else needs sudo anyway.
			if (user != ENVIRON["PROCWATCH_USER"]) next
			# Detached: no controlling terminal, or reparented to launchd.
			if (tty != "??" && ppid != 1) next
			# GUI app helpers and system daemons are not stray dev servers.
			if (comm ~ /\.app\/Contents\//) next
			if (comm ~ /^\/System\// || comm ~ /^\/usr\/libexec\// || comm ~ /^\/Library\/Apple\//) next
			n = split(comm, parts, "/")
			if (parts[n] !~ devre) next
			print
		}' | sort -t$'\t' -k3,3 -gr | head -n "$limit")

	echo "Detached dev processes (candidates, not a verdict — check CWD before killing)"
	render "$rows"
}

# resolve_targets validates the raw PID list in $1, printing one PID per line.
# Refusals and unknown PIDs go to stderr; a bad token fails the whole run.
resolve_targets() {
	local raw=$1 tok pid uid ok=0
	for tok in $(printf '%s' "$raw" | tr ',' ' '); do
		case $tok in
		'' | *[!0-9]*)
			echo "not a PID: $tok" >&2
			return 1
			;;
		esac
		pid=$((10#$tok))
		if [ "$pid" -le 1 ]; then
			echo "refusing PID $pid" >&2
			continue
		fi
		if [ "$pid" -eq "$SELF_PID" ] || [ "$pid" -eq "$SELF_PPID" ]; then
			echo "refusing PID $pid (that's procwatch or its parent)" >&2
			continue
		fi
		uid=$(ps -o uid= -p "$pid" 2>/dev/null | tr -d ' ')
		if [ -z "$uid" ]; then
			echo "no such process: $pid" >&2
			continue
		fi
		if [ "$uid" != "$SELF_UID" ]; then
			echo "refusing PID $pid (owned by uid $uid, not you)" >&2
			continue
		fi
		echo "$pid"
		ok=1
	done
	[ "$ok" -eq 1 ]
}

mode_kill() {
	local raw=$1 pids pid cmd dir waited alive=0
	if [ -z "$raw" ]; then
		echo "kill mode needs at least one PID" >&2
		usage >&2
		return 1
	fi

	pids=$(resolve_targets "$raw") || return 1

	echo "About to terminate:"
	while IFS= read -r pid; do
		dir=$(printf '%s\n' "$pid" | cwd_map | cut -f2)
		cmd=$(ps -o command= -p "$pid" 2>/dev/null)
		printf '  %-7s cwd: %s\n' "$pid" "${dir:--}"
		printf '          cmd: %s\n' "${cmd:-<gone>}"
	done <<EOF
$pids
EOF

	if [ -t 0 ]; then
		local reply
		printf 'Kill %s process(es)? [y/N] ' "$(printf '%s\n' "$pids" | wc -l | tr -d ' ')"
		read -r reply
		case $reply in
		y | Y | yes | YES) ;;
		*)
			echo "aborted"
			return 1
			;;
		esac
	fi

	while IFS= read -r pid; do
		if ! kill "$pid" 2>/dev/null; then
			echo "$pid: could not send TERM (already gone?)"
			continue
		fi
		waited=0
		while [ "$waited" -lt 15 ] && kill -0 "$pid" 2>/dev/null; do
			sleep 0.2
			waited=$((waited + 1))
		done
		if kill -0 "$pid" 2>/dev/null; then
			echo "$pid: survived TERM after 3s — force with: kill -9 $pid"
			alive=1
		else
			echo "$pid: terminated"
		fi
	done <<EOF
$pids
EOF

	return "$alive"
}

# check_limit validates an optional row-limit argument.
check_limit() {
	local raw=$1 fallback=$2
	if [ -z "$raw" ]; then
		echo "$fallback"
		return 0
	fi
	case $raw in
	'' | *[!0-9]*)
		echo "limit must be a number: $raw" >&2
		return 1
		;;
	esac
	if [ "$raw" -lt 1 ] || [ "$raw" -gt 200 ]; then
		echo "limit must be between 1 and 200: $raw" >&2
		return 1
	fi
	echo "$raw"
}

MODE=${1:-}
[ -n "$MODE" ] || MODE=hot
TARGET=${2:-}

# Exported for the awk owner check in mode_orphans, where ps prints names.
PROCWATCH_USER=$(id -un)
export PROCWATCH_USER

case $MODE in
hot)
	limit=$(check_limit "$TARGET" "$HOT_LIMIT") || exit 1
	mode_hot "$limit"
	;;
orphans)
	limit=$(check_limit "$TARGET" "$ORPHAN_LIMIT") || exit 1
	mode_orphans "$limit"
	;;
all)
	limit=$(check_limit "$TARGET" "$HOT_LIMIT") || exit 1
	mode_hot "$limit"
	echo
	mode_orphans "$ORPHAN_LIMIT"
	;;
kill)
	mode_kill "$TARGET"
	;;
-h | --help | help)
	usage
	;;
*)
	echo "unknown mode: $MODE" >&2
	usage >&2
	exit 1
	;;
esac
