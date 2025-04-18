/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
import * as React from "react";
import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";

import "../sass/footer.scss";

const Footer: React.FunctionComponent = (): React.JSX.Element => {
  return (
    <Box className="Footer" sx={{ width: "auto" }} component="footer">
      <Typography variant="h5" gutterBottom>
        Template application please replace with your own
      </Typography>
    </Box>
  );
};

export default Footer;
