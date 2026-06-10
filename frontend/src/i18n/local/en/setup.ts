export const setup = {
  title: "Set up Impress",
  subtitle: "Create an admin account and configure basic site info",
  steps: {
    welcome: "Welcome",
    site: "Site",
    admin: "Admin",
    content: "Content",
    finish: "Done",
  },
  welcome: {
    heading: "Welcome to Impress",
    body: "The database is configured via environment variables. Complete the steps below.",
    database: "Database type",
    next: "Get started",
  },
  site: {
    nameZh: "Site name (Chinese)",
    nameEn: "Site name (English)",
    defaultLocale: "Default locale",
    localeZh: "Chinese",
    localeEn: "English",
  },
  admin: {
    username: "Admin username",
    password: "Password",
    confirmPassword: "Confirm password",
    hint: "At least 8 characters with letters and digits",
  },
  content: {
    heading: "Initial content",
    blank: "Blank site",
    blankDesc: "Minimal pages and theme — good for a personal blog",
    demo: "Demo data",
    demoDesc: "Sample articles and consultancy content to explore features",
  },
  actions: {
    back: "Back",
    next: "Next",
    finish: "Complete setup",
    finishing: "Installing…",
  },
  success: {
    heading: "Setup complete",
    body: "Redirecting to sign in…",
  },
  errors: {
    passwordMismatch: "Passwords do not match",
    siteNameRequired: "Enter a site name in at least one language",
    setupFailed: "Setup failed. Please try again.",
  },
};
