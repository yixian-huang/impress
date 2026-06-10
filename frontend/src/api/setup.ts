import { http } from "./http";

export interface SetupStatus {
  installed: boolean;
  databaseType: "sqlite" | "postgres" | string;
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

export async function fetchSetupStatus(): Promise<SetupStatus> {
  const res = await http.get<SetupStatus>("/setup/status");
  return res.data;
}

export async function completeSetup(payload: SetupCompletePayload): Promise<void> {
  await http.post("/setup/complete", payload);
}
