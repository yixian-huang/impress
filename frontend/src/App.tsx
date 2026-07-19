import { BrowserRouter } from "react-router-dom";
import { AppRoutes } from "./router";
import { I18nextProvider } from "react-i18next";
import i18n from "./i18n";
import { AuthProvider } from "@/contexts/AuthContext";
import { GlobalConfigProvider } from "@/contexts/GlobalConfigContext";
import { ThemePagesProvider } from "@/contexts/ThemePagesContext";
import { BootstrapProvider } from "@/contexts/BootstrapContext";
import { ThemeProvider } from "@/theme";
import { ThemeManagerProvider } from "@/plugins/ThemeManagerContext";
import DocumentBranding from "@/components/DocumentBranding";


function App() {
  return (
    <I18nextProvider i18n={i18n}>
      <BrowserRouter basename={__BASE_PATH__}>
        <BootstrapProvider>
          <ThemeManagerProvider>
            <ThemeProvider>
              <ThemePagesProvider>
                <GlobalConfigProvider>
                  <DocumentBranding />
                  <AuthProvider>
                    <AppRoutes />
                  </AuthProvider>
                </GlobalConfigProvider>
              </ThemePagesProvider>
            </ThemeProvider>
          </ThemeManagerProvider>
        </BootstrapProvider>
      </BrowserRouter>
    </I18nextProvider>
  );
}

export default App;
