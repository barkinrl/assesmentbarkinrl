/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
import * as React from "react";
import { useState, MouseEvent, useEffect, useRef, useContext } from "react";
import { AppBar as MuiAppBar } from "@mui/material";
import {
  Toolbar,
  IconButton,
  Menu,
  MenuItem,
  Box,
  useMediaQuery,
} from "@mui/material";
import MenuIcon from "@mui/icons-material/Menu";
import MoreIcon from "@mui/icons-material/MoreVert";
import DashboardIcon from "@mui/icons-material/Dashboard";
import Link from "@mui/material/Link";
import { useTheme } from "@mui/material/styles";
import { AppContext } from "../app-context/usercontext";
import EditNoteIcon from "@mui/icons-material/EditNote";
import SaveIcon from "@mui/icons-material/Save";
import CancelIcon from "@mui/icons-material/Cancel";
import PublishIcon from "@mui/icons-material/Publish";
import AddCircleOutlineIcon from "@mui/icons-material/AddCircleOutline";
import { useAuth } from "react-oidc-context";

import "../sass/appbar.scss";

const AppBar: React.FunctionComponent = () => {
  const { page, updatePage, data } = useContext(AppContext);

  const auth = useAuth();

  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down("md")); // Detect small screens

  const [leftAnchorEl, setLeftAnchorEl] = useState<null | HTMLElement>(null);
  const [rightAnchorEl, setRightAnchorEl] = useState<null | HTMLElement>(null);

  const menuTimeout = 5000; // 5 seconds

  const handleLeftMenuOpen = (event: MouseEvent<HTMLButtonElement>) => {
    setLeftAnchorEl(event.currentTarget);
  };

  const handleRightMenuOpen = (event: MouseEvent<HTMLButtonElement>) => {
    setRightAnchorEl(event.currentTarget);
  };

  const handleMenuClose = () => {
    setLeftAnchorEl(null);
    setRightAnchorEl(null);
  };

  const handleEdit = () => {
    updatePage({ inEdit: true });
    handleMenuClose();
  };

  const handleCancel = () => {
    if (page.onCancel) {
      page.onCancel();
    }

    updatePage({ inEdit: false, isDirty: false });
    handleMenuClose();
  };

  const handleSave = () => {
    if (page.onSave) {
      if (!page.onSave()) {
        return;
      }
    }

    updatePage({ inEdit: false, isDirty: false });
    handleMenuClose();
  };

  const handleInsert = () => {
    if (page.onInsert) {
      page.onInsert();
    }

    handleMenuClose();
  };

  const handlePublish = () => {
    if (data.publish) {
      data.publish();
    }

    handleMenuClose();
  };

  // Auto close menu after a timeout of no interaction
  useEffect(() => {
    let autoCloseTimeout: number;

    if (leftAnchorEl || rightAnchorEl) {
      // Start a timeout to auto-close the menu after 'menuTimeout' milliseconds
      autoCloseTimeout = window.setTimeout(() => {
        handleMenuClose();
      }, menuTimeout);
    }

    return () => {
      // Clear the timeout if the component unmounts or if the menu is closed manually
      window.clearTimeout(autoCloseTimeout);
    };
  }, [leftAnchorEl, rightAnchorEl]);

  const toolBarRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const resizeHandler = () => {
      if (toolBarRef.current) {
        // max toolbar width is 1200px
        // if window width is less than 1200px, set the toolbar width to window width
        // else set toolbar width to 1200px
        toolBarRef.current.style.width = `${Math.min(window.innerWidth, 1200)}px`;
      }
    };

    resizeHandler();

    window.addEventListener("resize", resizeHandler);

    return () => {
      window.removeEventListener("resize", resizeHandler);
    };
  }, [toolBarRef]);

  const rightSideEditLinks = () => {
    if (!auth.isAuthenticated) {
      return null;
    }

    if (!page.isEditable) {
      return null;
    }

    if (!page.inEdit) {
      return (
        <>
          <Link color="inherit" underline="none" onClick={handleEdit}>
            <EditNoteIcon />
          </Link>
        </>
      );
    }

    return (
      <>
        <Link color="inherit" underline="none" onClick={handleInsert}>
          <AddCircleOutlineIcon />
        </Link>
        {page.isDirty && (
          <Link color="inherit" underline="none" onClick={handleSave}>
            <SaveIcon />
          </Link>
        )}
        <Link color="inherit" underline="none" onClick={handleCancel}>
          <CancelIcon />
        </Link>
      </>
    );
  };

  const rightSidePublishLink = () => {
    if (!auth.isAuthenticated) {
      return null;
    }

    if (data.isDirty) {
      return (
        <Link color="inherit" underline="none" onClick={handlePublish}>
          <PublishIcon />
        </Link>
      );
    }

    return null;
  };

  const rightSideEditMenu = () => {
    if (!auth.isAuthenticated) {
      return [];
    }

    if (!page.isEditable) {
      return [];
    }

    if (!page.inEdit) {
      return (
        <Menu
          anchorEl={rightAnchorEl}
          open={Boolean(rightAnchorEl)}
          onClose={handleMenuClose}
        >
          <MenuItem onClick={handleEdit}>
            <EditNoteIcon />
          </MenuItem>
          {data.isDirty && (
            <MenuItem onClick={handlePublish}>
              <PublishIcon />
            </MenuItem>
          )}
        </Menu>
      );
    }

    return (
      <Menu
        anchorEl={rightAnchorEl}
        open={Boolean(rightAnchorEl)}
        onClose={handleMenuClose}
      >
        <MenuItem onClick={handleInsert}>
          <AddCircleOutlineIcon />
          Insert
        </MenuItem>
        {page.isDirty && (
          <MenuItem onClick={handleSave}>
            <SaveIcon />
          </MenuItem>
        )}
        <MenuItem onClick={handleCancel}>
          <CancelIcon />
        </MenuItem>
        {data.isDirty && (
          <MenuItem onClick={handlePublish}>
            <PublishIcon />
          </MenuItem>
        )}
      </Menu>
    );
  };

  return (
    <MuiAppBar className="AppBar" sx={{ display: "grid" }}>
      <Toolbar sx={{ justifyContent: "space-between" }} ref={toolBarRef}>
        {/* Left side */}
        <Box sx={{ display: "flex", alignItems: "center" }}>
          {isMobile ? (
            <IconButton
              edge="start"
              color="inherit"
              aria-label="menu"
              onClick={handleLeftMenuOpen}
            >
              <MenuIcon />
            </IconButton>
          ) : (
            <Box sx={{ display: "flex", gap: 2 }}>
              <Link href="/" color="inherit" underline="none">
                <DashboardIcon />
                Dashboard
              </Link>
            </Box>
          )}
        </Box>

        {/* Right side */}
        <Box sx={{ display: "flex", alignItems: "center" }}>
          {isMobile ? (
            <IconButton
              edge="end"
              color="inherit"
              aria-label="menu"
              onClick={handleRightMenuOpen}
            >
              <MoreIcon />
            </IconButton>
          ) : (
            <Box sx={{ display: "flex", gap: 2 }}>
              {rightSideEditLinks()}
              {rightSidePublishLink()}
            </Box>
          )}
        </Box>
      </Toolbar>

      {/* Left-side menu for mobile */}
      <Menu
        anchorEl={leftAnchorEl}
        open={Boolean(leftAnchorEl)}
        onClose={handleMenuClose}
      >
        <MenuItem onClick={handleMenuClose} component={Link} href="/">
          <DashboardIcon />
          Dashboard
        </MenuItem>
      </Menu>

      {/* Right-side menu for mobile */}
      {rightSideEditMenu()}
    </MuiAppBar>
  );
};

export default AppBar;
