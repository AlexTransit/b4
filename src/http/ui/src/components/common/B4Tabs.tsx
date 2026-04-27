import { Tabs, TabsProps, Tab, TabProps, Stack, Box } from "@mui/material";
import { colors } from "@design";

export const B4Tabs = ({ sx, ...props }: TabsProps) => (
  <Tabs
    variant="scrollable"
    scrollButtons="auto"
    sx={{
      borderBottom: `1px solid ${colors.border.light}`,
      minHeight: 38,
      "& .MuiTabs-flexContainer": {
        gap: "4px",
      },
      "& .MuiTab-root": {
        color: colors.text.secondary,
        textTransform: "none",
        fontSize: 13,
        minHeight: 38,
        padding: "10px 12px",
        "&.Mui-selected": {
          color: colors.secondary,
        },
      },
      "& .MuiTabs-indicator": {
        bgcolor: colors.secondary,
        height: 2,
      },
      ...sx,
    }}
    {...props}
  />
);

interface B4TabProps extends Omit<TabProps, "label" | "icon"> {
  icon?: React.ReactElement;
  label: string;
  inline?: boolean;
  hasChanges?: boolean;
}

export const B4Tab = ({
  icon,
  label,
  inline,
  hasChanges,
  ...props
}: B4TabProps) => (
  <Tab
    icon={icon}
    iconPosition={inline ? "start" : undefined}
    label={
      hasChanges ? (
        <Stack direction="row" spacing={1} alignItems="center">
          <span>{label}</span>
          <Box
            sx={{
              width: 6,
              height: 6,
              borderRadius: "50%",
              bgcolor: colors.secondary,
            }}
          />
        </Stack>
      ) : (
        label
      )
    }
    {...props}
  />
);
