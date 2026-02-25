import { test, expect, ADMIN_API_KEY } from "../fixtures/auth";
import { AdminPage } from "../pages/admin-page";

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function isAdminMethodRequest(
  url: string,
  methodName: "StartSync" | "CancelSync" | "WatchSyncProgress",
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
    // All three mode buttons should be visible
    await expect(adminPage.modeButtons.auto).toBeVisible();
    await expect(adminPage.modeButtons.full).toBeVisible();
    await expect(adminPage.modeButtons.incremental).toBeVisible();
    // "自适应" should be selected by default (has bg-white class)
    await expect(adminPage.modeButtons.auto).toHaveClass(/bg-white/);
  });

  test("选择全量模式", async ({ authenticatedPage }) => {
    await adminPage.selectMode("full");
    // "全量" should be selected
    await expect(adminPage.modeButtons.full).toHaveClass(/bg-white/);
    // Others should not be selected
    await expect(adminPage.modeButtons.auto).not.toHaveClass(/bg-white/);
    await expect(adminPage.modeButtons.incremental).not.toHaveClass(/bg-white/);
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

    // UI should automatically show "同步进行中" without page.reload()
    // The button text changes to "同步进行中" when isRunning is true
    await expect(adminPage.startSyncButton).toContainText("同步进行中", {
      timeout: 5_000,
    });
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
