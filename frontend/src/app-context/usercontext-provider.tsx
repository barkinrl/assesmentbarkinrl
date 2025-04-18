/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
import * as React from "react";
import { useState, useEffect, useRef } from "react";
import { useAuth } from "react-oidc-context";

import {
  AppContext,
  UserContract,
  DataContract,
  PageContract,
} from "./usercontext";

interface Props {
  children: React.ReactNode;
}

export const AppContextProvider: React.FunctionComponent<Props> = (
  props: Props,
): React.JSX.Element => {
  const auth = useAuth();

  const authRef = useRef(auth);

  useEffect(() => {
    authRef.current = auth;
  }, [auth]);

  const [user, setUser] = useState<UserContract>({
    username: "admin",
    isAdmin: true,
  });

  const dataPublish = () => {
    return;
  };

  const [data, setData] = useState<DataContract>({
    isDirty: false,
    version: null,
    publish: dataPublish,
  });

  const dataRef = useRef(data);

  useEffect(() => {
    dataRef.current = data;
  }, [data]);

  const [page, setPage] = useState<PageContract>({
    isEditable: false,
    inEdit: false,
    isDirty: false,
  });

  const updateUser = (newUser: Partial<UserContract>) => {
    setUser((prevUser) => {
      return { ...prevUser, ...newUser };
    });
  };

  const updateData = (newData: Partial<DataContract>) => {
    setData((prevData) => {
      return { ...prevData, ...newData };
    });
  };

  const updatePage = (newPage: Partial<PageContract>) => {
    setPage((prevPage) => {
      return { ...prevPage, ...newPage };
    });
  };

  useEffect(() => {
    if (SSO_ENABLED) {
      if (auth.activeNavigator === "signinSilent") {
        updatePage({
          errorMessage: {
            title: "Authenticating",
            message: "Please wait...",
          },
        });
      }

      if (auth.isLoading) {
        updatePage({
          errorMessage: {
            title: "Loading",
            message: "Please wait...",
          },
        });

        return;
      }

      if (auth.error) {
        updateData({ version: null });

        updatePage({
          errorMessage: {
            title: "Error while getting namespaces",
            message: auth.error.message,
          },
        });

        return;
      }

      if (!auth.isAuthenticated) {
        updateData({ version: null });

        updatePage({
          errorMessage: {
            title: "Authenticating",
            message: "Please wait...",
          },
        });

        auth.signinRedirect();

        return;
      }
    }

    updatePage({ errorMessage: undefined });

    // fetch vpnGroups from /api?action=get_vpngroups endpoint
    // and update data.vpnGroups
    const data = JSON.stringify({
      action: "get_version",
    });

    let headers = {};

    if (SSO_ENABLED) {
      const access_token = auth.user?.access_token;
      const id_token = auth.user?.id_token;

      if (!access_token || !id_token) {
        updatePage({
          errorMessage: {
            title: "Error while getting namespaces",
            message: "Access/ID Token is not available",
          },
        });

        return;
      }

      headers = {
        Authorization: `Bearer ${access_token}`,
        "X-ID-Token": id_token,
      };
    }

    fetch("/api?data=" + data, {
      headers: headers,
    })
      .then((response) => response.json())
      .then((data) => {
        if (data.error) {
          throw new Error(data.error);
        }

        updateData({ version: data });
      })
      .catch((error) => {
        updatePage({
          errorMessage: {
            title: "Error while getting namespaces",
            message: error.message,
          },
        });
      });
  }, [auth]);

  return (
    <AppContext.Provider
      value={{
        user,
        data,
        page,
        updateUser,
        updateData,
        updatePage,
      }}
    >
      {props.children}
    </AppContext.Provider>
  );
};
