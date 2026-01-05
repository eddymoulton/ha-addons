export interface ConfigMetadata {
  id: string;
  path: string;
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

export type BackupType = "multiple" | "single" | "directory";

export interface ConfigBackupOptions {
  path: string;
  backupType: BackupType;
  maxBackups?: number;
  maxBackupAgeDays?: number;
  idNode?: string;
  friendlyNameNode?: string;
  includeFilePatterns?: string[];
  excludeFilePatterns?: string[];
}

export interface ConfigBackupOptionGroup {
  groupName: string;
  configs: ConfigBackupOptions[];
}

export interface AppSettings {
  homeAssistantConfigDir: string;
  backupDir: string;
  port: string;
  cronSchedule?: string;
  defaultMaxBackups?: number;
  defaultMaxBackupAgeDays?: number;
  configGroups: ConfigBackupOptionGroup[];
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

export interface ConfigResponse {
  groups: Record<string, ConfigMetadata[]>;
}
