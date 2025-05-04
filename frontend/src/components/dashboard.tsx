/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
import * as React from "react";
import { useEffect } from "react";
import { Box } from "@mui/material";
import Layout from "./layout";
import ConfigMapList from "./configmaplist";

const Dashboard: React.FunctionComponent = (): React.JSX.Element => {
  useEffect(() => {
    document.title = "Dashboard";
  }, []);

  return (
    <Layout>
      <Box sx={{ margin: "0 auto", width: "100%" }}>
        <ConfigMapList />
      </Box>
    </Layout>
  );
};

export default Dashboard;
