import { http } from "./http";
import type { EmailConfig } from "@/pages/admin/email-settings/types";

export async function getEmailSettings(): Promise<EmailConfig> {
  const res = await http.get<EmailConfig>("/admin/email-settings", {

  });
  return res.data;
}

export async function updateEmailSettings(config: EmailConfig): Promise<EmailConfig> {
  const res = await http.put<EmailConfig>("/admin/email-settings", config, {

  });
  return res.data;
}

export async function sendTestEmail(to: string): Promise<{ success: boolean; message: string }> {
  const res = await http.post<{ success: boolean; message: string }>("/admin/email-settings/test", { to }, {

  });
  return res.data;
}
