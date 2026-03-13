import { Box, ButtonBase, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
import { colors } from "@design";

const languages = [
  { code: "en", label: "EN" },
  { code: "ru", label: "RU" },
];

export const LanguageSwitcher = () => {
  const { i18n } = useTranslation();

  return (
    <Box
      sx={{
        display: "flex",
        justifyContent: "center",
        gap: 0.5,
        py: 1,
        px: 2,
      }}
    >
      {languages.map((lang) => (
        <ButtonBase
          key={lang.code}
          onClick={() => {
            i18n.changeLanguage(lang.code);
          }}
          sx={{
            px: 1.5,
            py: 0.5,
            borderRadius: 1,
            bgcolor:
              i18n.language === lang.code
                ? colors.accent.primary
                : "transparent",
            "&:hover": {
              bgcolor:
                i18n.language === lang.code
                  ? colors.accent.primaryHover
                  : colors.background.dark,
            },
          }}
        >
          <Typography
            variant="caption"
            sx={{
              fontWeight: i18n.language === lang.code ? 700 : 400,
              color: colors.text.primary,
            }}
          >
            {lang.label}
          </Typography>
        </ButtonBase>
      ))}
    </Box>
  );
};
