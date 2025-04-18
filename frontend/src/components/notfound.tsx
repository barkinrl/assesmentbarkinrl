/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
import * as React from "react";
import { useEffect } from "react";
import { useParams } from "react-router-dom";
import { Box, Alert, AlertTitle } from "@mui/material";
import Layout from "./layout";

const NotFound: React.FunctionComponent = (): React.JSX.Element => {
  const { "*": path } = useParams();

  useEffect(() => {
    document.title = "Page not found";
  }, [path]);

  return (
    <Layout>
      <Box sx={{ margin: "0 auto", width: "100%" }}>
        <Alert severity="error">
          <AlertTitle>Error</AlertTitle>
          {`Page not found: ${path}`}
        </Alert>
      </Box>
    </Layout>
  );
};

export default NotFound;
