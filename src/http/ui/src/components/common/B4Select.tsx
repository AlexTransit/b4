import {
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  SelectProps,
  FormHelperText,
} from "@mui/material";
import { colors } from "@design";

interface B4SelectProps extends Omit<SelectProps<string | number>, "variant"> {
  label: string;
  options: { value: string | number; label: string }[];
  helperText?: string;
}

export const B4Select = ({
  label,
  options,
  helperText,
  ...props
}: B4SelectProps) => {
  return (
    <FormControl fullWidth size="small">
      <InputLabel shrink sx={{ color: colors.text.secondary }}>{label}</InputLabel>
      <Select
        {...props}
        label={label}
        displayEmpty
        renderValue={(selected) => {
          const match = options.find((o) => o.value === selected);
          return match?.label ?? String(selected);
        }}
        sx={{
          bgcolor: colors.background.dark,
          "& .MuiOutlinedInput-notchedOutline": {
            borderColor: colors.border.default,
          },
          "&:hover .MuiOutlinedInput-notchedOutline": {
            borderColor: colors.border.medium,
          },
          "&.Mui-focused .MuiOutlinedInput-notchedOutline": {
            borderColor: colors.secondary,
          },

          ...props.sx,
        }}
      >
        {options.map((option) => (
          <MenuItem key={option.value} value={option.value}>
            {option.label}
          </MenuItem>
        ))}
      </Select>
      {helperText && (
        <FormHelperText sx={{ color: colors.text.secondary, ml: 0.1 }}>
          {helperText}
        </FormHelperText>
      )}
    </FormControl>
  );
};

export default B4Select;
