import { http } from "./http";

export interface SetupStatus {
  installed: boolean;
  databaseType: "sqlite" | "postgres" | string;
  bootstrapMode?: boolean;
  needsEnvConfig?: boolean;
  envSecretsLoaded?: boolean;
  envFilePath?: string;
  serverPort?: number;
}

export interface DatabaseConfig {
  type: "sqlite" | "postgres";
  sqlitePath?: string;
  postgres?: {
    host: string;
    port: number;
    user: string;
    password: string;
    dbname: string;
    sslmode?: string;
  };
}

export interface SetupCompletePayload {
  admin: {
    username: string;
    password: string;
  };
  site: {
    name: { zh?: string; en?: string };
    defaultLocale: "zh" | "en";
  };
  seedMode: "blank" | "demo";
}

export interface SaveEnvPayload {
  port: number;
  env: "development" | "production";
  database: DatabaseConfig;
}

export interface SaveEnvResult {
  success: boolean;
  restartRequired: boolean;
  envPath: string;
}

export async function fetchSetupStatus(): Promise<SetupStatus> {
  const res = await http.get<SetupStatus>("/setup/status");
  return res.data;
}

export async function testDatabase(database: DatabaseConfig): Promise<void> {
  await http.post("/setup/test-database", database);
}

export async function saveSetupEnv(payload: SaveEnvPayload): Promise<SaveEnvResult> {
  const res = await http.post<SaveEnvResult>("/setup/save-env", payload);
  return res.data;
}

export async function completeSetup(payload: SetupCompletePayload): Promise<void> {
  await http.post("/setup/complete", payload);
}
