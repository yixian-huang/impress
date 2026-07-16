import { http } from "@/api/http";

export type AIProviderName = "disabled" | "openai" | "anthropic";

export interface AIConfig {
  provider: AIProviderName;
  enabled: boolean;
  base_url?: string;
  model?: string;
  has_api_key: boolean;
  api_key_masked?: string;
}

export interface AIConfigInput {
  provider: AIProviderName;
  api_key?: string;
  base_url?: string;
  model?: string;
}

export interface AIHealthResult {
  provider: string;
  healthy: boolean;
  model?: string;
  message?: string;
}

export async function getAIConfig(): Promise<AIConfig> {
  const response = await http.get<AIConfig>("/admin/ai/config");
  return response.data;
}

export async function updateAIConfig(input: AIConfigInput): Promise<AIConfig> {
  const response = await http.put<AIConfig>("/admin/ai/config", input);
  return response.data;
}

export async function testAIConfig(input: AIConfigInput): Promise<AIHealthResult> {
  const response = await http.post<AIHealthResult>("/admin/ai/config/test", input);
  return response.data;
}
