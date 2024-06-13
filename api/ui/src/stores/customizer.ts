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

import { defineStore } from 'pinia';
import config from '@/config';

export const useCustomizerStore = defineStore({
  id: 'customizer',
  state: () => ({
    sidebar_drawer: config.sidebar_drawer,
    mini_sidebar: config.mini_sidebar,
    actTheme: config.actTheme,
    fontTheme: config.fontTheme
  }),

  getters: {},
  actions: {
    set_theme(payload: string) {
      this.actTheme = payload;
    },
    set_font(payload: string) {
      this.fontTheme = payload;
    },
    set_sidebar_draw() {
      this.sidebar_drawer = !this.sidebar_drawer;
    },
    set_sidebar_mini(payload: boolean) {
      this.mini_sidebar = payload;
    }
  }
});
