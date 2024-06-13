// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

export type ThemeTypes = {
  name: string;
  variables?: object;
  colors: {
    primary?: string;
    secondary?: string;
    info?: string;
    success?: string;
    accent?: string;
    warning?: string;
    error?: string;
    lightprimary?: string;
    lightsecondary?: string;
    lightsuccess?: string;
    lighterror?: string;
    lightwarning?: string;
    darkprimary?: string;
    darksecondary?: string;
    darkText?: string;
    lightText?: string;
    borderLight?: string;
    inputBorder?: string;
    containerBg?: string;
    surface?: string;
    background?: string;
    'on-surface-variant'?: string;
    gray100?: string;
    primary200?: string;
    secondary200?: string;
  };
};
