#!/usr/bin/env bash
set -u
set -o pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

required_tasks=(
  "guard:rest"
  "test:go"
  "test:web"
  "verify:quick"
  "verify:smoke"
  "verify:e2e"
)

failures=()

if [[ ! -f "$REPO_ROOT/Taskfile.yml" ]]; then
  failures+=("仓库根目录缺少 Taskfile.yml，因此无法建立 namespaced Task 公开任务表面。")
fi

list_output="$(task --dir "$REPO_ROOT" --list 2>&1)"
list_status=$?

if (( list_status != 0 )); then
  failures+=("执行 'task --list' 失败，无法列出公开任务表面。原始输出: $list_output")
fi

missing_tasks=()
for task_name in "${required_tasks[@]}"; do
  if [[ "$list_output" != *"$task_name"* ]]; then
    missing_tasks+=("$task_name")
  fi
done

if (( ${#missing_tasks[@]} > 0 )); then
  missing_tasks_text="$(printf '%s, ' "${missing_tasks[@]}")"
  missing_tasks_text="${missing_tasks_text%, }"
  failures+=("'task --list' 缺少期望公开任务: $missing_tasks_text")
fi

quick_output="$(task --dir "$REPO_ROOT" --summary verify:quick 2>&1)"
quick_status=$?

if (( quick_status != 0 )); then
  failures+=("无法发现 verify:quick 聚合入口。原始输出: $quick_output")
fi

if (( ${#failures[@]} > 0 )); then
  printf 'FAIL: Task 迁移的公开任务表面尚未就绪。\n'
  for failure in "${failures[@]}"; do
    printf '  - %s\n' "$failure"
  done
  exit 1
fi

printf 'PASS: namespaced Task 公开任务表面与 verify:quick 聚合入口已就绪。\n'
