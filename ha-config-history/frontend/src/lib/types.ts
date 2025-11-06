export interface ConfigMetadata {
  id: string;
  group: string;
  friendlyName: string;
  lastHash?: string;
  backupCount: number;
  backupsSize: number;
}

export interface BackupInfo {
  filename: string;
  date: string;
  size: number;
}

export interface BackupDiffResponse {
  type: "diff" | "content";
  unifiedDiff?: string;
  content?: string;
  oldContent?: string;
  newContent?: string;
  oldFilename?: string;
  newFilename?: string;
  isFirstBackup: boolean;
}

export type ComparisonMode = "previous" | "current" | "two-backups";

export interface ConfigBackupOptions {
  name: string;
  path: string;
  backupType: "multiple" | "single" | "directory";
  maxBackups?: number;
  maxBackupAgeDays?: number;
  idNode?: string;
  friendlyNameNode?: string;
}

export interface AppSettings {
  homeAssistantConfigDir: string;
  backupDir: string;
  port: string;
  cronSchedule?: string;
  defaultMaxBackups?: number;
  defaultMaxBackupAgeDays?: number;
  configs: ConfigBackupOptions[];
}

export interface UpdateSettingsResponse {
  success: boolean;
  warnings?: string[];
  error?: string;
}

export interface RestoreBackupResponse {
  success: boolean;
  message?: string;
  error?: string;
}
