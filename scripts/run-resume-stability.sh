#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

rounds=8
interrupt_seconds=8
root_workers=1
progress_every=1
progress_output="human"
resume_progress="true"
final_timeout_seconds=0
expected_docs=""
checkpoint_template=""
root_folder_ids=""
include_departments=""
department_ids=""

sync_extra_args=()

usage() {
  cat <<'EOF'
用途:
  自动化验证 sync-full 的断点续跑稳定性：
  1) 多轮定时中断 sync-full
  2) 每轮后读取 sync-progress
  3) 最终无中断跑到 done
  4) 对比 Meili 文档数量前后变化（可选校验 expected_docs）

用法:
  bash scripts/run-resume-stability.sh [options] [-- extra sync-full args]

选项:
  --rounds N                  中断轮数，默认 8
  --interrupt-seconds N       每轮运行 N 秒后发送 SIGINT，默认 8
  --root-workers N            sync-full root 并发，默认 1（建议稳定性测试固定为 1）
  --progress-every N          sync-full 进度保存频率，默认 1
  --progress-output MODE      sync-full 进度输出: human|json，默认 human
  --resume-progress BOOL      sync-full 是否断点续跑，默认 true
  --final-timeout-seconds N   最终收敛轮的超时秒数，0 表示不限时，默认 0
  --expected-docs N           最终 Meili numberOfDocuments 期望值（可选）
  --checkpoint-template PATH  透传给 sync-full 的 checkpoint 模板（可选）
  --root-folder-ids CSV       透传给 sync-full（可选）
  --include-departments BOOL  透传给 sync-full（可选）
  --department-ids CSV        透传给 sync-full（可选）
  -h, --help                  查看帮助

环境变量:
  MEILI_HOST / MEILI_API_KEY / MEILI_INDEX
  未设置时会尝试从 .env 读取。

示例:
  bash scripts/run-resume-stability.sh --rounds 6 --interrupt-seconds 10
  bash scripts/run-resume-stability.sh --expected-docs 2046506 -- --token YOUR_TOKEN
EOF
}

log() {
  local now
  now="$(date '+%F %T')"
  echo "[$now] $*"
}

read_dotenv_value() {
  local key="$1"
  local env_file="${2:-.env}"

  if [[ ! -f "$env_file" ]]; then
    return 0
  fi

  local line
  line="$(grep -E "^${key}=" "$env_file" | tail -n 1 || true)"
  if [[ -z "$line" ]]; then
    return 0
  fi

  echo "${line#*=}"
}

extract_json_number() {
  local key="$1"
  local payload="$2"
  local compact
  compact="$(echo "$payload" | tr -d '\n' || true)"
  echo "$compact" \
    | grep -o "\"${key}\":[[:space:]]*[0-9][0-9]*" \
    | head -n 1 \
    | sed -E "s/\"${key}\":[[:space:]]*([0-9][0-9]*)/\\1/"
}

extract_json_string() {
  local key="$1"
  local payload="$2"
  local compact
  compact="$(echo "$payload" | tr -d '\n' || true)"
  echo "$compact" \
    | grep -o "\"${key}\":[[:space:]]*\"[^\"]*\"" \
    | head -n 1 \
    | sed -E "s/\"${key}\":[[:space:]]*\"([^\"]*)\"/\\1/"
}

fetch_meili_stats() {
  local host="$1"
  local index="$2"
  local api_key="$3"
  local url="${host%/}/indexes/${index}/stats"

  if [[ -n "$api_key" ]]; then
    curl -fsS "$url" -H "Authorization: Bearer $api_key"
  else
    curl -fsS "$url"
  fi
}

run_sync_progress() {
  GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go run ./cmd/cli sync-progress
}

build_sync_full_args() {
  local -n target="$1"
  target=(
    go run ./cmd/cli sync-full
    "--resume-progress=${resume_progress}"
    "--root-workers=${root_workers}"
    "--progress-every=${progress_every}"
    "--progress-output=${progress_output}"
  )

  if [[ -n "$checkpoint_template" ]]; then
    target+=("--checkpoint-template=${checkpoint_template}")
  fi
  if [[ -n "$root_folder_ids" ]]; then
    target+=("--root-folder-ids=${root_folder_ids}")
  fi
  if [[ -n "$include_departments" ]]; then
    target+=("--include-departments=${include_departments}")
  fi
  if [[ -n "$department_ids" ]]; then
    target+=("--department-ids=${department_ids}")
  fi
  if [[ ${#sync_extra_args[@]} -gt 0 ]]; then
    target+=("${sync_extra_args[@]}")
  fi
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --rounds)
      rounds="$2"
      shift 2
      ;;
    --interrupt-seconds)
      interrupt_seconds="$2"
      shift 2
      ;;
    --root-workers)
      root_workers="$2"
      shift 2
      ;;
    --progress-every)
      progress_every="$2"
      shift 2
      ;;
    --progress-output)
      progress_output="$2"
      shift 2
      ;;
    --resume-progress)
      resume_progress="$2"
      shift 2
      ;;
    --final-timeout-seconds)
      final_timeout_seconds="$2"
      shift 2
      ;;
    --expected-docs)
      expected_docs="$2"
      shift 2
      ;;
    --checkpoint-template)
      checkpoint_template="$2"
      shift 2
      ;;
    --root-folder-ids)
      root_folder_ids="$2"
      shift 2
      ;;
    --include-departments)
      include_departments="$2"
      shift 2
      ;;
    --department-ids)
      department_ids="$2"
      shift 2
      ;;
    --)
      shift
      sync_extra_args=("$@")
      break
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "未知参数: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if ! [[ "$rounds" =~ ^[0-9]+$ ]]; then
  echo "--rounds 必须是非负整数" >&2
  exit 1
fi
if ! [[ "$interrupt_seconds" =~ ^[0-9]+$ ]] || [[ "$interrupt_seconds" -le 0 ]]; then
  echo "--interrupt-seconds 必须是正整数" >&2
  exit 1
fi
if ! [[ "$root_workers" =~ ^[0-9]+$ ]] || [[ "$root_workers" -le 0 ]]; then
  echo "--root-workers 必须是正整数" >&2
  exit 1
fi
if ! [[ "$progress_every" =~ ^[0-9]+$ ]] || [[ "$progress_every" -le 0 ]]; then
  echo "--progress-every 必须是正整数" >&2
  exit 1
fi
if [[ "$progress_output" != "human" && "$progress_output" != "json" ]]; then
  echo "--progress-output 仅支持 human|json" >&2
  exit 1
fi
if ! [[ "$final_timeout_seconds" =~ ^[0-9]+$ ]]; then
  echo "--final-timeout-seconds 必须是非负整数" >&2
  exit 1
fi
if [[ -n "$expected_docs" ]] && ! [[ "$expected_docs" =~ ^[0-9]+$ ]]; then
  echo "--expected-docs 必须是非负整数" >&2
  exit 1
fi

MEILI_HOST="${MEILI_HOST:-$(read_dotenv_value MEILI_HOST)}"
MEILI_API_KEY="${MEILI_API_KEY:-$(read_dotenv_value MEILI_API_KEY)}"
MEILI_INDEX="${MEILI_INDEX:-$(read_dotenv_value MEILI_INDEX)}"

if [[ -z "$MEILI_HOST" || -z "$MEILI_INDEX" ]]; then
  echo "缺少 MEILI_HOST 或 MEILI_INDEX（环境变量或 .env）" >&2
  exit 1
fi

log "配置: rounds=${rounds}, interrupt_seconds=${interrupt_seconds}, root_workers=${root_workers}, progress_every=${progress_every}, progress_output=${progress_output}"

log "读取 Meili 初始统计..."
stats_before="$(fetch_meili_stats "$MEILI_HOST" "$MEILI_INDEX" "$MEILI_API_KEY")"
docs_before="$(extract_json_number numberOfDocuments "$stats_before")"
if [[ -z "$docs_before" ]]; then
  echo "读取 Meili 初始文档数失败: $stats_before" >&2
  exit 1
fi
log "Meili 初始文档数: ${docs_before}"

declare -a sync_cmd
build_sync_full_args sync_cmd

interrupt_observed=0

for ((i = 1; i <= rounds; i++)); do
  log "第 ${i}/${rounds} 轮: 运行 ${interrupt_seconds}s 后发送 SIGINT"
  set +e
  timeout -s INT "${interrupt_seconds}s" "${sync_cmd[@]}"
  rc=$?
  set -e

  log "第 ${i}/${rounds} 轮退出码: ${rc}"
  if [[ "$rc" -eq 124 || "$rc" -eq 130 ]]; then
    interrupt_observed=1
  elif [[ "$rc" -ne 0 ]]; then
    echo "第 ${i}/${rounds} 轮出现异常退出码: ${rc}" >&2
    exit 1
  fi

  progress_json="$(run_sync_progress || true)"
  if [[ -z "$progress_json" ]]; then
    echo "读取 sync-progress 失败" >&2
    exit 1
  fi

  progress_status="$(extract_json_string status "$progress_json")"
  progress_files="$(extract_json_number filesIndexed "$progress_json")"
  progress_pages="$(extract_json_number pagesFetched "$progress_json")"
  progress_folders="$(extract_json_number foldersVisited "$progress_json")"
  progress_failed="$(extract_json_number failedRequests "$progress_json")"

  log "第 ${i}/${rounds} 轮进度: status=${progress_status:-unknown}, files=${progress_files:-?}, pages=${progress_pages:-?}, folders=${progress_folders:-?}, failed=${progress_failed:-?}"

  if [[ "$progress_status" == "error" ]]; then
    echo "检测到 sync-progress.status=error，测试中止" >&2
    echo "$progress_json" >&2
    exit 1
  fi
done

if [[ "$rounds" -gt 0 && "$interrupt_observed" -eq 0 ]]; then
  echo "警告: 中断轮中未观测到有效中断（可能任务太快完成），本次结果不能代表 resume 抗中断能力。" >&2
  echo "建议增加工作量或先清理/重建测试状态后重试。" >&2
fi

log "开始最终收敛轮（不主动中断）..."
if [[ "$final_timeout_seconds" -gt 0 ]]; then
  timeout "${final_timeout_seconds}s" "${sync_cmd[@]}"
else
  "${sync_cmd[@]}"
fi

final_progress="$(run_sync_progress || true)"
if [[ -z "$final_progress" ]]; then
  echo "读取最终 sync-progress 失败" >&2
  exit 1
fi
final_status="$(extract_json_string status "$final_progress")"
if [[ "$final_status" != "done" ]]; then
  echo "最终状态不是 done: ${final_status}" >&2
  echo "$final_progress" >&2
  exit 1
fi

log "读取 Meili 最终统计..."
stats_after="$(fetch_meili_stats "$MEILI_HOST" "$MEILI_INDEX" "$MEILI_API_KEY")"
docs_after="$(extract_json_number numberOfDocuments "$stats_after")"
if [[ -z "$docs_after" ]]; then
  echo "读取 Meili 最终文档数失败: $stats_after" >&2
  exit 1
fi

delta=$((docs_after - docs_before))
log "Meili 文档数: before=${docs_before}, after=${docs_after}, delta=${delta}"

if [[ -n "$expected_docs" && "$docs_after" -ne "$expected_docs" ]]; then
  echo "最终文档数不匹配: expected=${expected_docs}, actual=${docs_after}" >&2
  exit 1
fi

log "resume 稳定性测试完成: PASS"
