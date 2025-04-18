/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
import * as React from "react";
import useScrollTrigger from "@mui/material/useScrollTrigger";
import Container from "@mui/material/Container";
import Fab from "@mui/material/Fab";
import KeyboardArrowUpIcon from "@mui/icons-material/KeyboardArrowUp";
import Fade from "@mui/material/Fade";
import Box from "@mui/material/Box";

import "../sass/backtotop.scss";

interface ScrollTopProps {
  /**
   * Injected by the documentation to work in an iframe.
   * You won't need it on your project.
   */
  window?: () => Window;
  children?: React.ReactElement[] | React.ReactElement;
  fabRight?: number;
  ref?: React.Ref<HTMLDivElement>;
}

interface BackToTopProps {
  /**
   * Injected by the documentation to work in an iframe.
   * You won't need it on your project.
   */
  window?: () => Window;
  children?: React.ReactElement[] | React.ReactElement;
}

function ScrollTop(props: ScrollTopProps) {
  const { children, window, fabRight } = props;
  // Note that you normally won't need to set the window ref as useScrollTrigger
  // will default to window.
  // This is only being set here because the demo is in an iframe.
  const trigger = useScrollTrigger({
    target: window ? window() : undefined,
    disableHysteresis: true,
    threshold: 100,
  });

  const handleClick = (event: React.MouseEvent<HTMLDivElement>) => {
    const anchor = (
      (event.target as HTMLDivElement).ownerDocument || document
    ).querySelector("#back-to-top-anchor");

    if (anchor) {
      anchor.scrollIntoView({
        block: "center",
      });
    }
  };

  const ref = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    if (ref.current) {
      ref.current.style.right = `${fabRight}px`;
    }
  }, [fabRight]);

  return (
    <Fade in={trigger}>
      <Box
        ref={ref}
        onClick={handleClick}
        role="presentation"
        sx={{ position: "fixed", bottom: 16, right: 16 }}
      >
        {children}
      </Box>
    </Fade>
  );
}

export default function BackToTop(props: BackToTopProps) {
  const { children } = props;

  const containerRef = React.useRef<HTMLDivElement>(null);

  const [fabRight, setFabRight] = React.useState(0);

  React.useEffect(() => {
    const handleResize = () => {
      if (containerRef.current) {
        let tmp =
          (window.innerWidth + containerRef.current.offsetWidth) / 2 -
          containerRef.current.offsetWidth +
          40;

        if (tmp < 40) {
          tmp = 40;
        }

        setFabRight(tmp);
      }
    };

    handleResize();

    window.addEventListener("resize", handleResize);

    return () => {
      window.removeEventListener("resize", handleResize);
    };
  }, []);

  return (
    <React.Fragment>
      <div id="back-to-top-anchor" />
      <Container ref={containerRef}>{children}</Container>
      <ScrollTop fabRight={fabRight}>
        <Fab size="small" aria-label="scroll back to top">
          <KeyboardArrowUpIcon />
        </Fab>
      </ScrollTop>
    </React.Fragment>
  );
}
