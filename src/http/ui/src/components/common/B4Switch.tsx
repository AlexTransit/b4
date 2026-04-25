import {
  FormControlLabel,
  Switch,
  SwitchProps,
  Typography,
  Box,
} from "@mui/material";
import { colors } from "@design";

interface B4SwitchProps extends Omit<SwitchProps, "checked" | "onChange"> {
  label: string;
  checked: boolean;
  description?: string;
  disabled?: boolean;
  onChange: (checked: boolean) => void;
}

export const B4Switch = ({
  label,
  checked,
  description,
  onChange,
  disabled,
  ...props
}: B4SwitchProps) => (
  <FormControlLabel
    disabled={disabled}
    control={
      <Switch
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
        {...props}
      />
    }
    label={
      <Box>
        <Typography sx={{ color: colors.text.primary, fontWeight: 500 }}>
          {label}
        </Typography>
        {description && (
          <Typography
            variant="caption"
            sx={{
              display: "block",
              color: colors.text.secondary,
              mt: "2px",
            }}
          >
            {description}
          </Typography>
        )}
      </Box>
    }
    sx={{
      alignItems: "flex-start",
      ml: 0,
      mr: 0,
      gap: "12px",
      "& .MuiFormControlLabel-label": {
        marginTop: "1px",
      },
    }}
  />
);

export default B4Switch;
