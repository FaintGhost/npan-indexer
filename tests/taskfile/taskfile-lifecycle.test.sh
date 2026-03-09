#!/usr/bin/env bash
set -u -o pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
taskfile="$repo_root/Taskfile.yml"

failures=0

pass() {
  printf 'PASS  %s\n' "$1"
}

fail() {
  printf 'FAIL  %s\n' "$1"
  failures=$((failures + 1))
}

extract_task_block() {
  local task_name="$1"

  [[ -f "$taskfile" ]] || return 1

  awk -v task="$task_name" '
    BEGIN {
      in_task = 0
    }

    $0 ~ "^[[:space:]]{2}" task ":[[:space:]]*$" {
      in_task = 1
      print
      next
    }

    in_task && $0 ~ "^[[:space:]]{2}[A-Za-z0-9:_-]+:[[:space:]]*$" {
      exit
    }

    in_task {
      print
    }
  ' "$taskfile"
}

assert_task_declared() {
  local task_name="$1"
  local block

  block="$(extract_task_block "$task_name")"
  if [[ -n "$block" ]]; then
    pass "已声明生命周期任务 $task_name"
  else
    fail "缺少生命周期任务 $task_name"
  fi
}

assert_task_contains() {
  local task_name="$1"
  local needle="$2"
  local description="$3"
  local block

  block="$(extract_task_block "$task_name")"
  if [[ -z "$block" ]]; then
    fail "$description（原因：任务 $task_name 缺失）"
    return
  fi

  if grep -Fq -- "$needle" <<<"$block"; then
    pass "$description"
  else
    fail "$description（缺少文本：$needle）"
  fi
}

assert_task_not_contains() {
  local task_name="$1"
  local needle="$2"
  local description="$3"
  local block

  block="$(extract_task_block "$task_name")"
  if [[ -z "$block" ]]; then
    fail "$description（原因：任务 $task_name 缺失）"
    return
  fi

  if grep -Fq -- "$needle" <<<"$block"; then
    fail "$description（检测到不应出现的文本：$needle）"
  else
    pass "$description"
  fi
}

assert_task_order() {
  local task_name="$1"
  local first="$2"
  local second="$3"
  local description="$4"
  local block
  local first_line
  local second_line

  block="$(extract_task_block "$task_name")"
  if [[ -z "$block" ]]; then
    fail "$description（原因：任务 $task_name 缺失）"
    return
  fi

  first_line="$(grep -Fn -- "$first" <<<"$block" | head -n 1 | cut -d: -f1)"
  second_line="$(grep -Fn -- "$second" <<<"$block" | head -n 1 | cut -d: -f1)"

  if [[ -z "$first_line" || -z "$second_line" ]]; then
    fail "$description（缺少顺序锚点：$first / $second）"
    return
  fi

  if (( first_line < second_line )); then
    pass "$description"
  else
    fail "$description（当前顺序不满足：$first 应先于 $second）"
  fi
}

print_summary() {
  if (( failures == 0 )); then
    printf '\nTask 生命周期验证通过：%d 个断言全部满足。\n' 0
  else
    printf '\nTask 生命周期 RED 验证失败：%d 个断言未满足。\n' "$failures"
    printf '说明：当前仓库尚未完整提供 stack:ci:* / smoke:run / e2e:run / verify:* 生命周期任务语义。\n'
  fi
}

printf 'Running Task lifecycle RED verification against %s\n\n' "$taskfile"

if [[ ! -f "$taskfile" ]]; then
  fail "缺少根级 Taskfile.yml，当前尚未暴露 Task 生命周期入口。"
  fail "缺少生命周期任务集合：stack:ci:up、stack:ci:down、stack:ci:logs、smoke:run、e2e:run、verify:smoke、verify:e2e、verify:all。"
  fail "verify:smoke 语义缺失：未能证明 cold-start compose up、默认 BASE_URL=http://localhost:11323、默认 METRICS_URL=http://localhost:19091、执行 smoke_test.sh、并在结束时 down --volumes。"
  fail "verify:e2e 语义缺失：未能证明 cold-start compose up、执行 Playwright 容器验证、并在结束时 down --volumes。"
  fail "verify:all 语义缺失：未能证明先执行 verify:quick、再执行 verify:e2e，且任一步失败返回非零。"
  print_summary
  exit 1
fi

assert_task_declared "stack:ci:up"
assert_task_declared "stack:ci:down"
assert_task_declared "stack:ci:logs"
assert_task_declared "smoke:run"
assert_task_declared "e2e:run"
assert_task_declared "verify:smoke"
assert_task_declared "verify:e2e"
assert_task_declared "verify:all"

assert_task_contains "stack:ci:up" "docker-compose.ci.yml" "stack:ci:up 使用 CI compose 文件"
assert_task_contains "stack:ci:up" "up --build -d --wait --wait-timeout 120" "stack:ci:up 具备 cold-start 等待语义"
assert_task_contains "stack:ci:down" "docker-compose.ci.yml" "stack:ci:down 使用 CI compose 文件"
assert_task_contains "stack:ci:down" "down --volumes" "stack:ci:down 包含卷清理语义"
assert_task_contains "stack:ci:logs" "docker-compose.ci.yml" "stack:ci:logs 使用 CI compose 文件"
assert_task_contains "stack:ci:logs" "logs" "stack:ci:logs 暴露日志采集语义"

assert_task_contains "smoke:run" "tests/smoke/smoke_test.sh" "smoke:run 调用 smoke 验证脚本"
assert_task_contains "smoke:run" "BASE_URL=http://localhost:11323" "smoke:run 声明默认 BASE_URL"
assert_task_contains "smoke:run" "METRICS_URL=http://localhost:19091" "smoke:run 声明默认 METRICS_URL"

assert_task_contains "e2e:run" "--profile e2e" "e2e:run 使用 Playwright profile"
assert_task_contains "e2e:run" "playwright" "e2e:run 声明 Playwright 容器链路"

assert_task_contains "verify:smoke" "task: stack:ci:up" "verify:smoke 依次编排 stack:ci:up"
assert_task_contains "verify:smoke" "task: smoke:run" "verify:smoke 依次编排 smoke:run"
assert_task_contains "verify:smoke" "task: stack:ci:down" "verify:smoke 包含清理任务"
assert_task_contains "verify:smoke" "defer:" "verify:smoke 声明失败也清理的 defer 语义"
assert_task_order "verify:smoke" "task: stack:ci:up" "task: smoke:run" "verify:smoke 先启动环境再运行 smoke"

assert_task_contains "verify:e2e" "task: stack:ci:up" "verify:e2e 依次编排 stack:ci:up"
assert_task_contains "verify:e2e" "task: smoke:run" "verify:e2e 依次编排 smoke:run"
assert_task_contains "verify:e2e" "task: e2e:run" "verify:e2e 依次编排 e2e:run"
assert_task_contains "verify:e2e" "task: stack:ci:down" "verify:e2e 包含清理任务"
assert_task_contains "verify:e2e" "defer:" "verify:e2e 声明失败也清理的 defer 语义"
assert_task_order "verify:e2e" "task: stack:ci:up" "task: smoke:run" "verify:e2e 先启动环境再运行 smoke"
assert_task_order "verify:e2e" "task: smoke:run" "task: e2e:run" "verify:e2e 先执行 smoke 再运行 Playwright"

assert_task_contains "verify:all" "task: verify:quick" "verify:all 串联快速验证入口"
assert_task_contains "verify:all" "task: verify:e2e" "verify:all 串联完整 E2E 入口"
assert_task_order "verify:all" "task: verify:quick" "task: verify:e2e" "verify:all 先执行 verify:quick 再执行 verify:e2e"
assert_task_not_contains "verify:all" "deps:" "verify:all 不使用 deps 表达顺序语义"

print_summary

if (( failures > 0 )); then
  exit 1
fi

exit 0
