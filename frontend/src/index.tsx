/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
import * as React from "react";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import {
  RouterProvider,
  RouteObject,
  createBrowserRouter,
} from "react-router-dom";
import { CacheProvider } from "@emotion/react";
import createCache from "@emotion/cache";

import { AuthProvider, AuthProviderProps } from "react-oidc-context";
import { WebStorageStateStore } from "oidc-client-ts";
import { Log } from "oidc-client-ts";

import Dashboard from "./components/dashboard";
import NotFound from "./components/notfound";

import { AppContextProvider } from "./app-context/usercontext-provider";

let oidcConfig: AuthProviderProps | null = null;

if (SSO_ENABLED) {
  oidcConfig = {
    authority: SSO_AUTHORITY_URL,
    client_id: SSO_CLIENT_ID,
    scope: "openid profile email",
    redirect_uri: window.location.href,
    onSigninCallback: () => {
      window.history.replaceState({}, document.title, window.location.pathname);
    },
    userStore: new WebStorageStateStore({ store: window.localStorage }),
    monitorSession: true,
  };
}

Log.setLogger(console);
Log.setLevel(Log.INFO);

const rootContainer = document.getElementById("root");

if (!rootContainer) {
  throw new Error("Root container not found");
}

const root = createRoot(rootContainer);

const routes: RouteObject[] = [
  { path: "/", element: <Dashboard /> },
  { path: "*", element: <NotFound /> },
];

// get none from html/head/link[rel=stylesheet]
const nonce = document
  .querySelector("link[rel=stylesheet]")
  ?.getAttribute("nonce");

if (!nonce) {
  throw new Error("Nonce not found");
}

const cache = createCache({
  key: "csp-nonced-emotion-cache",
  nonce: nonce,
  prepend: true,
});

root.render(
  <StrictMode>
    <CacheProvider value={cache}>
      <AuthProvider {...oidcConfig}>
        <AppContextProvider>
          <RouterProvider router={createBrowserRouter(routes)} />
        </AppContextProvider>
      </AuthProvider>
    </CacheProvider>
  </StrictMode>,
);
