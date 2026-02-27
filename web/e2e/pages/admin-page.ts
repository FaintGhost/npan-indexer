import type { Locator, Page } from "@playwright/test";

export class AdminPage {
  readonly page: Page;
  readonly apiKeyInput: Locator;
  readonly submitButton: Locator;
  readonly startSyncButton: Locator;
  readonly cancelSyncButton: Locator;
  readonly modeButtons: {
    full: Locator;
    incremental: Locator;
  };
  readonly dialog: Locator;
  readonly confirmDialogMessage: Locator;
  readonly confirmDialogCancelButton: Locator;
  readonly syncMessage: Locator;
  readonly errorMessage: Locator;
  readonly backToSearchLink: Locator;
  readonly syncStatusBadge: Locator;

  constructor(page: Page) {
    this.page = page;
    this.apiKeyInput = page.locator('input[type="password"]');
    this.submitButton = page.locator('button[type="submit"]');
    this.startSyncButton = page.getByRole("button", {
      name: /启动同步|同步进行中|启动中/,
    });
    this.cancelSyncButton = page.getByRole("button", { name: "取消同步" });
    this.modeButtons = {
      full: page.getByRole("button", { name: "全量" }),
      incremental: page.getByRole("button", { name: "增量" }),
    };
    this.dialog = page.locator(".fixed.inset-0");
    this.confirmDialogMessage = page.locator(".fixed.inset-0 p");
    this.confirmDialogCancelButton = page.getByRole("button", {
      name: "取消",
      exact: true,
    });
    this.syncMessage = page.locator(".border-emerald-200");
    this.errorMessage = page.locator(".border-rose-200");
    this.backToSearchLink = page.getByRole("link", { name: /返回搜索/ });
    this.syncStatusBadge = page.locator("text=运行中");
  }

  async goto(): Promise<void> {
    let lastError: unknown;
    for (let attempt = 0; attempt < 2; attempt += 1) {
      try {
        await this.page.goto("/admin/", {
          waitUntil: "domcontentloaded",
          timeout: 12_000,
        });
        await Promise.any([
          this.page
            .getByRole("heading", { name: "同步管理" })
            .waitFor({ state: "visible", timeout: 5_000 }),
          this.apiKeyInput.waitFor({ state: "visible", timeout: 5_000 }),
        ]);
        return;
      } catch (error) {
        lastError = error;
        if (attempt === 1) {
          break;
        }
        await this.page.waitForTimeout(400);
      }
    }
    throw lastError;
  }

  async submitApiKey(key: string): Promise<void> {
    await this.apiKeyInput.fill(key);
    await this.submitButton.click();
  }

  async injectApiKey(key: string): Promise<void> {
    await this.page.context().addInitScript((k: string) => {
      localStorage.setItem("npan_admin_api_key", k);
    }, key);
  }

  async selectMode(mode: "full" | "incremental"): Promise<void> {
    await this.modeButtons[mode].click();
  }

  async waitForAuthComplete(): Promise<void> {
    await this.dialog.waitFor({ state: "hidden" });
  }
}
