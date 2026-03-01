import { Container, Stack } from "@mui/material";
import { DetectorRunner } from "./Detector";

export function DetectorPage() {
  return (
    <Container
      maxWidth={false}
      sx={{
        height: "100%",
        display: "flex",
        flexDirection: "column",
        overflow: "auto",
        py: 3,
      }}
    >
      <Stack spacing={3}>
        <DetectorRunner />
      </Stack>
    </Container>
  );
}
