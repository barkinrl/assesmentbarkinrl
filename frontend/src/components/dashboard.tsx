/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
import * as React from "react";
import { useEffect, useContext } from "react";
import { Box, Alert, AlertTitle } from "@mui/material";
import Layout from "./layout";
import { AppContext } from "../app-context/usercontext";

const Dashboard: React.FunctionComponent = (): React.JSX.Element => {
  const { data } = useContext(AppContext);

  useEffect(() => {
    document.title = "Dashboard";
  }, []);

  return (
    <Layout>
      <Box sx={{ margin: "0 auto", width: "100%" }}>
        <Alert severity="info">
          <AlertTitle>Info</AlertTitle>
          This is a sample dashboard page. You can add your own content here.
          <br />
          {data && data.version && (
            <span>
              Version: {data.version.version}
              <br />
              Build Date: {data.version.build_time}
              <br />
              Go Version: {data.version.go_version}
            </span>
          )}
        </Alert>
      </Box>
    </Layout>
  );
};

export default Dashboard;
