import { Box, Typography } from "@mui/material";
import { colors, fonts, radiusPx } from "@design";

type Tone = "primary" | "secondary" | "muted";

interface StatCardProps {
  label: string;
  value: string | number;
  unit?: string;
  sub?: string;
  tone?: Tone;
}

const dotColor: Record<Tone, string> = {
  primary: colors.primary,
  secondary: colors.secondary,
  muted: colors.text.disabled,
};

const splitValue = (
  raw: string | number,
): { value: string; unit?: string } => {
  if (typeof raw === "number") return { value: String(raw) };
  const match = /^(-?[\d.,]+)([A-Za-z%]+)$/.exec(raw);
  return match ? { value: match[1], unit: match[2] } : { value: raw };
};

export const StatCard = ({
  label,
  value,
  unit,
  sub,
  tone = "primary",
}: StatCardProps) => {
  const split = splitValue(value);
  const renderedUnit = unit ?? split.unit;
  const renderedValue = split.value;
  return (
    <Box
      sx={{
        bgcolor: colors.background.paper,
        border: `1px solid ${colors.border.default}`,
        borderRadius: `${radiusPx.md}px`,
        p: "16px 18px",
        display: "flex",
        flexDirection: "column",
        justifyContent: "space-between",
        gap: "10px",
        minWidth: 0,
        minHeight: 0,
        width: "100%",
        height: "100%",
      }}
    >
      <Typography
        component="div"
        sx={{
          fontSize: 10,
          letterSpacing: "0.18em",
          textTransform: "uppercase",
          color: colors.text.secondary,
          opacity: 0.8,
          lineHeight: 1,
        }}
      >
        {label}
      </Typography>
      <Typography
        component="div"
        sx={{
          fontSize: 30,
          fontWeight: 700,
          color: colors.text.primary,
          lineHeight: 1,
          letterSpacing: "-0.015em",
          fontFeatureSettings: '"tnum"',
        }}
      >
        {renderedValue}
        {renderedUnit && (
          <Box
            component="span"
            sx={{
              fontSize: 16,
              fontWeight: 600,
              color: colors.text.secondary,
              ml: "2px",
              letterSpacing: 0,
            }}
          >
            {renderedUnit}
          </Box>
        )}
      </Typography>
      <Box
        sx={{
          fontFamily: fonts.mono,
          fontSize: 11,
          color: colors.text.secondary,
          display: "flex",
          alignItems: "center",
          gap: "6px",
          minHeight: 14,
          lineHeight: 1,
        }}
      >
        <Box
          component="span"
          sx={{
            width: 5,
            height: 5,
            borderRadius: "50%",
            bgcolor: dotColor[tone],
            flexShrink: 0,
          }}
        />
        {sub ?? " "}
      </Box>
    </Box>
  );
};
