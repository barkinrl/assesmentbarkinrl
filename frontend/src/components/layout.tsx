/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
import * as React from "react";
import { useContext } from "react";
import { Box } from "@mui/material";
import { ThemeProvider } from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";
import Alert from "@mui/material/Alert";
import AlertTitle from "@mui/material/AlertTitle";

import { useAuth } from "react-oidc-context";

import Footer from "./footer";
import AppBar from "./appbar";
import BackToTop from "./backtotop";
import { theme } from "../theme/theme-provider";
import { AppContext } from "../app-context/usercontext";
import "../sass/app.scss";

interface Props {
  /**
   * Injected by the documentation to work in an iframe.
   * You won't need it on your project.
   */
  window?: () => Window;
  children?: React.ReactElement[] | React.ReactElement;
}

const Layout: React.FunctionComponent<Props> = ({ children }) => {
  const { page } = useContext(AppContext);

  const auth = useAuth();

  const real_body = (
    <>
      <AppBar />
      <Box
        component="main"
        sx={{
          display: "flex",
          flexGrow: 1,
          padding: "2rem 0",
          paddingTop: "80px",
        }}
      >
        {children}
      </Box>
    </>
  );

  let body = null;

  if (SSO_ENABLED) {
    if (auth.isAuthenticated) {
      body = real_body;
    }
  } else {
    body = real_body;
  }

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Box className="App">
        <BackToTop>
          <Box
            sx={{
              display: "flex",
              flexDirection: "column",
              minHeight: "100vh",
            }}
          >
            {body}
            {page.errorMessage && (
              <Box sx={{ width: "100%", display: "block" }}>
                <Alert severity="error">
                  <AlertTitle>{page.errorMessage.title}</AlertTitle>
                  {page.errorMessage.message}
                </Alert>
              </Box>
            )}
            <Footer />
          </Box>
        </BackToTop>
      </Box>
    </ThemeProvider>
  );
};

export default Layout;
