import { IconButton, Menu, MenuItem, ListItemIcon, ListItemText } from '@mui/material';
import LanguageIcon from '@mui/icons-material/Language';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

const languages = [
  { code: 'es', label: 'Español', flag: '🇦🇷' },
  { code: 'en', label: 'English', flag: '🇺🇸' },
  { code: 'pt', label: 'Português', flag: '🇧🇷' },
];

export default function LanguageSwitcher() {
  const { i18n } = useTranslation();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);

  const handleOpen = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleSelect = (code: string) => {
    i18n.changeLanguage(code);
    handleClose();
  };

  return (
    <>
      <IconButton color="inherit" onClick={handleOpen} size="small">
        <LanguageIcon />
      </IconButton>
      <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={handleClose}>
        {languages.map((lang) => (
          <MenuItem
            key={lang.code}
            selected={i18n.language === lang.code}
            onClick={() => handleSelect(lang.code)}
          >
            <ListItemIcon sx={{ minWidth: 32 }}>{lang.flag}</ListItemIcon>
            <ListItemText>{lang.label}</ListItemText>
          </MenuItem>
        ))}
      </Menu>
    </>
  );
}
