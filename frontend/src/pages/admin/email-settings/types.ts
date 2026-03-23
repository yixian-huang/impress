export interface SMTPConfig {
  host: string;
  port: number;
  username: string;
  password: string;
  from: string;
  fromName: string;
  useTLS: boolean;
  insecureSkipVerify: boolean;
}

export interface ReceiverConfig {
  enabled: boolean;
  emails: string;
}

export interface AutoReplyConfig {
  enabled: boolean;
}

export interface EmailTemplate {
  subject: string;
  body: string;
}

export interface TemplatesConfig {
  autoReply: Record<string, EmailTemplate>;
  forward: Record<string, EmailTemplate>;
}

export interface EmailConfig {
  smtp: SMTPConfig;
  receiver: ReceiverConfig;
  autoReply: AutoReplyConfig;
  templates: TemplatesConfig;
}
