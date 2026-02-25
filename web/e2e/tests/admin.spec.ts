import { test, expect, ADMIN_API_KEY } from "../fixtures/auth";
import { AdminPage } from "../pages/admin-page";

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function isAdminMethodRequest(
  url: string,
  methodName:
    | "StartSync"
    | "CancelSync"
    | "WatchSyncProgress"
    | "GetIndexStats"
    | "InspectRoots",
): boolean {
  return url.includes(`/npan.v1.AdminService/${methodName}`);
}

function isFullSyncMode(value: unknown): boolean {
  return (
    value === "SYNC_MODE_FULL" ||
    value === "FULL" ||
    value === 2 ||
    value === "2"
  );
}

function buildMockFullProgress(status: "SYNC_STATUS_IDLE" | "SYNC_STATUS_INTERRUPTED") {
  const rootProgressItem = {
    rootFolderId: "1001",
    status: "done",
    estimatedTotalDocs: "11",
    stats: {
      foldersVisited: 1,
      filesIndexed: 10,
      filesDiscovered: 10,
      skippedFiles: 0,
      pagesFetched: 1,
      failedRequests: 0,
      startedAt: "0",
      endedAt: "0",
    },
    updatedAt: "0",
  };

  return {
    state: {
      status,
      mode: "SYNC_MODE_FULL",
      startedAt: "0",
      updatedAt: String(Date.now()),
      roots: ["1001", "1002"],
      rootNames: {
        "1001": "A",
        "1002": "B",
      },
      completedRoots: [],
      aggregateStats: {
        foldersVisited: 0,
        filesIndexed: 0,
        filesDiscovered: 0,
        skippedFiles: 0,
        pagesFetched: 0,
        failedRequests: 0,
        startedAt: "0",
        endedAt: "0",
      },
      rootProgress: {
        "1001": rootProgressItem,
        "1002": {
          ...rootProgressItem,
          rootFolderId: "1002",
          estimatedTotalDocs: "21",
        },
      },
      catalogRoots: ["1001", "1002"],
      catalogRootNames: {
        "1001": "A",
        "1002": "B",
      },
      catalogRootProgress: {
        "1001": rootProgressItem,
        "1002": {
          ...rootProgressItem,
          rootFolderId: "1002",
          estimatedTotalDocs: "21",
        },
      },
    },
  };
}

test.describe("Admin 认证流程", () => {
  let adminPage: AdminPage;

  test.beforeEach(async ({ page }) => {
    adminPage = new AdminPage(page);
  });

  // Test 1: 未认证时显示 API Key 对话框
  test("未认证时显示 API Key 对话框", async ({ page }) => {
    await adminPage.goto();
    // Dialog should be visible
    await expect(adminPage.dialog).toBeVisible();
    // Password input should be visible
    await expect(adminPage.apiKeyInput).toBeVisible();
    // Submit button should show "确认"
    await expect(adminPage.submitButton).toBeVisible();
    await expect(adminPage.submitButton).toHaveText("确认");
  });

  // Test 2: 空 API Key 显示本地错误
  test("空 API Key 显示本地错误", async ({ page }) => {
    await adminPage.goto();
    // Click submit without entering key
    await adminPage.submitButton.click();
    // Should show local error
    await expect(page.getByText("请输入 API Key")).toBeVisible();
    // Dialog should still be visible
    await expect(adminPage.dialog).toBeVisible();
  });

  // Test 3: 错误 API Key 显示服务端错误
  test("错误 API Key 显示服务端错误", async ({ page }) => {
    await adminPage.goto();
    // Enter wrong key and submit
    await adminPage.submitApiKey("wrong-key-00000");
    // Wait for server response and error
    await expect(page.getByText("API Key 无效")).toBeVisible({
      timeout: 5_000,
    });
    // Dialog should still be visible
    await expect(adminPage.dialog).toBeVisible();
  });

  // Test 4: 正确 API Key 进入管理界面
  test("正确 API Key 进入管理界面", async ({ page }) => {
    await adminPage.goto();
    // Enter correct key
    await adminPage.submitApiKey(ADMIN_API_KEY);
    // Dialog should disappear
    await adminPage.waitForAuthComplete();
    // Should show sync management UI
    await expect(page.getByRole("heading", { name: "同步管理" })).toBeVisible();
    // Check localStorage
    const storedKey = await page.evaluate(() =>
      localStorage.getItem("npan_admin_api_key"),
    );
    expect(storedKey).toBe(ADMIN_API_KEY);
  });

  // Test 5: 刷新页面保持认证状态
  test("刷新页面保持认证状态", async ({ authenticatedPage }) => {
    const authAdminPage = new AdminPage(authenticatedPage);
    await authAdminPage.goto();
    // Should NOT show dialog
    await expect(authAdminPage.dialog).not.toBeVisible();
    // Should show sync management
    await expect(
      authenticatedPage.getByRole("heading", { name: "同步管理" }),
    ).toBeVisible();
    // Reload page
    await authenticatedPage.reload();
    // Should still be authenticated
    await expect(authAdminPage.dialog).not.toBeVisible();
    await expect(
      authenticatedPage.getByRole("heading", { name: "同步管理" }),
    ).toBeVisible();
  });

  // Test 6: 返回搜索链接
  test("返回搜索链接导航到首页", async ({ authenticatedPage }) => {
    const authAdminPage = new AdminPage(authenticatedPage);
    await authAdminPage.goto();
    await expect(authAdminPage.dialog).not.toBeVisible();
    // Click back to search
    await authAdminPage.backToSearchLink.click();
    // Should navigate to /
    await expect(authenticatedPage).toHaveURL(/\/$/);
  });
});

test.describe("Admin 同步控制", () => {
  let adminPage: AdminPage;

  test.beforeEach(async ({ authenticatedPage }) => {
    adminPage = new AdminPage(authenticatedPage);
    await adminPage.goto();
    // Wait for auth to be complete (dialog should not show)
    await expect(adminPage.dialog).not.toBeVisible();
    await expect(
      authenticatedPage.getByRole("heading", { name: "同步管理" }),
    ).toBeVisible();
  });

  test("显示同步模式选择器", async ({ authenticatedPage }) => {
    // Two mode buttons should be visible
    await expect(adminPage.modeButtons.full).toBeVisible();
    await expect(adminPage.modeButtons.incremental).toBeVisible();
    // "全量" should be selected by default
    await expect(adminPage.modeButtons.full).toHaveClass(/bg-white/);
  });

  test("选择全量模式", async ({ authenticatedPage }) => {
    await adminPage.selectMode("full");
    // "全量" should be selected
    await expect(adminPage.modeButtons.full).toHaveClass(/bg-white/);
    // Other mode should not be selected
    await expect(adminPage.modeButtons.incremental).not.toHaveClass(/bg-white/);
  });

  test("未建索引时增量模式不可用并显示提示", async ({ authenticatedPage }) => {
    await authenticatedPage.route("**/npan.v1.AdminService/GetIndexStats", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ documentCount: "0" }),
      });
    });

    await authenticatedPage.reload();
    await expect(
      authenticatedPage.getByRole("heading", { name: "同步管理" }),
    ).toBeVisible();
    await expect(authenticatedPage.getByText("请先执行一次全量索引")).toBeVisible();
    await expect(adminPage.modeButtons.incremental).toBeDisabled();
  });

  test("中断后再次全量启动保持断点续传语义", async ({ authenticatedPage }) => {
    const startPayloads: Record<string, unknown>[] = [];
    let phase: "before-restart" | "after-restart" = "before-restart";

    await authenticatedPage.route("**/npan.v1.AdminService/GetIndexStats", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ documentCount: "100" }),
      });
    });

    await authenticatedPage.route("**/npan.v1.AdminService/GetSyncProgress", async (route) => {
      const payload =
        phase === "before-restart"
          ? buildMockFullProgress("SYNC_STATUS_IDLE")
          : buildMockFullProgress("SYNC_STATUS_INTERRUPTED");
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(payload),
      });
    });

    await authenticatedPage.route("**/npan.v1.AdminService/WatchSyncProgress", async (route) => {
      await route.fulfill({
        status: 501,
        contentType: "application/json",
        body: JSON.stringify({ code: "unimplemented" }),
      });
    });

    await authenticatedPage.route("**/npan.v1.AdminService/StartSync", async (route) => {
      const body: unknown = route.request().postDataJSON();
      if (isRecord(body)) {
        startPayloads.push(body);
      }
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ message: "ok" }),
      });
    });

    await authenticatedPage.reload();
    await expect(authenticatedPage.getByText("当前已勾选 2 / 2")).toBeVisible();

    await adminPage.startSyncButton.click();
    await expect.poll(() => startPayloads.length).toBe(1);

    phase = "after-restart";
    await authenticatedPage.reload();
    await expect(adminPage.startSyncButton).toBeEnabled();

    await adminPage.startSyncButton.click();
    await expect.poll(() => startPayloads.length).toBe(2);

    const second = startPayloads[1];
    expect(isFullSyncMode(second.mode)).toBe(true);
    expect(second.forceRebuild).toBeUndefined();
    expect(second.resumeProgress).toBeUndefined();
    expect(second.rootFolderIds).toEqual(["1001", "1002"]);
  });

  test("中断后全量强制重建应显式关闭断点续传", async ({ authenticatedPage }) => {
    const startPayloads: Record<string, unknown>[] = [];

    await authenticatedPage.route("**/npan.v1.AdminService/GetIndexStats", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ documentCount: "100" }),
      });
    });

    await authenticatedPage.route("**/npan.v1.AdminService/GetSyncProgress", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(buildMockFullProgress("SYNC_STATUS_INTERRUPTED")),
      });
    });

    await authenticatedPage.route("**/npan.v1.AdminService/WatchSyncProgress", async (route) => {
      await route.fulfill({
        status: 501,
        contentType: "application/json",
        body: JSON.stringify({ code: "unimplemented" }),
      });
    });

    await authenticatedPage.route("**/npan.v1.AdminService/StartSync", async (route) => {
      const body: unknown = route.request().postDataJSON();
      if (isRecord(body)) {
        startPayloads.push(body);
      }
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ message: "ok" }),
      });
    });

    await authenticatedPage.reload();
    await expect(authenticatedPage.getByRole("heading", { name: "同步管理" })).toBeVisible();

    await authenticatedPage.getByRole("button", { name: /展开/ }).click();
    await authenticatedPage.getByRole("switch", { name: /选择根目录 1001/ }).click();
    await authenticatedPage.getByRole("switch", { name: /选择根目录 1002/ }).click();

    await authenticatedPage.getByRole("switch", { name: /强制重建索引/i }).click();
    await adminPage.startSyncButton.click();
    await authenticatedPage.getByRole("button", { name: "确认重建" }).click();

    await expect.poll(() => startPayloads.length).toBe(1);
    const payload = startPayloads[0];
    expect(isFullSyncMode(payload.mode)).toBe(true);
    expect(payload.forceRebuild).toBe(true);
    expect(payload.resumeProgress).toBe(false);
    expect(payload.rootFolderIds).toBeUndefined();
  });

  test("运行中仅保留取消和刷新目录详情", async ({ authenticatedPage }) => {
    const syncResponse = authenticatedPage.waitForResponse(
      (r) =>
        r.request().method() === "POST" &&
        isAdminMethodRequest(r.url(), "StartSync"),
      { timeout: 5_000 },
    );
    await adminPage.startSyncButton.click();
    await syncResponse;

    await expect(adminPage.startSyncButton).toBeDisabled({ timeout: 5_000 });
    await expect(adminPage.modeButtons.full).toBeDisabled();
    await expect(adminPage.modeButtons.incremental).toBeDisabled();
    await expect(adminPage.cancelSyncButton).toBeVisible();

    await authenticatedPage.getByRole("button", { name: /刷新目录详情/ }).click();
    await expect(authenticatedPage.getByRole("button", { name: /刷新目录详情/ })).toBeVisible();
  });

  test("启动同步发送正确请求", async ({ authenticatedPage }) => {
    // Select full mode
    await adminPage.selectMode("full");

    // Monitor the POST request
    const syncRequest = authenticatedPage.waitForRequest(
      (req) =>
        req.method() === "POST" && isAdminMethodRequest(req.url(), "StartSync"),
      { timeout: 5_000 },
    );

    // Click start sync
    await adminPage.startSyncButton.click();
    const request = await syncRequest;

    // Verify request has correct headers and body
    expect(request.headers()["x-api-key"]).toBeTruthy();
    const body: unknown = request.postDataJSON();
    expect(isRecord(body)).toBe(true);
    if (!isRecord(body)) {
      return;
    }
    expect(isFullSyncMode(body.mode)).toBe(true);
  });

  test("取消同步触发确认对话框", async ({ authenticatedPage }) => {
    // Start sync first
    const syncResponse = authenticatedPage.waitForResponse(
      (r) =>
        r.request().method() === "POST" &&
        isAdminMethodRequest(r.url(), "StartSync"),
      { timeout: 5_000 },
    );
    await adminPage.startSyncButton.click();
    await syncResponse;

    // Wait for cancel button to appear (if sync is running)
    // Note: with dummy NPA_TOKEN, sync may fail fast. Only test cancel if button appears
    try {
      await adminPage.cancelSyncButton.waitFor({
        state: "visible",
        timeout: 3_000,
      });
    } catch {
      // If cancel button doesn't appear (sync already finished), skip this test
      test.skip();
      return;
    }

    // Set up dialog handler - click confirm in custom dialog
    // Click cancel button to open confirm dialog
    await adminPage.cancelSyncButton.click();

    // Wait for confirm dialog to appear
    await expect(adminPage.dialog).toBeVisible();
    await expect(adminPage.confirmDialogMessage).toContainText(
      "确认取消当前正在进行的同步任务",
    );

    // Monitor CancelSync Connect request (POST)
    const cancelRequest = authenticatedPage.waitForRequest(
      (req) =>
        req.method() === "POST" && isAdminMethodRequest(req.url(), "CancelSync"),
      { timeout: 5_000 },
    );

    // Click confirm button in dialog
    await authenticatedPage.getByRole("button", { name: "确认取消" }).click();
    await cancelRequest;
  });

  test("取消确认框点击取消不发请求", async ({ authenticatedPage }) => {
    // Start sync first
    const syncResponse = authenticatedPage.waitForResponse(
      (r) =>
        r.request().method() === "POST" &&
        isAdminMethodRequest(r.url(), "StartSync"),
      { timeout: 5_000 },
    );
    await adminPage.startSyncButton.click();
    await syncResponse;

    // Wait for cancel button
    try {
      await adminPage.cancelSyncButton.waitFor({
        state: "visible",
        timeout: 3_000,
      });
    } catch {
      test.skip();
      return;
    }

    // Click cancel button to open confirm dialog
    await adminPage.cancelSyncButton.click();

    // Wait for confirm dialog to appear
    await expect(adminPage.dialog).toBeVisible();

    // Track if CancelSync request is sent
    let cancelSent = false;
    authenticatedPage.on("request", (req) => {
      if (
        req.method() === "POST" &&
        isAdminMethodRequest(req.url(), "CancelSync")
      ) {
        cancelSent = true;
      }
    });

    // Click the cancel button inside confirm dialog (dismiss)
    await adminPage.confirmDialogCancelButton.click();

    // Dialog should close
    await expect(adminPage.dialog).not.toBeVisible();

    // Wait briefly and verify no CancelSync request was sent
    await authenticatedPage.waitForTimeout(500);
    expect(cancelSent).toBe(false);
  });

  test("启动同步后 UI 自动显示 running 状态", async ({ authenticatedPage }) => {
    // Monitor POST request and response
    const syncResponse = authenticatedPage.waitForResponse(
      (r) =>
        r.request().method() === "POST" &&
        isAdminMethodRequest(r.url(), "StartSync"),
      { timeout: 5_000 },
    );

    // Click start sync
    await adminPage.startSyncButton.click();

    // Wait for POST to complete
    await syncResponse;

    // 新交互中按钮文案固定为“启动同步”，运行中通过禁用和取消按钮体现状态
    await expect(adminPage.startSyncButton).toContainText("启动同步", {
      timeout: 5_000,
    });
    await expect(adminPage.startSyncButton).toBeDisabled();
    await expect(adminPage.cancelSyncButton).toBeVisible();
  });

  test("WatchSyncProgress 不应出现 Flusher internal 错误", async ({
    authenticatedPage,
  }) => {
    const internalErrors: string[] = [];
    authenticatedPage.on("console", (msg) => {
      if (msg.type() !== "error") {
        return;
      }
      const text = msg.text();
      if (
        text.includes("http.Flusher") ||
        text.includes("WatchSyncProgress") ||
        text.includes('code: "internal"')
      ) {
        internalErrors.push(text);
      }
    });
    authenticatedPage.on("pageerror", (err) => {
      const text = err.message;
      if (
        text.includes("http.Flusher") ||
        text.includes("WatchSyncProgress") ||
        text.includes("internal")
      ) {
        internalErrors.push(text);
      }
    });

    await authenticatedPage.reload();
    await expect(
      authenticatedPage.getByRole("heading", { name: "同步管理" }),
    ).toBeVisible();
    await authenticatedPage.waitForTimeout(3_500);

    expect(internalErrors).toEqual([]);
  });
});

test.describe("Admin 边界场景", () => {
  test("认证过期后显示对话框", async ({ page }) => {
    const adminPage = new AdminPage(page);

    // Navigate to admin first, then set localStorage
    await page.goto("/admin/");
    await page.evaluate((key: string) => {
      localStorage.setItem("npan_admin_api_key", key);
    }, ADMIN_API_KEY);

    // Reload to pick up the key
    await page.reload();
    await expect(adminPage.dialog).not.toBeVisible();
    await expect(page.getByRole("heading", { name: "同步管理" })).toBeVisible();

    // Clear localStorage to simulate expiry
    await page.evaluate(() => localStorage.removeItem("npan_admin_api_key"));

    // Reload page
    await page.reload();

    // Should show API Key dialog again
    await expect(adminPage.dialog).toBeVisible();
  });
});
