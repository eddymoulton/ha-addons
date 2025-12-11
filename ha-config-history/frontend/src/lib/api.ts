import type {
  ConfigMetadata,
  BackupInfo,
  BackupDiffResponse,
  AppSettings,
  UpdateSettingsResponse,
  RestoreBackupResponse,
  ConfigResponse,
} from "./types";

const API_BASE = window.location.href.replace(/\/+$/, "") || "";

export class ApiClient {
  async getConfigs(): Promise<ConfigResponse> {
    const response = await fetch(`${API_BASE}/configs`);
    if (!response.ok) {
      throw new Error(`Failed to fetch configs: ${response.statusText}`);
    }
    return response.json();
  }

  async getConfigBackups(path: string, id: string): Promise<BackupInfo[]> {
    const response = await fetch(`${API_BASE}/configs/${path}/${id}/backups`);
    if (!response.ok) {
      throw new Error(`Failed to fetch backups: ${response.statusText}`);
    }
    return response.json();
  }

  async getBackupContent(
    path: string,
    id: string,
    filename: string
  ): Promise<string> {
    const response = await fetch(
      `${API_BASE}/configs/${path}/${id}/backups/${filename}`
    );
    if (!response.ok) {
      throw new Error(`Failed to fetch backup content: ${response.statusText}`);
    }
    return response.text();
  }

  async compareBackups(
    path: string,
    id: string,
    leftFilename: string,
    rightFilename: string
  ): Promise<BackupDiffResponse> {
    const response = await fetch(
      `${API_BASE}/configs/${path}/${id}/compare/${encodeURIComponent(
        leftFilename
      )}/diff/${encodeURIComponent(rightFilename)}`
    );
    if (!response.ok) {
      throw new Error(`Failed to fetch backup diff: ${response.statusText}`);
    }
    return response.json();
  }

  async getSettings(): Promise<AppSettings> {
    const response = await fetch(`${API_BASE}/settings`);
    if (!response.ok) {
      throw new Error(`Failed to fetch settings: ${response.statusText}`);
    }
    return response.json();
  }

  async updateSettings(settings: AppSettings): Promise<UpdateSettingsResponse> {
    const response = await fetch(`${API_BASE}/settings`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(settings),
    });
    if (!response.ok) {
      throw new Error(`Failed to update settings: ${response.statusText}`);
    }
    return response.json();
  }

  async restoreBackup(
    path: string,
    id: string,
    filename: string
  ): Promise<RestoreBackupResponse> {
    const response = await fetch(
      `${API_BASE}/configs/${path}/${id}/backups/${encodeURIComponent(
        filename
      )}/restore`,
      {
        method: "POST",
      }
    );
    if (!response.ok) {
      throw new Error(`Failed to restore backup: ${response.statusText}`);
    }
    return response.json();
  }

  async triggerBackup(): Promise<{ status: string }> {
    const response = await fetch(`${API_BASE}/backup`, {
      method: "POST",
    });
    if (!response.ok) {
      throw new Error(`Failed to trigger backup: ${response.statusText}`);
    }
    return response.json();
  }

  async deleteBackup(
    path: string,
    id: string,
    filename: string
  ): Promise<{ status: string }> {
    const response = await fetch(
      `${API_BASE}/configs/${path}/${id}/backups/${encodeURIComponent(
        filename
      )}`,
      {
        method: "DELETE",
      }
    );
    if (!response.ok) {
      throw new Error(`Failed to delete backup: ${response.statusText}`);
    }
    return response.json();
  }

  async deleteAllBackups(path: string, id: string): Promise<{ status: string }> {
    const response = await fetch(`${API_BASE}/configs/${path}/${id}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      throw new Error(`Failed to delete all backups: ${response.statusText}`);
    }
    return response.json();
  }
}

export const api = new ApiClient();
