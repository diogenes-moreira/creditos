import React, { useMemo } from "react";
import { Autocomplete, TextField, Grid } from "@mui/material";
import { Country, State, City } from "country-state-city";
import type { ICountry, IState } from "country-state-city";
import { useTranslation } from "react-i18next";

const LATAM_CODES = [
  "AR", "BO", "BR", "CL", "CO", "CR", "CU", "DO", "EC", "SV",
  "GT", "HN", "MX", "NI", "PA", "PY", "PE", "PR", "UY", "VE",
];

interface LocationSelectorProps {
  country: string;
  province: string;
  city: string;
  onChange: (field: "country" | "province" | "city", value: string) => void;
  errors?: { country?: boolean; province?: boolean; city?: boolean };
}

const LocationSelector: React.FC<LocationSelectorProps> = ({
  country,
  province,
  city,
  onChange,
  errors,
}) => {
  const { t } = useTranslation();

  const countries = useMemo(
    () => Country.getAllCountries().filter((c) => LATAM_CODES.includes(c.isoCode)),
    []
  );

  const selectedCountry = useMemo(
    () => countries.find((c) => c.name === country) || null,
    [countries, country]
  );

  const states = useMemo(
    () => (selectedCountry ? State.getStatesOfCountry(selectedCountry.isoCode) : []),
    [selectedCountry]
  );

  const selectedState = useMemo(
    () => states.find((s) => s.name === province) || null,
    [states, province]
  );

  const cityNames = useMemo(
    () =>
      selectedCountry && selectedState
        ? City.getCitiesOfState(selectedCountry.isoCode, selectedState.isoCode).map((c) => c.name)
        : [],
    [selectedCountry, selectedState]
  );

  return (
    <>
      <Grid item xs={12} sm={4}>
        <Autocomplete
          options={countries}
          getOptionLabel={(o: ICountry) => o.name}
          value={selectedCountry}
          onChange={(_, val) => {
            onChange("country", val?.name || "");
            onChange("province", "");
            onChange("city", "");
          }}
          renderInput={(params) => (
            <TextField
              {...params}
              label={t("registration.country")}
              error={errors?.country}
              fullWidth
            />
          )}
          isOptionEqualToValue={(opt, val) => opt.isoCode === val.isoCode}
        />
      </Grid>
      <Grid item xs={12} sm={4}>
        <Autocomplete
          options={states}
          getOptionLabel={(o: IState) => o.name}
          value={selectedState}
          onChange={(_, val) => {
            onChange("province", val?.name || "");
            onChange("city", "");
          }}
          renderInput={(params) => (
            <TextField
              {...params}
              label={t("registration.province")}
              error={errors?.province}
              fullWidth
            />
          )}
          isOptionEqualToValue={(opt, val) => opt.isoCode === val.isoCode}
          disabled={!selectedCountry}
        />
      </Grid>
      <Grid item xs={12} sm={4}>
        <Autocomplete
          freeSolo
          options={cityNames}
          value={city}
          onChange={(_, val) => {
            onChange("city", (val as string) || "");
          }}
          onInputChange={(_, val, reason) => {
            if (reason === "input") {
              onChange("city", val);
            }
          }}
          renderInput={(params) => (
            <TextField
              {...params}
              label={t("registration.city")}
              error={errors?.city}
              fullWidth
            />
          )}
          disabled={!selectedState}
        />
      </Grid>
    </>
  );
};

export default LocationSelector;
