#!/usr/bin/env bash
# OpenTMD 本机配置信息获取脚本
#
# 用法:
#   ./scripts/sysinfo.sh
#   ./scripts/sysinfo.sh --json    # JSON 格式输出
#
# 环境变量:
#   SYSINFO_VERBOSE=1  显示更多硬件详细信息

set -euo pipefail

# ── 格式化辅助 ──────────────────────────────────────────────────
log()    { printf '  %-18s %s\n' "$1" "$2"; }
section(){ printf '\n\033[1;34m%s\033[0m\n' "$1"; }
warn()   { printf '  %-18s \033[33m%s\033[0m\n' "$1" "$*" >&2; }
hr()     { printf '  %-18s\n' '────────────────────'; }

# ── 工具函数 ──────────────────────────────────────────────────
command_exists() { command -v "$1" >/dev/null 2>&1; }

trim() { sed 's/^[[:space:]]*//;s/[[:space:]]*$//'; }

# ── 系统信息采集 ──────────────────────────────────────────────

get_os() {
  if [ -f /etc/os-release ]; then
    . /etc/os-release
    echo "${NAME} ${VERSION_ID} (${VERSION_CODENAME:-unknown})"
  else
    uname -s
  fi
}

get_kernel() {
  uname -r
}

get_arch() {
  uname -m
}

get_hostname() {
  hostname 2>/dev/null || echo "unknown"
}

get_uptime() {
  local uptime_sec
  if [ -f /proc/uptime ]; then
    uptime_sec=$(cut -d' ' -f1 /proc/uptime | cut -d'.' -f1)
  else
    uptime_sec=$(sysctl -n kern.boottime 2>/dev/null | awk '{print $4}' | tr -d ',')
    uptime_sec=$(( $(date +%s) - uptime_sec ))
  fi

  local days=$(( uptime_sec / 86400 ))
  local hours=$(( (uptime_sec % 86400) / 3600 ))
  local mins=$(( (uptime_sec % 3600) / 60 ))

  if [ "$days" -gt 0 ]; then
    echo "${days}d ${hours}h ${mins}m"
  else
    echo "${hours}h ${mins}m"
  fi
}

get_load() {
  local load
  load=$(cat /proc/loadavg 2>/dev/null | awk '{print $1", "$2", "$3}') || \
  load=$(uptime 2>/dev/null | sed 's/.*load averages*://' | trim) || \
  load="N/A"
  echo "$load"
}

# ── CPU 信息 ──────────────────────────────────────────────────

get_cpu_model() {
  if command_exists lscpu; then
    lscpu | grep 'Model name' | cut -d: -f2 | trim
  elif [ -f /proc/cpuinfo ]; then
    grep -m1 'model name' /proc/cpuinfo | cut -d: -f2 | trim
  else
    sysctl -n machdep.cpu.brand_string 2>/dev/null || echo "N/A"
  fi
}

get_cpu_cores() {
  if command_exists lscpu; then
    lscpu | grep '^CPU(s)' | awk '{print $2}'
  elif [ -f /proc/cpuinfo ]; then
    grep -c '^processor' /proc/cpuinfo
  else
    sysctl -n hw.ncpu 2>/dev/null || echo "N/A"
  fi
}

get_cpu_threads_per_core() {
  if command_exists lscpu; then
    lscpu | grep 'Thread(s) per core' | awk '{print $NF}'
  else
    echo "N/A"
  fi
}

get_cpu_min_freq() {
  if command_exists lscpu; then
    lscpu | grep 'CPU min MHz' | awk '{print $NF}'
  fi
}

get_cpu_max_freq() {
  if command_exists lscpu; then
    lscpu | grep -E 'CPU max MHz|CPU MHz' | head -1 | awk '{print $NF}'
  fi
}

get_cpu_cache() {
  if command_exists lscpu; then
    lscpu | grep 'L3 cache' | awk '{print $3, $4}'
  fi
}

get_cpu_arch_x86() {
  if command_exists lscpu; then
    local flags
    flags=$(lscpu | grep Flags | head -1)
    local result=""
    echo "$flags" | grep -q 'sse4_1' && result="${result}SSE4.1 "
    echo "$flags" | grep -q 'sse4_2' && result="${result}SSE4.2 "
    echo "$flags" | grep -q 'avx2'   && result="${result}AVX2 "
    echo "$flags" | grep -q 'avx512' && result="${result}AVX-512 "
    echo "$flags" | grep -q 'aes'    && result="${result}AES-NI "
    echo "$flags" | grep -q 'vmx'    && result="${result}VT-x"
    echo "$flags" | grep -q 'svm'    && result="${result}AMD-V"
    [ -n "$result" ] && echo "$result" || echo "N/A"
  fi
}

get_cpu_arch_arm() {
  if command_exists lscpu; then
    local flags
    flags=$(lscpu | grep Flags | head -1)
    local result=""
    echo "$flags" | grep -q 'aes' && result="${result}AES "
    echo "$flags" | grep -q 'neon' && result="${result}NEON "
    echo "$flags" | grep -q 'fp' && result="${result}FP "
    echo "$flags" | grep -q 'asimd' && result="${result}ASIMD"
    [ -n "$result" ] && echo "$result" || echo "N/A"
  fi
}

# ── 内存信息 ──────────────────────────────────────────────────

get_memory() {
  if command_exists free; then
    free -h | awk '/^Mem:/{print $2, "total,", $3, "used,", $4, "avail,", $7, "avail(cache)"}'
  elif command_exists vm_stat; then
    local page_size page_count
    page_size=$(vm_stat | awk '/page size/{print $8}')
    page_count=$(vm_stat | awk '/free/{print $3}' | sed 's/\.//')
    echo "$(( page_size * page_count / 1024 / 1024 / 1024 ))G free"
  else
    echo "N/A"
  fi
}

get_swap() {
  if command_exists free; then
    free -h | awk '/^Swap:/{print $2, "total,", $3, "used"}'
  elif command_exists swapctl; then
    swapctl -l | awk 'NR>1{printf "%s total, %s used", $2/1024/1024"G", $3/1024/1024"G"}'
  else
    echo "N/A"
  fi
}

get_memory_detail() {
  if command_exists dmidecode && [ "${SYSINFO_VERBOSE:-0}" = "1" ]; then
    dmidecode -t memory 2>/dev/null | grep -E '^\s+(Size|Type|Speed|Manufacturer):' | head -20
  fi
}

# ── 磁盘信息 ──────────────────────────────────────────────────

get_disk_info() {
  local cmd
  if command_exists df; then
    cmd="df -h --total 2>/dev/null" || cmd="df -h"
    $cmd | awk 'NR==1{next} /^\/dev/ || /^total/{printf "%s  %s  %s  %s  %s  %s\n", $1, $2, $3, $4, $5, $6}'
  else
    echo "N/A"
  fi
}

get_disk_summary() {
  if command_exists df; then
    df -h --total 2>/dev/null | awk '/total/{printf "%s total, %s used, %s avail (%s)", $3, $4, $5, $6}'
  elif command_exists df; then
    df -h / | awk 'NR>1{printf "%s total, %s used, %s avail", $2, $3, $4}'
  fi
}

# ── GPU 信息 ──────────────────────────────────────────────────

get_gpu_info() {
  if command_exists lspci; then
    lspci 2>/dev/null | grep -iE 'vga|3d|display' | sed 's/^[^ ]* //' || echo "N/A"
  elif command_exists system_profiler; then
    system_profiler SPDisplaysDataType 2>/dev/null | grep -E 'Chipset|VRAM|Vendor' | trim
  else
    echo "N/A"
  fi
}

# ── 网络信息 ──────────────────────────────────────────────────

get_network_info() {
  local ip4 ip6
  ip4=$(ip -4 addr show 2>/dev/null | grep 'inet ' | awk '{print $2}' | grep -v '^127\.' | head -3 | tr '\n' ' ') || \
  ip4=$(ifconfig 2>/dev/null | grep 'inet ' | awk '{print $2}' | grep -v '^127\.' | head -3 | tr '\n' ' ')
  ip6=$(ip -6 addr show 2>/dev/null | grep 'inet6 ' | awk '{print $2}' | grep -v '^::1' | head -3 | tr '\n' ' ') || \
  ip6=$(ifconfig 2>/dev/null | grep 'inet6 ' | awk '{print $2}' | grep -v '^::1' | head -3 | tr '\n' ' ')
  echo "IPv4: ${ip4:-N/A}"
  echo "IPv6: ${ip6:-N/A}"
}

# ── Shell / 终端 ──────────────────────────────────────────────

get_shell_info() {
  echo "Shell: ${SHELL:-N/A}"
  echo "Term:  ${TERM:-N/A}"
}

# ── 开发工具 ──────────────────────────────────────────────────

get_dev_tools() {
  local tools=("go" "node" "npm" "python3" "rustc" "cargo" "docker" "git" "make" "cmake" "gcc" "clang")
  for tool in "${tools[@]}"; do
    if command_exists "$tool"; then
      local ver
        ver=$("$tool" --version 2>/dev/null | head -1 | sed 's/^[^0-9]*//' | trim | cut -d' ' -f1)
      printf "  %-10s %s\n" "$tool" "$ver"
    fi
  done
}

# ── JSON 输出模式 ──────────────────────────────────────────

output_json() {
  local os=$(get_os)
  local kernel=$(get_kernel)
  local arch=$(get_arch)
  local hostname=$(get_hostname)
  local uptime=$(get_uptime)
  local load=$(get_load)
  local cpu_model=$(get_cpu_model)
  local cpu_cores=$(get_cpu_cores)
  local mem=$(get_memory)
  local disk=$(get_disk_summary)
  local gpu=$(get_gpu_info)

  cat <<EOF
{
  "hostname": "$hostname",
  "os": "$os",
  "kernel": "$kernel",
  "arch": "$arch",
  "uptime": "$uptime",
  "load": "$load",
  "cpu": {
    "model": "$cpu_model",
    "cores": $cpu_cores
  },
  "memory": "$mem",
  "disk": "$disk",
  "gpu": "$gpu"
}
EOF
  exit 0
}

# ── 主程序 ──────────────────────────────────────────────────

main() {
  # 检测 JSON 模式
  for arg in "$@"; do
    [ "$arg" = "--json" ] && output_json
  done

  echo ""
  echo "╔══════════════════════════════════════════╗"
  echo "║      OpenTMD — 本机配置信息              ║"
  echo "╚══════════════════════════════════════════╝"
  echo ""

  # ── 系统概览 ──
  section "■ 系统概览"
  log "主机名"     "$(get_hostname)"
  log "操作系统"   "$(get_os)"
  log "内核版本"   "$(get_kernel)"
  log "架构"       "$(get_arch)"
  log "运行时间"   "$(get_uptime)"
  log "负载"       "$(get_load)"

  # ── CPU ──
  section "■ CPU"
  log "型号"       "$(get_cpu_model)"
  log "核心数"     "$(get_cpu_cores) 核"
  if [ "${SYSINFO_VERBOSE:-0}" = "1" ]; then
    local threads_per_core
    threads_per_core=$(get_cpu_threads_per_core)
    [ -n "$threads_per_core" ] && log "每核线程"   "$threads_per_core"
    local min_freq max_freq
    min_freq=$(get_cpu_min_freq)
    max_freq=$(get_cpu_max_freq)
    [ -n "$min_freq" ] && log "主频范围"   "${min_freq} ~ ${max_freq} MHz"
    local cache
    cache=$(get_cpu_cache)
    [ -n "$cache" ] && log "L3 缓存"     "$cache"
  fi
  local arch_ext
  arch_ext=$(get_cpu_arch_x86)
  [ "$arch_ext" != "N/A" ] && log "指令集"      "$arch_ext"

  # ── 内存 ──
  section "■ 内存"
  log "内存"       "$(get_memory)"
  log "交换分区"   "$(get_swap)"
  if [ "${SYSINFO_VERBOSE:-0}" = "1" ]; then
    local mem_detail
    mem_detail=$(get_memory_detail)
    [ -n "$mem_detail" ] && echo "$mem_detail" | while IFS= read -r line; do
      log " " "$line"
    done
  fi

  # ── 磁盘 ──
  section "■ 磁盘"
  log "概要"       "$(get_disk_summary)"
  echo ""
  get_disk_info | while IFS= read -r line; do
    printf "  %-10s %s\n" "" "$line"
  done

  # ── GPU ──
  section "■ GPU"
  log "显卡"       "$(get_gpu_info | tr '\n' ' ')"

  # ── 网络 ──
  section "■ 网络"
  get_network_info | while IFS= read -r line; do
    log "" "$line"
  done

  # ── Shell ──
  section "■ Shell"
  get_shell_info | while IFS= read -r line; do
    log "" "$line"
  done

  # ── 开发工具 ──
  section "■ 开发工具"
  get_dev_tools

  echo ""
}

main "$@"
